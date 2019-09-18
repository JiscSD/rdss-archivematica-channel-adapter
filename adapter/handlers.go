package adapter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"

	"github.com/cenkalti/backoff/v3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var (
	UnknownTenantErr = errors.New("unknown tenantJiscID")
)

// handleMetadataCreateRequest handles the reception of Metadata Create
// messages.
func (c *Adapter) handleMetadataCreateRequest(msg *message.Message) error {
	body, err := msg.MetadataCreateRequest()
	if err != nil {
		return err
	}
	amClient := c.registry.Get(msg.MessageHeader.TenantJiscID)
	if amClient == nil {
		return errors.Wrap(UnknownTenantErr, strconv.Itoa(int(msg.MessageHeader.TenantJiscID)))
	}
	researchObject := body.InferResearchObject()
	id, err := c.startTransfer(amClient, researchObject)
	if err != nil {
		return errors.Wrap(err, "transfer cannot be started")
	}
	c.logger.Debugf("The transfer has started successfully, id: %s", id)
	if err := c.storage.AssociateResearchObject(c.ctx, researchObject.ObjectUUID.String(), id); err != nil {
		// We don't want to discard the message at this point.
		c.logger.Errorf("Error trying to persist the research object: %v", err)
	}
	aipid, err := amclient.WaitUntilStored(c.ctx, amClient, id)
	if err != nil {
		return err
	}
	aipuuid, err := message.ParseUUID(aipid)
	if err != nil {
		return errors.Wrap(err, "SIP UUID is invalid")
	}
	var (
		packageTypeAIP        = message.PackageTypeEnum_AIP
		packageContainerType  = message.ContainerTypeEnum_zip
		preservationEventType = message.PreservationEventTypeEnum_informationPackageCreation
	)
	err = c.broker.Preservation.Event(c.ctx, &message.PreservationEventRequest{
		InformationPackage: message.InformationPackage{
			ObjectUUID:           researchObject.ObjectUUID,
			PackageUUID:          aipuuid,
			PackageType:          &packageTypeAIP,
			PackageContainerType: &packageContainerType,
			PackagePreservationEvent: message.PreservationEvent{
				PreservationEventValue: uuid.New().String(),
				PreservationEventType:  &preservationEventType,
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "PreservationEvent message could not be sent")
	}
	return nil
}

// handleMetadataUpdateRequest handles the reception of Metadata Update
// messages. It may result in a package being reingested if it's been already
// preserved before.
func (c *Adapter) handleMetadataUpdateRequest(msg *message.Message) error {
	logger := c.logger.WithFields(logrus.Fields{"handler": "MetadataUpdate", "message": msg.ID()})
	body, err := msg.MetadataUpdateRequest()
	if err != nil {
		return err
	}
	amClient := c.registry.Get(msg.MessageHeader.TenantJiscID)
	if amClient == nil {
		return errors.Wrap(UnknownTenantErr, string(msg.MessageHeader.TenantJiscID))
	}
	researchObject := body.InferResearchObject()
	// Determine if the message is pointing to a previous dataset.
	var match *message.IdentifierRelationship
	for _, item := range researchObject.ObjectRelatedIdentifier {
		if item.RelationType != message.RelationTypeEnum_isNewVersionOf {
			continue
		}
		match = &item
		break // If there's more than one match we're not going to care.
	}
	if match == nil || match.Identifier.IdentifierValue == "" {
		logger.Debug("Ignoring message.")
		return nil // Stop here, ignore message.
	}
	// Determine match.IdentifierValue's (ObjectUUID) is a known dataset.
	transferID, err := c.storage.GetResearchObject(c.ctx, match.Identifier.IdentifierValue)
	if err != nil {
		logger.WithFields(logrus.Fields{"err": err, "IdentifierValue": match.Identifier.IdentifierValue}).Warn("Cannot fetch or find associated object in the local store.")
		return nil
	}
	// At this point we know the previous transferID so we could reingest.
	// In this first iteration we're just starting a new transfer.
	logger.WithFields(logrus.Fields{"transferID": transferID, "TODO": "Implement real reingest."}).Debug("Reingesting transfer.")
	_, err = c.startTransfer(amClient, researchObject)
	if err != nil {
		return err
	}
	return nil
}

func (c *Adapter) startTransfer(amClient *amclient.Client, researchObject *message.ResearchObject) (string, error) {
	// Ignore messages with no files listed.
	if len(researchObject.ObjectFile) == 0 {
		return "", nil
	}
	t, err := amClient.TransferSession(researchObject.ObjectTitle)
	if err != nil {
		return "", errors.Wrap(err, "transfer session cannot be initialized")
	}
	t.WithProcessingConfig(archivematicaProcessingConfig)
	// Process dataset metadata.
	describeDataset(t, researchObject)
	for _, file := range researchObject.ObjectFile {
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
		return "", err
	}
	return t.Start()
}

// retry is a retry-backoff time provider that manages times between retries for the http storage type.
// It can be nil in which case the default scheme will be used. The S3 download includes its own
// retry scheme (http://docs.aws.amazon.com/general/latest/gr/api-retries.html)
func downloadFile(logger logrus.FieldLogger, ctx context.Context, s3Client s3.ObjectStorage, httpClient *http.Client, target afero.File,
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
func describeDataset(t *amclient.TransferSession, f *message.ResearchObject) {
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
		t.Describe("dcterms.issued", item.DateValue.String())
		t.Describe("dc.publicationYear", item.DateValue.String())
	}

	for _, item := range f.ObjectOrganisationRole {
		t.Describe("dc.publisher", item.Organisation.OrganisationName)
	}

	for _, item := range f.ObjectPersonRole {
		if item.Role == message.PersonRoleEnum_dataCreator {
			t.Describe("dc.creatorName", item.Person.PersonGivenNames)
		}
		if item.Role == message.PersonRoleEnum_publisher {
			t.Describe("dc.publisher", item.Person.PersonGivenNames)
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
