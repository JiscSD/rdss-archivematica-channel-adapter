package integration

import (
	"encoding/json"
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
	config = config.WithRegion(awsRegion)
	if *flagDebug {
		config = config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}
	config = config.WithCredentials(credentials.NewStaticCredentials(
		awsAccessKeyID, awsSecretAccessKey, awsTokenKey))
	config = config.WithS3ForcePathStyle(true)
	config.DisableSSL = aws.Bool(true)
	return session.Must(session.NewSession(config))
}

func s3Client() *s3.S3 {
	return s3.New(awsSession(awsS3Endpoint))
}

func dynamodbClient() *dynamodb.DynamoDB {
	return dynamodb.New(awsSession(awsDynamoDBEndpoint))
}

func sqsClient() *sqs.SQS {
	return sqs.New(awsSession(awsSQSEndpoint))
}

func snsClient() *sns.SNS {
	return sns.New(awsSession(awsSNSEndpoint))
}

func registerPipeline(t *testing.T, tenantJiscID int, p *ammock.Pipeline) {
	t.Helper()
	_, err := awsDynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(awsRegistryTable),
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
		QueueUrl:    aws.String(awsQueueMain),
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
		Bucket: aws.String(awsBucket),
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

func purgeDynamoDBTable(t *testing.T, table string, keyName string) {
	t.Helper()

	// Assuming that all entries fit in a single scan.
	res, err := awsDynamoDBClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String(table),
	})
	if err != nil {
		t.Fatal("Cannot scan DynamoDB table:", err)
	}

	reqs := []*dynamodb.WriteRequest{}
	for _, item := range res.Items {
		val := item[keyName]
		reqs = append(reqs, &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: map[string]*dynamodb.AttributeValue{
					keyName: val,
				},
			},
		})
	}

	awsDynamoDBClient.BatchWriteItem(&dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			table: reqs,
		},
	})
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
