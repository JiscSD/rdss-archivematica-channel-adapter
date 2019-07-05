package broker_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message/specdata"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type sqsMock struct {
	sqsiface.SQSAPI
	count int
}

func (m *sqsMock) DeleteMessageWithContext(aws.Context, *sqs.DeleteMessageInput, ...request.Option) (*sqs.DeleteMessageOutput, error) {
	return &sqs.DeleteMessageOutput{}, nil
}

func (m *sqsMock) ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	m.count++
	switch m.count {
	case 1:
		blob := specdata.MustAsset("messages/example_message.json")
		return &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				&sqs.Message{
					Body: aws.String(string(blob)),
				},
			},
		}, nil
	default:
		// When this method is called again.
		time.Sleep(time.Millisecond * 1)
		return &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{},
		}, nil
	}
}

type snsMock struct {
	snsiface.SNSAPI
}

func (m *snsMock) PublishWithContext(aws.Context, *sns.PublishInput, ...request.Option) (*sns.PublishOutput, error) {
	return &sns.PublishOutput{}, nil
}

type dynaMock struct {
	dynamodbiface.DynamoDBAPI
}

func (m *dynaMock) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return &dynamodb.GetItemOutput{}, nil
}

func (m *dynaMock) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, nil
}

func ExampleNew() {
	// Create the broker client.
	b, _ := broker.New(
		logrus.StandardLogger(),
		&sqsMock{},
		"http://localhost:4576/queue/main",
		&snsMock{},
		"arn:aws:sns:us-east-1:123456789012:main",
		"arn:aws:sns:us-east-1:123456789012:invalid",
		"arn:aws:sns:us-east-1:123456789012:error",
		&dynaMock{},
		"local_data_repository",
		"strict",
		prometheus.NewCounter(prometheus.CounterOpts{}),
	)

	var wg sync.WaitGroup
	wg.Add(1)

	// We can subscribe a handler for a particular message type. The handler
	// is executed by the broker as soon as the message is received.
	b.Subscribe(message.MessageTypeEnum_MetadataCreate, func(m *message.Message) error {
		defer wg.Done()
		fmt.Println("[MetadataCreate] Message received!")
		return nil
	})

	// Run the broker client.
	go b.Run()

	// We can use the broker client to publish messages too.
	b.Metadata.Create(context.TODO(), &message.MetadataCreateRequest{})

	// Stop the broker client - but not until our handler runs.
	wg.Wait()
	b.Stop()

	// Output:
	// [MetadataCreate] Message received!
}
