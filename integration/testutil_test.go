package integration

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message/specdata"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/integration/ammock"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func awsSession(endpoint string) *session.Session {
	config := aws.NewConfig()
	config = config.WithEndpoint(endpoint)
	config = config.WithRegion(region)
	if *flagDebug {
		config = config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}
	config = config.WithCredentials(credentials.NewStaticCredentials(
		accessKeyID, secretAccessKey, tokenKey))
	config = config.WithS3ForcePathStyle(true)
	config.DisableSSL = aws.Bool(true)
	return session.Must(session.NewSession(config))
}

func s3Client() *s3.S3 {
	return s3.New(awsSession(s3Endpoint))
}

func dynamodbClient() *dynamodb.DynamoDB {
	return dynamodb.New(awsSession(dynamodbEndpoint))
}

func sqsClient() *sqs.SQS {
	return sqs.New(awsSession(sqsEndpoint))
}

func snsClient() *sns.SNS {
	return sns.New(awsSession(snsEndpoint))
}

func registerPipeline(t *testing.T, tenantJiscID int, p *ammock.Pipeline) {
	t.Helper()
	_, err := awsDynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(testRegistryTable),
		Item: map[string]*dynamodb.AttributeValue{
			"tenantJiscID": {S: aws.String(strconv.Itoa(tenantJiscID))},
			"url":          {S: aws.String(p.URL)},
			"user":         {S: aws.String(p.User)},
			"key":          {S: aws.String(p.Key)},
			"transferDir":  {S: aws.String(p.TransferDir)},
		},
	})
	if err != nil {
		t.Fatal("Cannot create registry item: ", err)
	}
}

// sendMessage sends a message to the main queue.
func sendMessage(t *testing.T, body string) {
	t.Helper()
	_, err := sqsClient().SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(testQueueMain),
		MessageBody: aws.String(body),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func putKnownObject(t *testing.T, key string) (string, int) {
	const blob = "known-blob"
	res, err := awsS3Client.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(strings.NewReader(blob)),
		Bucket: aws.String(testBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatal("Cannot store object in S3:", err)
	}
	return *res.ETag, len(blob)
}

func purgeQueue(t *testing.T, queueURL string) {
	t.Helper()
	_, err := awsSQSClient.PurgeQueue(&sqs.PurgeQueueInput{
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		t.Fatal("Cannot purge the queue: ", err)
	}
}

func newMetadataCreateMessage(t *testing.T, tenantJiscID uint64, objectTitle string, storagePlatform message.StorageTypeEnum, fileName, fileLocation, fileMD5 string, fileSize int) string {
	t.Helper()
	blob, err := specdata.Asset("messages/example_message.json")
	if err != nil {
		t.Fatal("Cannot read example_message.json fixture:", err)
	}
	m := message.New(message.MessageTypeEnum_MetadataCreate, message.MessageClassEnum_Command)
	err = json.Unmarshal(blob, m)
	if err != nil {
		t.Fatal("Cannot unmarshall example_message.json fixture:", err)
	}
	msg, err := m.MetadataCreateRequest()
	if err != nil {
		t.Fatal("Cannot read MetadataCreate body message:", err)
	}

	m.MessageHeader.ID = message.NewUUID()
	m.MessageHeader.TenantJiscID = tenantJiscID

	researchObject := msg.InferResearchObject()
	researchObject.ObjectUUID = message.NewUUID()
	researchObject.ObjectTitle = objectTitle
	researchObject.ObjectFile[0].FileStoragePlatform.StoragePlatformType = storagePlatform
	researchObject.ObjectFile[0].FileStorageLocation = fileLocation
	researchObject.ObjectFile[0].FileName = fileName
	researchObject.ObjectFile[0].FileSize = fileSize
	researchObject.ObjectFile[0].FileChecksum[0].ChecksumValue = fileMD5

	res, err := json.Marshal(m)
	if err != nil {
		t.Fatal("Cannot marshal created message:", err)
	}
	return string(res)
}

// subscriptions is a test util that knows how to verify that messages have been
// sent to SNS topics by asserting the contents of SQS queues that have been
// subscribed to the topics. Use the public Assert* methods to verify.
type subscriptions struct {
	t   *testing.T
	sns *sns.SNS
	sqs *sqs.SQS

	echoQueueURLMain    string
	echoQueueURLError   string
	echoQueueURLInvalid string
}

const (
	// Name of the queues where published messages are going to be echoed.
	echoQueueNameMain    = "sns-echo-main"
	echoQueueNameError   = "sns-echo-error"
	echoQueueNameInvalid = "sns-echo-invalid"
)

func subscriber(t *testing.T) *subscriptions {
	s := &subscriptions{
		t:   t,
		sns: awsSNSClient,
		sqs: awsSQSClient,
	}

	s.echoQueueURLMain = s.createQueue(echoQueueNameMain)
	purgeQueue(t, s.echoQueueURLMain)
	s.subscribeQueueToTopic(testTopicMain, echoQueueNameMain)

	s.echoQueueURLError = s.createQueue(echoQueueNameError)
	purgeQueue(t, s.echoQueueURLError)
	s.subscribeQueueToTopic(testTopicError, echoQueueNameError)

	s.echoQueueURLInvalid = s.createQueue(echoQueueNameInvalid)
	purgeQueue(t, s.echoQueueURLInvalid)
	s.subscribeQueueToTopic(testTopicInvalid, echoQueueNameInvalid)

	return s
}

func (s *subscriptions) createQueue(name string) string {
	res, err := s.sqs.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		s.t.Fatalf("Cannot create queue %s: %s", name, err)
	}
	return *res.QueueUrl
}

// subscribeQueueToTopic subscribes a SQS queue to a SNS topic.
func (s *subscriptions) subscribeQueueToTopic(topicARN, queueName string) {
	endpoint := fmt.Sprintf("arn:aws:sns:%s:%s:%s", region, accountID, queueName)
	_, err := s.sns.Subscribe(&sns.SubscribeInput{
		TopicArn: aws.String(topicARN),
		Protocol: aws.String("sqs"),
		Endpoint: aws.String(endpoint),
		Attributes: map[string]*string{
			"RawMessageDelivery": aws.String("true"),
		},
	})
	if err != nil {
		s.t.Fatal(err)
	}
}

func (s *subscriptions) receiveMessages(queueURL string, max, wait int64) []*sqs.Message {
	res, err := s.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(1),
		QueueUrl:            aws.String(queueURL),
		WaitTimeSeconds:     aws.Int64(wait),
	})
	if err != nil {
		s.t.Fatal(err)
	}
	return res.Messages
}

func (s *subscriptions) deleteMessage(t *testing.T, queueURL string, m *sqs.Message) {
	t.Helper()
	_, err := s.sqs.DeleteMessage(&sqs.DeleteMessageInput{
		ReceiptHandle: m.ReceiptHandle,
		QueueUrl:      aws.String(queueURL),
	})
	if err != nil {
		t.Fatal("Cannot delete the message:", err)
	}
}

func (s *subscriptions) assertMessageReceived(queueURL, message string) {
	s.t.Helper()
	msgs := s.receiveMessages(queueURL, 1, 1)
	if len(msgs) < 1 {
		s.t.Error("Message not received in queue", queueURL)
		return
	}
	s.deleteMessage(s.t, queueURL, msgs[0])
	if have, want := *msgs[0].Body, message; message != "" && have != want {
		s.t.Errorf("Unexpected message received in queue %s; have %+v, want %+v", queueURL, have, want)
	}
}

func (s *subscriptions) assertMessageNotReceived(queueURL string) {
	s.t.Helper()
	msgs := s.receiveMessages(queueURL, 1, 1)
	if len(msgs) > 0 {
		s.deleteMessage(s.t, queueURL, msgs[0])
		s.t.Errorf("Message received in queue %s (len=%d)", queueURL, len(msgs))
		return
	}
}

func (s *subscriptions) AssertMainMessageReceived(message string) {
	s.t.Helper()
	s.assertMessageReceived(s.echoQueueURLMain, message)
}

func (s *subscriptions) AssertMainMessageNotReceived() {
	s.t.Helper()
	s.assertMessageNotReceived(s.echoQueueURLMain)
}

func (s *subscriptions) AssertErrorMessageReceived(message string) {
	s.t.Helper()
	s.assertMessageReceived(s.echoQueueURLError, message)
}

func (s *subscriptions) AssertErrorMessageNotReceived() {
	s.t.Helper()
	s.assertMessageNotReceived(s.echoQueueURLError)
}

func (s *subscriptions) AssertInvalidMessageReceived(message string) {
	s.t.Helper()
	s.assertMessageReceived(s.echoQueueURLInvalid, message)
}

func (s *subscriptions) AssertInvalidMessageNotReceived() {
	s.t.Helper()
	s.assertMessageNotReceived(s.echoQueueURLInvalid)
}

func (s *subscriptions) AssertNoMoreIncomingMessages() {
	s.t.Helper()
	res, err := s.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(1),
		QueueUrl:            aws.String(testQueueMain),
		WaitTimeSeconds:     aws.Int64(0),
	})
	if err != nil {
		s.t.Fatal("Cannot check emptyness check on main queue:", err)
	}
	if len(res.Messages) > 0 {
		s.t.Fatal("Main queue is not empty")
	}
}
