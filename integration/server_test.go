// +build integration

// https://www.ardanlabs.com/blog/2019/03/integration-testing-in-go-executing-tests-with-docker.html

package integration

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/app"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	accessKeyID     = "123"
	secretAccessKey = "xyz"
	tokenKey        = ""
	region          = "us-east-1"

	s3Endpoint       = "http://127.0.0.1:4572"
	dynamodbEndpoint = "http://127.0.0.1:4569"
	sqsEndpoint      = "http://127.0.0.1:4576"
	snsEndpoint      = "http://127.0.0.1:4575"

	testBucket          = "test_adapter_bucket"
	testRepositoryTable = "test_adapter_repository"
	testProcessingTable = "test_adapter_processing"
	testQueueMain       = "test_adapter_queue_main"    // http://localhost:4576/queue/test_adapter_queue_main
	testTopicMain       = "test_adapter_topic_main"    // arn:aws:sns:us-east-1:123456789012:test_adapter_topic_main
	testTopicInvalid    = "test_adapter_topic_invalid" // arn:aws:sns:us-east-1:123456789012:test_adapter_topic_invalid
	testTopicError      = "test_adapter_topic_error"   // arn:aws:sns:us-east-1:123456789012:test_adapter_topic_error
)

var (
	s3_client       = s3Client()
	dynamodb_client = dynamodbClient()
	sqs_client      = sqsClient()
	sns_client      = snsClient()
	tmpdir, _       = ioutil.TempDir("", "")

	testQueueMainURL    = fmt.Sprintf("%s/queue/%s", sqsEndpoint, testQueueMain)
	testTopicMainARN    = fmt.Sprintf("arn:aws:sns:%s:123456789012:%s", region, testTopicMain)
	testTopicInvalidARN = fmt.Sprintf("arn:aws:sns:%s:123456789012:%s", region, testTopicInvalid)
	testTopicErrorARN   = fmt.Sprintf("arn:aws:sns:%s:123456789012:%s", region, testTopicError)
)

var (
	flagDebug = flag.Bool("debug", false, "")
)

func TestIntegration(t *testing.T) {
	run(t, []string{"server"})
}

func run(t *testing.T, args []string) {
	provision(t)
	config(t)

	done := make(chan struct{})

	go func() {
		var (
			stdout bytes.Buffer
			stderr bytes.Buffer
		)
		// cmd := app.RootCommand(bufio.NewWriter(&stdout), bufio.NewWriter(&stderr))
		cmd := app.RootCommand(os.Stdout, os.Stderr)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err, stdout, stderr)
		}
		println("completed?")
		done <- struct{}{}
	}()

	sqs_client.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(testQueueMainURL),
		MessageBody: aws.String("{}"),
	})

	<-done
}

func awsSession(endpoint string) *session.Session {
	config := aws.NewConfig()
	config = config.WithEndpoint(endpoint)
	config = config.WithRegion(region)
	if *flagDebug {
		config = config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}
	config = config.WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, tokenKey))
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

func provision(t *testing.T) {
	var err error

	// Create S3 bucket.
	{
		_, err = s3_client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String("bucket"),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create DynamoDB tables.
	{
		_, err = dynamodb_client.CreateTable(&dynamodb.CreateTableInput{
			TableName: aws.String("rdss_archivematica_processing"),
			KeySchema: []*dynamodb.KeySchemaElement{
				&dynamodb.KeySchemaElement{
					AttributeName: aws.String("ID"),
					KeyType:       aws.String("HASH"),
				},
			},
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				&dynamodb.AttributeDefinition{
					AttributeName: aws.String("ID"),
					AttributeType: aws.String("S"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10),
				WriteCapacityUnits: aws.Int64(10),
			},
		})
		if err != nil {
			aerr := err.(awserr.Error)
			switch aerr.Code() {
			default:
				t.Fatal(aerr)
			case dynamodb.ErrCodeResourceInUseException:
			}
		}

		_, err = dynamodb_client.CreateTable(&dynamodb.CreateTableInput{
			TableName: aws.String("rdss_archivematica_local_data_repository"),
			KeySchema: []*dynamodb.KeySchemaElement{
				&dynamodb.KeySchemaElement{
					AttributeName: aws.String("objectUUID"),
					KeyType:       aws.String("HASH"),
				},
			},
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				&dynamodb.AttributeDefinition{
					AttributeName: aws.String("objectUUID"),
					AttributeType: aws.String("S"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10),
				WriteCapacityUnits: aws.Int64(10),
			},
		})
		if err != nil {
			aerr := err.(awserr.Error)
			switch aerr.Code() {
			default:
				t.Fatal(aerr)
			case dynamodb.ErrCodeResourceInUseException:
			}
		}
	}

	// Create SQS queue.
	{
		_, err = sqs_client.CreateQueue(&sqs.CreateQueueInput{
			Attributes: map[string]*string{
				"ReceiveMessageWaitTimeSeconds": aws.String("20"),
				"VisibilityTimeout":             aws.String("30"),
				"FifoQueue":                     aws.String("false"),
			},
			QueueName: aws.String(testQueueMain),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create SNS topics.
	{
		_, err = sns_client.CreateTopic(&sns.CreateTopicInput{
			Name: aws.String(testTopicMain),
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = sns_client.CreateTopic(&sns.CreateTopicInput{
			Name: aws.String(testTopicInvalid),
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err := sns_client.CreateTopic(&sns.CreateTopicInput{
			Name: aws.String(testTopicError),
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func config(t *testing.T) {
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AMCLIENT.TRANSFER_DIR", tmpdir)
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REPOSITORY_TABLE", testRepositoryTable)
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.PROCESSING_TABLE", testProcessingTable)
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_RECV_MAIN_ADDR", fmt.Sprintf("%s/queue/%s", sqsEndpoint, testQueueMain))
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_MAIN_ADDR", fmt.Sprintf("arn:aws:sns:%s:123456789012:%s", region, testTopicMain))
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_INVALID_ADDR", fmt.Sprintf("arn:aws:sns:%s:123456789012:%s", region, testTopicInvalid))
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_EROR_ADDR", fmt.Sprintf("arn:aws:sns:%s:123456789012:%s", region, testTopicError))
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_PROFILE", "dynamodb")
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_ENDPOINT", dynamodbEndpoint)
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_PROFILE", "s3")
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_ENDPOINT", s3Endpoint)
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_PROFILE", "sqs")
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_ENDPOINT", sqsEndpoint)
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_PROFILE", "sns")
	os.Setenv("RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_ENDPOINT", snsEndpoint)
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	os.Setenv("AWS_CONFIG_FILE", "/home/jesus/my_shared_config")
}
