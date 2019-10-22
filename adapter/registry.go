package adapter

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/amclient"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sirupsen/logrus"
)

var reloadFrequency = 10 * time.Second

type registryRecord struct {
	TenantJiscID             string `dynamodbav:"tenantJiscID"`
	ArchivematicaURL         string `dynamodbav:"url"`
	ArchivematicaUser        string `dynamodbav:"user"`
	ArchivematicaKey         string `dynamodbav:"key"`
	ArchivematicaTransferDir string `dynamodbav:"transferDir"`
}

type Registry struct {
	ctx            context.Context
	cancel         context.CancelFunc
	logger         logrus.FieldLogger
	dynamodbClient dynamodbiface.DynamoDBAPI
	dynamodbTable  string
	reloadCh       chan struct{}
	stopCh         chan chan struct{}
	r              map[uint64]*amclient.Client
	sync.RWMutex
}

// NewRegistry returns a usable registry.
func NewRegistry(logger logrus.FieldLogger, dynamodbClient dynamodbiface.DynamoDBAPI, dynamodbTable string) (*Registry, error) {
	r := &Registry{
		logger:         logger,
		dynamodbClient: dynamodbClient,
		dynamodbTable:  dynamodbTable,
		reloadCh:       make(chan struct{}),
		stopCh:         make(chan chan struct{}),
		r:              make(map[uint64]*amclient.Client),
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	if err := r.load(); err != nil {
		return nil, errors.Wrap(err, "registry failed to load from source")
	}
	go r.loop()
	return r, nil
}

// load retrieves the registry records from DynamoDB into the local registry
// data structure with initialized clients.
func (r *Registry) load() error {
	res, err := r.dynamodbClient.ScanWithContext(r.ctx, &dynamodb.ScanInput{
		TableName:      aws.String(r.dynamodbTable),
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return errors.Wrap(err, "failed to scan registry")
	}
	if len(res.Items) < 1 {
		r.logger.WithField("table", r.dynamodbTable).Warn("Registry has been loaded but it is empty")
		return nil
	}
	recs := []registryRecord{}
	if err := dynamodbattribute.UnmarshalListOfMaps(res.Items, &recs); err != nil {
		return errors.Wrap(err, "failed to unmarshal registry records")
	}
	newMap := make(map[uint64]*amclient.Client)
	for _, rec := range recs {
		i, err := strconv.ParseInt(rec.TenantJiscID, 10, 64)
		if err != nil {
			return errors.Wrap(err, "failed to parse tenantJiscID")
		}
		c, err := amclient.New(
			http.DefaultClient,
			rec.ArchivematicaURL,
			rec.ArchivematicaUser,
			rec.ArchivematicaKey,
			amclient.SetFsPath(rec.ArchivematicaTransferDir))
		if err != nil {
			return errors.Wrapf(err, "failed to create client for tenantJiscID %s", rec.TenantJiscID)
		}
		newMap[uint64(i)] = c
	}
	r.Lock()
	r.r = newMap
	r.Unlock()
	return nil
}

func (r *Registry) loop() {
	ticker := time.NewTicker(reloadFrequency)
	for {
		select {
		case ch := <-r.stopCh:
			r.cancel()
			close(ch)
			return
		case <-r.ctx.Done():
			return
		case <-ticker.C:
		case <-r.reloadCh:
		}
		_ = r.load()
	}
}

// Get a client for a given tenant.
func (r *Registry) Get(tenantID uint64) *amclient.Client {
	r.RLock()
	defer r.RUnlock()
	return r.r[tenantID]
}

func (r *Registry) Log() {
	r.RLock()
	defer r.RUnlock()
	for tenantID, client := range r.r {
		r.logger.WithFields(logrus.Fields{
			"tenantJiscID": tenantID,
			"url":          client.BaseURL.String(),
		}).Warn("Registry entry found")
	}
}

// Reload is a non-blocking request to reload the registry. The operation is
// omitted if it is already happening.
func (r *Registry) Reload() {
	select {
	case r.reloadCh <- struct{}{}:
		r.logger.Warn("Reloading registry")
		return
	default:
		r.logger.Warn("The registry is currently reloading the entries")
	}
}

func (r *Registry) Stop() {
	ch := make(chan struct{})
	r.stopCh <- ch
	<-ch
}
