package consumer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"
)

// Consumer is the component that subscribes to the broker and interacts with
// Archivematica.
type Consumer interface {
	Start()
}

// ConsumerImpl is an implementation of Consumer.
type ConsumerImpl struct {
	broker    *broker.Broker
	ctx       context.Context
	logger    log.FieldLogger
	amc       *amclient.Client
	s3        s3.ObjectStorage
	depositFs afero.Fs
}

// MakeConsumer returns a new ConsumerImpl which implements Consumer
func MakeConsumer(
	ctx context.Context,
	logger log.FieldLogger,
	broker *broker.Broker,
	amc *amclient.Client,
	s3 s3.ObjectStorage,
	depositFs afero.Fs) *ConsumerImpl {
	return &ConsumerImpl{
		ctx:       ctx,
		logger:    logger,
		broker:    broker,
		amc:       amc,
		s3:        s3,
		depositFs: depositFs,
	}
}

// Start implements Consumer
func (c *ConsumerImpl) Start() {
	c.broker.SubscribeType(message.MessageTypeMetadataCreate, c.handleMetadataCreateRequest)

	<-c.ctx.Done()
	c.broker.Close()
	c.logger.Info("Consumer says good-bye!")
}

// handleMetadataCreateRequest handles the reception of a Metadata Create
// messages.
func (c *ConsumerImpl) handleMetadataCreateRequest(msg *message.Message) error {
	body, err := msg.MetadataCreateRequest()
	if err != nil {
		return err
	}

	// Ignore messages with no files listed.
	if len(body.ObjectFile) == 0 {
		return nil
	}

	t, err := c.amc.TransferSession(body.ObjectTitle, c.depositFs)
	if err != nil {
		return err
	}

	// Download automated workflow.
	err = t.ProcessingConfig("automated")
	if err != nil {
		c.logger.Warningf("Failed to download `automated` processing configuration: %s", err)
	}

	// Process dataset metadata.
	describeDataset(t, body)

	for _, file := range body.ObjectFile {
		// Add checksum metadata. We're not going to verify checksums at this
		// point because this is something meant to do by Archivematica.
		for _, c := range file.FileChecksum {
			switch c.ChecksumType {
			case message.ChecksumTypeEnum_md5:
				t.ChecksumMD5(file.FileName, c.ChecksumValue)
			case message.ChecksumTypeEnum_sha256:
				t.ChecksumSHA256(file.FileName, c.ChecksumValue)
			}
		}

		// Download and describe each file.
		// Using an anonymous function so I can use defer inside this loop.
		var err error
		func() {
			var f afero.File
			f, err = t.Create(file.FileName)
			if err != nil {
				c.logger.Errorf("Error creating %s: %v", file.FileName, err)
				return
			}
			defer f.Close()
			if err = downloadFile(c.logger, c.ctx, c.s3, http.DefaultClient, f, file.FileStorageType, file.FileStorageLocation); err != nil {
				return
			}
			describeFile(t, file.FileName, &file)
		}()
		if err != nil {
			return err
		}
	}

	return t.Start()
}

func downloadFile(logger log.FieldLogger, ctx context.Context, s3Client s3.ObjectStorage, httpClient *http.Client, target afero.File, storageType message.StorageTypeEnum, storageLocation string) error {
	logger.Debugf("Saving %s into %s", storageLocation, target.Name())
	var (
		n      int64
		err    = fmt.Errorf("unsupported storage location type: %s", storageType)
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(ctx, time.Minute*30)
	defer cancel()
	switch storageType {
	case message.StorageTypeEnum_HTTP:
		n, err = downloadFileHTTP(ctx, httpClient, target, storageLocation)

	case message.StorageTypeEnum_S3:
		n, err = s3Client.Download(ctx, target, storageLocation)
	}
	if err != nil {
		logger.Errorf("Error downloading %s: %s", storageLocation, err)
		return err
	}
	logger.Debugf("Downloaded %s - %d bytes written", storageLocation, n)
	return nil
}

func downloadFileHTTP(ctx context.Context, httpClient *http.Client, target io.Writer, storageLocation string) (int64, error) {
	req, err := http.NewRequest("GET", storageLocation, nil)
	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d (%s)", resp.StatusCode, resp.Status)
	}
	return io.Copy(target, resp.Body)
}

// describeDataset maps properties from a research object into a CSV entry
// in the `metadata.csv` file used in `amclient`.
func describeDataset(t *amclient.TransferSession, f *message.MetadataCreateRequest) {
	t.Describe("dc.title", f.ObjectTitle)
	t.Describe("dc.type", f.ObjectResourceType.String())

	for _, item := range f.ObjectIdentifier {
		t.Describe("dc.identifier", item.IdentifierValue)
	}

	for _, item := range f.ObjectDate {
		if item.DateType != message.DateTypeEnum_published {
			continue
		}
		t.Describe("dcterms.issued", item.DateValue)
	}

	for _, item := range f.ObjectOrganisationRole {
		t.Describe("dc.publisher", item.Organisation.OrganisationName)
	}

	for _, item := range f.ObjectPersonRole {
		if item.Role != message.PersonRoleEnum_dataCreator {
			continue
		}
		t.Describe("dc.contributor", item.Person.PersonGivenName)
	}
}

// describeFile maps properties from an intellectual asset into a CSV entry
// in the `metadata.csv` file used in `amclient`.
func describeFile(t *amclient.TransferSession, name string, f *message.File) {
	n := fmt.Sprintf("objects/%s", name)
	t.DescribeFile(n, "dc.identifier", f.FileIdentifier)
	t.DescribeFile(n, "dc.title", f.FileName)
}
