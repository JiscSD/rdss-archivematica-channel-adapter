package integration

import (
	"flag"
	"fmt"
	"testing"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/integration/adapter"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/integration/ammock"
)

const (
	awsAccountID       = "123456789012"
	awsAccessKeyID     = "123"
	awsSecretAccessKey = "xyz"
	awsTokenKey        = ""
	awsRegion          = "us-east-1"

	awsS3Endpoint       = "http://localhost:4572"
	awsDynamoDBEndpoint = "http://localhost:4569"
	awsSQSEndpoint      = "http://localhost:4576"
	awsSNSEndpoint      = "http://localhost:4575"

	awsBucket          = "bucket"
	awsRepositoryTable = "rdss_archivematica_adapter_local_data_repository"
	awsProcessingTable = "rdss_archivematica_adapter_processing_state"
	awsRegistryTable   = "rdss_archivematica_adapter_registry"
	awsQueueMain       = "http://localhost:4576/queue/main"
	awsTopicMain       = "arn:aws:sns:us-east-1:123456789012:main"
	awsTopicInvalid    = "arn:aws:sns:us-east-1:123456789012:invalid"
	awsTopicError      = "arn:aws:sns:us-east-1:123456789012:error"
)

var (
	awsS3Client       = s3Client()
	awsDynamoDBClient = dynamodbClient()
	awsSQSClient      = sqsClient()
	awsSNSClient      = snsClient()

	serverEnvironment = []string{
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REPOSITORY_TABLE=%s", awsRepositoryTable),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.PROCESSING_TABLE=%s", awsProcessingTable),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REGISTRY_TABLE=%s", awsRegistryTable),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_RECV_MAIN_ADDR=%s", awsQueueMain),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_MAIN_ADDR=%s", awsTopicMain),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_INVALID_ADDR=%s", awsTopicInvalid),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_ERROR_ADDR=%s", awsTopicError),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_ENDPOINT=%s", awsDynamoDBEndpoint),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_ENDPOINT=%s", awsS3Endpoint),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_ENDPOINT=%s", awsSQSEndpoint),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_ENDPOINT=%s", awsSNSEndpoint),
		fmt.Sprintf("AWS_S3_FORCE_PATH_STYLE=true"),
		fmt.Sprintf("AWS_REGION=%s", awsRegion),
		fmt.Sprintf("AWS_ACCESS_KEY=%s", awsAccessKeyID),
		fmt.Sprintf("AWS_SECRET_KEY=%s", awsSecretAccessKey),
	}
)

var (
	flagDebug = flag.Bool("debug", false, "")
)

// TestInvalidMessage confirms that the adapter reacts to invalid messages by
// sending them to the corresponding invalid message queue.
func TestInvalidMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	defer cleanUp(t)

	s := subscriber(t)
	defer s.cleanUp()

	sendMessage(t, "invalid message")

	cmd := adapter.Server().WithEnv(serverEnvironment)
	stop := cmd.RunBackground(t)
	defer stop()

	s.AssertInvalidMessageReceived("invalid message")
	s.AssertInvalidQueueIsEmpty()
}

// TestLocalDataRepository confirms that the same message delivered twice
// results in the adapter refusing the second attempt.
func TestLocalDataRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	defer cleanUp(t)

	s := subscriber(t)
	defer s.cleanUp()

	msg := newMetadataCreateMessage(
		t, 1, "Research dataset 1", message.StorageTypeEnum_S3,
		"foobar", "foobar", "foobar", 12345)
	sendMessage(t, msg)
	sendMessage(t, msg)

	cmd := adapter.Server().WithEnv(serverEnvironment)
	stop := cmd.RunBackground(t)
	defer stop()

	// We're looking at error messages because that's the expected result in
	// the adapter when a registry is not populated.
	s.AssertErrorMessageReceived("")
	s.AssertErrorQueueIsEmpty()
}

// TestPipelineRouting confirms that the adapter is routing transfers requests
// to the different pipelines registered in the registry according to the
// tenantJiscID.TestPipelineRouting
//
// In this test, we're going to set up two pipelines:
// * pipe1, associated to tenantJiscID 1.
// * pipe2, associated to tenantJiscID 2.
//
// When a message is received, the adapter must look up the tenantJiscID and
// route the request (given that it's a MetadataCreate) to the corresponding
// pipeline.
func TestPipelineRouting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	defer cleanUp(t)

	s := subscriber(t)
	defer s.cleanUp()

	// Create first pipeline and store it in the registry.
	pipe1 := ammock.New(t)
	defer pipe1.Stop()
	registerPipeline(t, 1, pipe1)

	// Create second pipeline and store it in the registry.
	pipe2 := ammock.New(t)
	defer pipe2.Stop()
	registerPipeline(t, 2, pipe2)

	// Put a file in the S3 bucket.
	const filename = "dataset.zip"
	md5sum, size := putKnownObject(t, filename)
	location := fmt.Sprintf("s3://%s/%s", awsBucket, filename)

	cmd := adapter.Server().WithEnv(serverEnvironment)
	stop := cmd.RunBackground(t)
	defer stop()

	// Send MetadataCreate using tenantJiscID=1.
	sendMessage(t, newMetadataCreateMessage(t,
		1, "Research dataset 1", message.StorageTypeEnum_S3,
		filename, location, md5sum, size))

	// We're expecting a presevation event from Archivematica.
	// TODO: verify the payload.
	s.AssertMainMessageReceived("")

	pipe1.AssertAPIUsed()
	pipe1.AssertTransferDirIsNotEmpty()

	pipe2.AssertAPINotUsed()
	pipe2.AssertTransferDirIsEmpty()
}

// cleanUp attempts to leave the resources in its initial state.
func cleanUp(t *testing.T) {
	purgeQueue(t, awsQueueMain)
	purgeDynamoDBTable(t, awsRepositoryTable, "ID")
	purgeDynamoDBTable(t, awsProcessingTable, "objectUUID")
	purgeDynamoDBTable(t, awsRegistryTable, "tenantJiscID")
}
