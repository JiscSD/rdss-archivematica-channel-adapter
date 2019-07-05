package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/pkg/errors"
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
	sqsClient          sqsiface.SQSAPI
	sqsQueueMainURL    string
	snsClient          snsiface.SNSAPI
	snsTopicMainARN    string
	snsTopicInvalidARN string
	snsTopicErrorARN   string
	repo               Repository
	validator          message.Validator
	ctx                context.Context
	cancel             context.CancelFunc
	messages           chan *sqs.Message
	stop               chan chan struct{}
	Metadata           MetadataService
	incomingMessages   prometheus.Counter
	subscriptions
}

func New(
	logger logrus.FieldLogger,
	sqsClient sqsiface.SQSAPI, sqsQueueMainURL string,
	snsClient snsiface.SNSAPI, snsTopicMainARN, snsTopicInvalidARN, snsTopicErrorARN string,
	repo Repository,
	validationMode string,
	incomingMessages prometheus.Counter) (*Broker, error) {
	b := &Broker{
		logger:             logger,
		sqsClient:          sqsClient,
		sqsQueueMainURL:    sqsQueueMainURL,
		snsTopicMainARN:    snsTopicMainARN,
		snsTopicInvalidARN: snsTopicInvalidARN,
		snsTopicErrorARN:   snsTopicErrorARN,
		snsClient:          snsClient,
		repo:               repo,
		messages:           make(chan *sqs.Message),
		stop:               make(chan chan struct{}),
		incomingMessages:   incomingMessages,
	}

	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.subscriptions.s = make(map[message.MessageTypeEnum]MessageHandler)
	b.Metadata = &MetadataServiceOp{broker: b}

	var err error
	b.validator, err = message.NewValidator(validationMode)
	if err != nil {
		return nil, errors.Wrap(err, "validator setup failed")
	}

	return b, nil
}

func (b *Broker) Run() {
	go b.processor()
	b.loop()
}

// processor launches a processing goroutine for each message received.
func (b *Broker) processor() {
	for m := range b.messages {
		go func(m *sqs.Message) {
			go b.processMessage(m)
		}(m)
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

func (b *Broker) processMessage(m *sqs.Message) {
	b.incomingMessages.Inc()

	// Payload unmarshal.
	msg := &message.Message{}
	err := json.Unmarshal([]byte(*m.Body), msg)
	if err != nil {
		b.invalidMessage(m, bErrors.NewWithError(bErrors.GENERR001, err))
		return
	}

	// Payload validation.
	if err := b.validate(msg); err != nil {
		b.invalidMessage(m, bErrors.NewWithError(bErrors.GENERR001, err))
		return
	}

	logger := b.logger.WithFields(logrus.Fields{
		"messageID": msg.ID(),
		"type":      msg.MessageHeader.MessageType.String(),
		"class":     msg.MessageHeader.MessageClass.String(),
	})

	// Do nothing if the message has been seen before.
	// Best effort, i.e. we continue on errors checking the local repo repo.
	found, err := b.seenBefore(msg)
	if err != nil {
		logger.Error("Local data repository check failed: ", err)
	} else if found {
		logger.Warning("Message found in the local data repository.")
		b.deleteMessage(m.ReceiptHandle)
		return
	}

	// Run the handler in panic recovery mode.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("handler goroutine panic! %s", r)
			}
		}()
		err = b.HandleMessage(msg)
	}()
	wg.Wait()

	if err != nil {
		logger.Error("Handler failure: ", err)
		b.errorMessage(msg, bErrors.NewWithError(bErrors.GENERR006, err), m.ReceiptHandle)
		return
	}

	b.deleteMessage(m.ReceiptHandle)
}

// deleteMessage does best effort to delete a message from SQS but it does not
// change behaviour in the caller, i.e. error is logged but not returned.
func (b *Broker) deleteMessage(receiptHandle *string) {
	_, err := b.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(b.sqsQueueMainURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		b.logger.Error("Message could not be removed from SQS: ", err)
	}
}

// publishMessage puts a message into a topic.
func (b *Broker) publishMessage(topicARN string, payload string) error {
	_, err := b.snsClient.PublishWithContext(b.ctx, &sns.PublishInput{
		Message:  aws.String(payload),
		TopicArn: aws.String(topicARN),
	})
	return err
}

// seenBefore returns whether the message is already in the repository.
//
// As a side effect, the message is cached in the repo when it wasn't there so
// the second time this function is called for the same message the returned
// value should be true.
func (b *Broker) seenBefore(msg *message.Message) (bool, error) {
	item, err := b.repo.Get(msg.ID())
	if err != nil {
		return false, err
	}
	if item != nil {
		return true, nil
	}
	if err := b.repo.Put(msg); err != nil {
		return false, err
	}
	return false, nil
}

// validate returns whether the message is valid according to the spec.
//
// The logging events sent in this method are documented - don't change them!
func (b *Broker) validate(msg *message.Message) error {
	res, err := b.validator.Validate(msg)
	if err != nil {
		return errors.Wrap(err, "validation failed")
	}
	if res.Valid() {
		return nil
	}
	message.ValidateVersion(msg.MessageHeader.Version, res)

	var (
		logger  = b.logger.WithField("messageID", msg.ID())
		valErrs = res.Errors()
		count   = len(valErrs)
	)
	logger.Debugf("JSON Schema validator found %d issues.", count)
	for _, re := range valErrs {
		b.logger.WithFields(logrus.Fields{
			"messageID": msg.ID(),
		}).Debugf("- %s", re.Description())
	}
	if _, ok := b.validator.(*message.NoOpValidator); ok {
		return nil
	}
	return fmt.Errorf("message has unexpected format, %d errors found", count)
}

// invalidMessage puts a message into the Invalid Message Queue.
func (b *Broker) invalidMessage(m *sqs.Message, specErr error) {
	defer b.deleteMessage(m.ReceiptHandle)
	if err := b.publishMessage(b.snsTopicInvalidARN, *m.Body); err != nil {
		b.logger.Error("A message could not be sent to the Invalid Message Queue: ", err)
	}
	b.logger.Debug("Message sent ot the invalid messge queue")
}

// errorMessage puts a message into the Error Message Queue.
func (b *Broker) errorMessage(msg *message.Message, specErr error, receiptHandle *string) {
	defer b.deleteMessage(receiptHandle)
	msg.TagError(specErr)
	logger := b.logger.WithFields(logrus.Fields{"id": msg.ID(), "specErr": specErr})
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("A message could not be marshalled before sending to the Error Message Queue: ", err)
		return
	}
	if err = b.publishMessage(b.snsTopicErrorARN, string(data)); err != nil {
		logger.Error("A message could not be sent to the Error Message Queue: ", err)
	}
	b.logger.Debug("Message sent ot the error messge queue")
}

func (b *Broker) Stop() {
	ch := make(chan struct{})
	b.stop <- ch
	<-ch
}
