package integration

// TODO:
// - topic registration can be duplicated, causing redirections to be sent more than once?
//   provision.sh is not removing it? dc reforce recreate will?
// - adapter shutdown should be graceful otherwise sqs shows some issues

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/integration/adapter"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/integration/ammock"
)

const (
	accountID       = "123456789012"
	accessKeyID     = "123"
	secretAccessKey = "xyz"
	tokenKey        = ""
	region          = "us-east-1"

	s3Endpoint       = "http://localhost:4572"
	dynamodbEndpoint = "http://localhost:4569"
	sqsEndpoint      = "http://localhost:4576"
	snsEndpoint      = "http://localhost:4575"

	testBucket          = "bucket"
	testRepositoryTable = "rdss_archivematica_adapter_local_data_repository"
	testProcessingTable = "rdss_archivematica_adapter_processing_state"
	testRegistryTable   = "rdss_archivematica_adapter_registry"
	testQueueMain       = "http://localhost:4576/queue/main"
	testTopicMain       = "arn:aws:sns:us-east-1:123456789012:main"
	testTopicInvalid    = "arn:aws:sns:us-east-1:123456789012:invalid"
	testTopicError      = "arn:aws:sns:us-east-1:123456789012:error"
)

var (
	flagDebug = flag.Bool("debug", false, "")

	awsS3Client       = s3Client()
	awsDynamoDBClient = dynamodbClient()
	awsSQSClient      = sqsClient()
	awsSNSClient      = snsClient()

	serverEnvironment = []string{
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REPOSITORY_TABLE=%s", testRepositoryTable),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.PROCESSING_TABLE=%s", testProcessingTable),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REGISTRY_TABLE=%s", testRegistryTable),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_RECV_MAIN_ADDR=%s", testQueueMain),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_MAIN_ADDR=%s", testTopicMain),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_INVALID_ADDR=%s", testTopicInvalid),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_ERROR_ADDR=%s", testTopicError),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_ENDPOINT=%s", dynamodbEndpoint),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_ENDPOINT=%s", s3Endpoint),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_ENDPOINT=%s", sqsEndpoint),
		fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_ENDPOINT=%s", snsEndpoint),
		fmt.Sprintf("AWS_S3_FORCE_PATH_STYLE=true"),
		fmt.Sprintf("AWS_REGION=%s", region),
		fmt.Sprintf("AWS_ACCESS_KEY=%s", accessKeyID),
		fmt.Sprintf("AWS_SECRET_KEY=%s", secretAccessKey),
	}
)

// TestInvalidMessage confirms that the adapter reacts to invalid messages by
// sending them to the corresponding invalid message queue.
func TestInvalidMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	defer cleanup(t)

	s := subscriber(t)

	sendMessage(t, "invalid message")

	cmd := adapter.Server().WithEnv(serverEnvironment)
	stop := cmd.RunBackground(t)
	defer func() {
		stop()
		time.Sleep(time.Second * 1)
	}()

	s.AssertInvalidMessageReceived("invalid message")
	s.AssertNoMoreIncomingMessages()
}

// TestLocalDataRepository confirms that the same message delivered twice
// results in the adapter refusing the second attempt.
func TestLocalDataRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	defer cleanup(t)

	s := subscriber(t)

	msg := newMetadataCreateMessage(
		t, 1, "Research dataset 1", message.StorageTypeEnum_S3,
		"foobar", "foobar", "foobar", 12345)
	sendMessage(t, msg)
	sendMessage(t, msg)

	cmd := adapter.Server().WithEnv(serverEnvironment)
	stop := cmd.RunBackground(t)
	defer func() {
		stop()
		time.Sleep(time.Second * 1)
	}()

	// We're looking at error messages because that's the expected result in
	// the adapter when a registry is not populated.
	s.AssertErrorMessageReceived("")

	// TODO: confirm that the echo error queue is empty, using get attributes
	// probably better than draining?
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
	defer cleanup(t)

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
	location := fmt.Sprintf("s3://%s/%s", testBucket, filename)

	cmd := adapter.Server().WithEnv(serverEnvironment)
	stop := cmd.RunBackground(t)
	defer func() {
		stop()
		time.Sleep(time.Second * 1)
	}()

	// Send MetadataCreate using tenantJiscID=1.
	sendMessage(t, newMetadataCreateMessage(t,
		1, "Research dataset 1", message.StorageTypeEnum_S3,
		filename, location, md5sum, size))

	time.Sleep(1 * time.Second)

	// TODO: pipe api and assertions are not implemented yet
	// TODO: confirm that preservation event is published by the adapter
	pipe1.AssertAPIUsed()
	pipe2.AssertAPINotUsed()
}

func cleanup(t *testing.T) {
	purgeQueue(t, testQueueMain)
	// TODO: empty sqs-sns queues
	// TODO: empty local data repository
	// TODO: empty local processing state table
}
