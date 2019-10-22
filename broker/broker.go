package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	bErrors "github.com/JiscSD/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	// maxNumberOfMessages is the number of messages that we want to receive
	// from SQS incoming batches.
	maxNumberOfMessages = 1

	// waitTimeSeconds is the longest we're waiting on each SQS receive poll.
	waitTimeSeconds = 1
)

// Broker is a RDSS client using the SQS and SNS services.
//
// Messages are received from sqsQueueMainURL and sent to an interna channel
// (messages). The channel is unbuffered so the receiver controls how often we
// are going to receive from SQS. However, the current processor is unbounded,
// i.e. processMessage is launched on a new goroutine for each message received.
//
// The message processor will:
//
// * Extract, unmarshal and validate the message payload.
//
// * Reject messages that have been received before.
//
// * Run the designated handler and capture the returned error.
//
// In case of errors, messages are sent to the {Invalid,Error} Message Queue
// according to the behaviour described in the RDSS API specification.
//
// Messages are deleted from SQS as soon as they're processed. This includes
// cases where the processing have failed, e.g. validation or handler error.
// The visibility timeout is set by the SQS queue owner under the assumption
// that the underlying preservation system is capable to process the requests
// within the window given (the maximum is 12 hours).
//
// Potential improvements:
//
// * Create a limited number of processors to avoid bursting.
//
// * Increase throughput: sqs.DeleteMessageBatch, multiple consumers, etc...
//   Low priority since we don't expect many messages.
//
// * Handlers could take a long time to complete. Do we need cancellation?
//   What are we doing when we exceed the visibility timeout? Is the adapter
//   accountable?
//
type Broker struct {
	logger             logrus.FieldLogger
	validator          message.Validator
	sqsClient          sqsiface.SQSAPI
	sqsQueueMainURL    string
	snsClient          snsiface.SNSAPI
	snsTopicMainARN    string
	snsTopicInvalidARN string
	snsTopicErrorARN   string
	ctx                context.Context
	cancel             context.CancelFunc
	messages           chan *sqs.Message
	stop               chan chan struct{}
	Metadata           MetadataService
	Preservation       PreservationService
	incomingMessages   prometheus.Counter
	subscriptions
	repository
}

// New returns a usable Broker.
func New(
	logger logrus.FieldLogger, validator message.Validator,
	sqsClient sqsiface.SQSAPI, sqsQueueMainURL string,
	snsClient snsiface.SNSAPI, snsTopicMainARN, snsTopicInvalidARN, snsTopicErrorARN string,
	dynamodbClient dynamodbiface.DynamoDBAPI, dynamodbTable string,
	incomingMessages prometheus.Counter) *Broker {
	b := &Broker{
		logger:             logger,
		validator:          validator,
		sqsClient:          sqsClient,
		sqsQueueMainURL:    sqsQueueMainURL,
		snsTopicMainARN:    snsTopicMainARN,
		snsTopicInvalidARN: snsTopicInvalidARN,
		snsTopicErrorARN:   snsTopicErrorARN,
		snsClient:          snsClient,
		messages:           make(chan *sqs.Message),
		stop:               make(chan chan struct{}),
		incomingMessages:   incomingMessages,
		repository:         repository{client: dynamodbClient, table: dynamodbTable},
	}
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.subscriptions.s = make(map[message.MessageTypeEnum]MessageHandler)
	b.Metadata = &MetadataServiceOp{broker: b}
	b.Preservation = &PreservationServiceOp{broker: b}

	go b.processor()

	return b
}

// Run starts the processing.
func (b *Broker) Run() {
	b.loop()
}

// processor of delivered messages. Processing is performed in two phases:
//
// Phase 1: extract the payload, validate it and store the ID in the local data
// repository. This is a blocking operation because we want to prevent the
// consuming from processing until we have a chance to update the local data
// repository.
//
// Phase 2: launch a goroutine to handle the message to a handler and perform
// the rest of the processing asynchronously.
func (b *Broker) processor() {
	for m := range b.messages {
		msg, err := b.openMessage(m)
		if err != nil {
			b.deleteMessage(m.ReceiptHandle)
			continue
		}
		go b.processMessage(m.ReceiptHandle, msg)
	}
}

// loop sends messages received from sqsQueueMainURL to the internal messages
// channel which is unbuffered so the receiver has control over how often we
// receive.
func (b *Broker) loop() {
	for {
		select {
		case ch := <-b.stop:
			b.cancel()
			close(b.messages)
			close(ch)
			return
		default:
			out, err := b.sqsClient.ReceiveMessageWithContext(b.ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(b.sqsQueueMainURL),
				MaxNumberOfMessages: aws.Int64(maxNumberOfMessages),
				WaitTimeSeconds:     aws.Int64(waitTimeSeconds),
			})
			if err != nil {
				b.logger.Errorf("Error receiving a message from SQS: %s", err)
				time.Sleep(1 * time.Second)
			} else {
				for _, m := range out.Messages {
					b.messages <- m
				}
			}
		}
	}
}

// openMessage performs initial validation and returns the underlying RDSS
// message.
func (b *Broker) openMessage(m *sqs.Message) (*message.Message, error) {
	b.incomingMessages.Inc()

	var stream = []byte(*m.Body)

	// Pass the message through the validation/transformation service.
	result, err := b.validator.Validate(b.ctx, stream)

	// We give up when the validator reports validation issues, but we'll
	// continue in case of other errors, e.g. service is down.
	var validErr = &message.ValidationError{}
	if errors.As(err, validErr) {
		b.invalidMessage(m, bErrors.NewWithError(bErrors.GENERR001, err))
		b.logger.Warning("Validation service reported schema issues: ", validErr)
		return nil, err
	}
	if err != nil {
		b.logger.Warning("Validation service reported a problem: ", err)
	} else {
		stream = result // Use the validator stream only if error-free.
	}

	// Payload unmarshal.
	msg := &message.Message{}
	err = json.Unmarshal(stream, msg)
	if err != nil {
		b.invalidMessage(m, bErrors.NewWithError(bErrors.GENERR001, err))
		return nil, err
	}

	if msg.MessageHeader.Version != message.Version {
		err := fmt.Errorf("version %s is not supported, only %s", msg.MessageHeader.Version, message.Version)
		b.invalidMessage(m, bErrors.NewWithError(bErrors.GENERR001, err))
		return nil, err
	}

	seen, err := b.seenBeforeOrStore(msg)

	// Not having access to the local data repository should not be a reason
	// to prevent its processing hence we return nil.
	if err != nil {
		b.logger.Warning("Local data repository check failed: ", err)
		return nil, err
	}

	// Giving up on known messages.
	if seen {
		b.logger.Warning("Message found in the local data repository.")
		return nil, errors.New("message seen")
	}

	return msg, nil
}

// processMessage handles the message to the handler. The message is deleted
// from the queue when the handler completes without errors.
func (b *Broker) processMessage(receiptHandle *string, msg *message.Message) {
	logger := b.logger.WithFields(logrus.Fields{
		"messageID": msg.ID(),
		"type":      msg.MessageHeader.MessageType.String(),
		"class":     msg.MessageHeader.MessageClass.String(),
	})

	var (
		err error
		wg  sync.WaitGroup
	)

	// Run the handler in panic recovery mode.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("handler goroutine panic! %s %s", r, debug.Stack())
			}
		}()
		err = b.handleMessage(msg)
	}()
	wg.Wait()

	if err != nil {
		logger.Error("Handler failure: ", err)
		b.errorMessage(msg, bErrors.NewWithError(bErrors.GENERR006, err), receiptHandle)
		return
	}

	b.deleteMessage(receiptHandle)
}

// deleteMessage does best effort to delete a message from SQS. It does not
// return since we're not reacting to them at the moment.
func (b *Broker) deleteMessage(receiptHandle *string) {
	_, err := b.sqsClient.DeleteMessageWithContext(b.ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(b.sqsQueueMainURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		b.logger.Error("Message could not be removed from SQS: ", err)
	}
}

// publishMessage puts a message into a SNS topic.
func (b *Broker) publishMessage(topicARN string, payload string) error {
	_, err := b.snsClient.PublishWithContext(b.ctx, &sns.PublishInput{
		Message:  aws.String(payload),
		TopicArn: aws.String(topicARN),
	})
	return err
}

// invalidMessage puts a message into the Invalid Message Queue.
func (b *Broker) invalidMessage(m *sqs.Message, specErr error) {
	arn := b.snsTopicInvalidARN
	if arn == "" {
		b.logger.WithField("error-queue", "invalid[disabled]").Warn(specErr)
		return
	}

	if err := b.publishMessage(arn, *m.Body); err != nil {
		b.logger.Error("A message could not be sent to the Invalid Message Queue: ", err)
	}
	b.logger.Debug("Message sent to the Invalid Message Queue")
}

// errorMessage puts a message into the Error Message Queue.
func (b *Broker) errorMessage(msg *message.Message, specErr error, receiptHandle *string) {
	defer b.deleteMessage(receiptHandle)

	arn := b.snsTopicErrorARN
	if arn == "" {
		b.logger.WithField("error-queue", "error[disabled]").Warn(specErr)
		return
	}

	msg.TagError(specErr)
	logger := b.logger.WithFields(logrus.Fields{"id": msg.ID(), "specErr": specErr})
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("A message could not be marshalled before sending to the Error Message Queue: ", err)
		return
	}
	if err = b.publishMessage(arn, string(data)); err != nil {
		logger.Error("A message could not be sent to the Error Message Queue: ", err)
	}
	b.logger.Debug("Message sent to the Error Message Queue")
}

// Request sends a fire-and-forget request to RDSS.
func (b *Broker) Request(_ context.Context, msg *message.Message) error {
	payload, err := msg.MarshalJSON()
	if err != nil {
		return err
	}
	return b.publishMessage(b.snsTopicMainARN, string(payload))
}

// RequestResponse sends a request and waits until a response is received.
func (b *Broker) RequestResponse(context.Context, *message.Message) (*message.Message, error) {
	return nil, errors.New("not implemented yet")
}

// Stop blocks until the broker terminates.
func (b *Broker) Stop() {
	ch := make(chan struct{})
	b.stop <- ch
	<-ch
}
