package integration

import (
	"flag"
	"fmt"
	"io/ioutil"
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
	flagDebug                 = flag.Bool("debug", false, "")
	flagValidationServiceAddr = flag.String("valsvc", "", "Address of the Schema Service HTTP API")
)

// TestValidation confirms that the adapter reacts to invalid messages by
// sending them to the corresponding invalid message queue.
func TestValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if *flagValidationServiceAddr == "" {
		t.Skip("skipping: needs -valsvc flag")
	}
	defer cleanUp(t)

	var env = serverEnvironment
	if *flagValidationServiceAddr != "" {
		env = append(serverEnvironment[:0:0], serverEnvironment...)
		env = append(env, fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.VALIDATION_SERVICE_ADDR=%s", *flagValidationServiceAddr))
	}

	s := subscriber(t)
	defer s.cleanUp()

	invalidMessage := `{
  "messageHeader": {
	"version": "4.0.0",
	"messageType": "MetadataCreate"
  },
  "messageBody": {}
}`

	sendMessage(t, invalidMessage)

	cmd := adapter.Server().WithEnv(env)
	stop := cmd.RunBackground(t)
	defer stop()

	s.AssertInvalidMessageReceived(invalidMessage)
	s.AssertInvalidQueueIsEmpty()
}

// TestConversion confirms that the adapter converts old messages relying on the
// schema service.
func TestConversion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if *flagValidationServiceAddr == "" {
		t.Skip("skipping: needs -valsvc flag")
	}
	defer cleanUp(t)

	var env = serverEnvironment
	if *flagValidationServiceAddr != "" {
		env = append(serverEnvironment[:0:0], serverEnvironment...)
		env = append(env, fmt.Sprintf("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.VALIDATION_SERVICE_ADDR=%s", *flagValidationServiceAddr))
	}

	s := subscriber(t)
	defer s.cleanUp()

	// Create pipeline and store it in the registry.
	pipeline := ammock.New(t)
	defer pipeline.Stop()
	registerPipeline(t, 133, pipeline)

	// message-metadata-create-v302.json has the corresponding attributes
	// populated, e.g. checksum, size, etc...
	putKnownObject(t, "dataset.zip")

	cmd := adapter.Server().WithEnv(env)
	stop := cmd.RunBackground(t)
	defer stop()

	// Send MetadataCreate v3.0.2 message using tenantJiscID=1.
	msg, _ := ioutil.ReadFile("./testdata/message-metadata-create-v302.json")
	sendMessage(t, string(msg))

	// We're expecting a presevation event from Archivematica.
	// TODO: verify the payload.
	s.AssertMainMessageReceived("")

	pipeline.AssertAPIUsed()
	pipeline.AssertTransferDirIsNotEmpty()
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
