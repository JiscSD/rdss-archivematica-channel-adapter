package integration

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

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
	s.subscribeQueueToTopic(awsTopicMain, echoQueueNameMain)

	s.echoQueueURLError = s.createQueue(echoQueueNameError)
	s.subscribeQueueToTopic(awsTopicError, echoQueueNameError)

	s.echoQueueURLInvalid = s.createQueue(echoQueueNameInvalid)
	s.subscribeQueueToTopic(awsTopicInvalid, echoQueueNameInvalid)

	return s
}

func (s *subscriptions) createQueue(name string) string {
	s.t.Helper()
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
	s.t.Helper()
	endpoint := fmt.Sprintf("arn:aws:sns:%s:%s:%s", awsRegion, awsAccountID, queueName)
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
	s.t.Helper()
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

func (s *subscriptions) cleanUp() {
	s.t.Helper()

	// Delete topic subscriptions.
	res, err := awsSNSClient.ListSubscriptions(&sns.ListSubscriptionsInput{})
	if err != nil {
		s.t.Fatal("Cannot list subscriptions:", err)
	}
	for _, s := range res.Subscriptions {
		awsSNSClient.Unsubscribe(&sns.UnsubscribeInput{
			SubscriptionArn: s.SubscriptionArn,
		})
	}

	purgeQueue(s.t, s.echoQueueURLMain)
	purgeQueue(s.t, s.echoQueueURLError)
	purgeQueue(s.t, s.echoQueueURLInvalid)
}

// Queue observability helpers.

func (s *subscriptions) assertMessageReceived(queueURL, message string) {
	s.t.Helper()
	msgs := s.receiveMessages(queueURL, 1, 5)
	if len(msgs) < 1 {
		s.t.Error("Message not received in queue", queueURL)
		return
	}
	s.deleteMessage(s.t, queueURL, msgs[0])
	if have, want := *msgs[0].Body, message; message != "" && have != want {
		s.t.Errorf("Unexpected message received in queue %s; have %+v, want %+v", queueURL, have, want)
	}
}

//nolint: unused
func (s *subscriptions) assertMessageNotReceived(queueURL string) {
	s.t.Helper()
	msgs := s.receiveMessages(queueURL, 1, 1)
	if len(msgs) > 0 {
		s.deleteMessage(s.t, queueURL, msgs[0])
		s.t.Errorf("Message received in queue %s (len=%d)", queueURL, len(msgs))
		return
	}
}

func (s *subscriptions) assertQueueIsEmpty(queueURL string) {
	s.t.Helper()
	c := s.countQueueItems(s.echoQueueURLInvalid)
	count, _ := strconv.Atoi(c)
	if count > 0 {
		s.t.Fatal("Invalid queue is not empty")
	}
}

func (s *subscriptions) countQueueItems(queueURL string) string {
	s.t.Helper()
	res, err := awsSQSClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueURL),
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
		},
	})
	if err != nil {
		s.t.Fatal("Cannot read queue attributes:", err)
	}
	return *res.Attributes["ApproximateNumberOfMessages"]
}

// Main message topic.

func (s *subscriptions) AssertMainMessageReceived(message string) {
	s.t.Helper()
	s.assertMessageReceived(s.echoQueueURLMain, message)
}

func (s *subscriptions) AssertMainQueueIsEmpty() {
	s.t.Helper()
	s.assertQueueIsEmpty(s.echoQueueURLError)
}

// Error message topic.

func (s *subscriptions) AssertErrorMessageReceived(message string) {
	s.t.Helper()
	s.assertMessageReceived(s.echoQueueURLError, message)
}

func (s *subscriptions) AssertErrorQueueIsEmpty() {
	s.t.Helper()
	s.assertQueueIsEmpty(s.echoQueueURLError)
}

// Invalid message topic.

func (s *subscriptions) AssertInvalidMessageReceived(message string) {
	s.t.Helper()
	s.assertMessageReceived(s.echoQueueURLInvalid, message)
}

func (s *subscriptions) AssertInvalidQueueIsEmpty() {
	s.t.Helper()
	s.assertQueueIsEmpty(s.echoQueueURLInvalid)
}

// Incoming message queue.

func (s *subscriptions) AssertInconmingQueueIsEmpty() {
	s.t.Helper()
	s.assertQueueIsEmpty(awsQueueMain)
}
