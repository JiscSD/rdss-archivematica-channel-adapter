package consumer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
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
	broker     *broker.Broker
	ctx        context.Context
	logger     log.FieldLogger
	amc        *amclient.Client
	s3         s3.ObjectStorage
	amSharedFs afero.Fs
}

// MakeConsumer returns a new ConsumerImpl which implements Consumer
func MakeConsumer(
	ctx context.Context,
	logger log.FieldLogger,
	broker *broker.Broker,
	amc *amclient.Client,
	s3 s3.ObjectStorage,
	amSharedFs afero.Fs) *ConsumerImpl {
	return &ConsumerImpl{
		ctx:        ctx,
		logger:     logger,
		broker:     broker,
		amc:        amc,
		s3:         s3,
		amSharedFs: amSharedFs,
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

	t, err := c.amc.TransferSession(body.ObjectTitle, c.amSharedFs)
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
		func() {
			var f afero.File
			f, err = t.Create(file.FileName)
			if err != nil {
				c.logger.Errorf("Error creating %s: %v", file.FileName, err)
				return
			}
			defer f.Close()
			if err = downloadFile(c.logger, c.ctx, c.s3, http.DefaultClient, f, file.FileStoragePlatform.StoragePlatformType, file.FileStorageLocation, nil); err != nil {
				return
			}
			describeFile(t, file.FileName, &file)
		}()

		// Just a single error is enough for us to halt the transfer completely.
		if err == nil {
			continue
		}
		defer func() {
			if err := t.Destroy(); err != nil {
				c.logger.Warningf("Error destroying transfer: %v", err)
			}
		}()
		return err
	}

	id, err := t.Start()
	if err != nil {
		return err
	}
	c.logger.Debugf("The transfer has started successfully, id: %s", id)

	return nil
}

// retry is a retry-backoff time provider that manages times between retries for the http storage type.
// It can be nil in which case the default scheme will be used. The S3 download includes its own
// retry scheme (http://docs.aws.amazon.com/general/latest/gr/api-retries.html)
func downloadFile(logger log.FieldLogger, ctx context.Context, s3Client s3.ObjectStorage, httpClient *http.Client, target afero.File,
	storageType message.StorageTypeEnum, storageLocation string, retry backoff.BackOff) error {
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
		n, err = downloadFileHTTP(ctx, httpClient, target, storageLocation, retry)

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

func downloadFileHTTP(ctx context.Context, httpClient *http.Client, target io.Writer, storageLocation string, retry backoff.BackOff) (int64, error) {
	// Use exponential backoff algorithm if the user doesn't provide one.
	if retry == nil {
		retry = backoff.NewExponentialBackOff()
	}
	// Create a BackOffContext to stop retrying after the context is canceled.
	cb := backoff.WithContext(retry, ctx)

	// Create the Request.
	req, err := http.NewRequest("GET", storageLocation, nil)
	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)

	// This is the operation that we want to retry.
	var resp *http.Response
	op := func() error {
		var err error
		resp, err = httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("unexpected status code: %d (%s)", resp.StatusCode, resp.Status)
		}
		return nil
	}

	if err := backoff.Retry(op, cb); err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return io.Copy(target, resp.Body)
}

// describeDataset maps properties from a research object into a CSV entry
// in the `metadata.csv` file used in `amclient`.
// No need to assign the identifierType now as the XSD has a fixed value of "DOI"
// If this gets more types in future it can be added in the ObjectIdentifier loop with
//  t.Describe("identifierType", item.IdentifierType)
func describeDataset(t *amclient.TransferSession, f *message.MetadataCreateRequest) {
	t.Describe("dc.title", f.ObjectTitle)
	t.Describe("dc.type", f.ObjectResourceType.String())

	for _, item := range f.ObjectIdentifier {
		t.Describe("dc.identifier", item.IdentifierValue)
		// No need to assign this now as the XSD has a fixed value of "DOI"
		// t.Describe("identifierType", item.IdentifierType)
	}

	for _, item := range f.ObjectDate {
		if item.DateType != message.DateTypeEnum_published {
			continue
		}
		t.Describe("dcterms.issued", item.DateValue)
		t.Describe("dc.publicationYear", item.DateValue)
	}

	for _, item := range f.ObjectOrganisationRole {
		t.Describe("dc.publisher", item.Organisation.OrganisationName)
	}

	for _, item := range f.ObjectPersonRole {
		if item.Role == message.PersonRoleEnum_dataCreator {
			t.Describe("dc.creatorName", item.Person.PersonGivenName)
		}
		if item.Role == message.PersonRoleEnum_publisher {
			t.Describe("dc.publisher", item.Person.PersonGivenName)
		}
	}
}

// describeFile maps properties from an intellectual asset into a CSV entry
// in the `metadata.csv` file used in `amclient`.
func describeFile(t *amclient.TransferSession, name string, f *message.File) {
	n := fmt.Sprintf("objects/%s", name)
	t.DescribeFile(n, "dc.identifier", f.FileIdentifier)
	t.DescribeFile(n, "dc.title", f.FileName)
}
