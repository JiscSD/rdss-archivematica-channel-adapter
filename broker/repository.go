package broker

import (
	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
)

// repositoryMessage is a minifed version of message.Message meant to be stored
// in the local data repository as specified in the RDSS API docs.
type repositoryMessage struct {
	MessageID    string                 `dynamodbav:"ID"`
	MessageClass string                 `dynamodbav:"messageClass"`
	MessageType  string                 `dynamodbav:"messageType"`
	Sequence     string                 `dynamodbav:"sequence"`
	Position     int                    `dynamodbav:"position"`
	Status       repositoryMessageState `dynamodbav:"status"`
}

type repositoryMessageState int

const (
	_                              repositoryMessageState = iota
	repositoryMessageStateReceived repositoryMessageState = iota
	repositoryMessageStateSent
	repositoryMessageStateToSend
)

func (s repositoryMessageState) String() string {
	switch s {
	case repositoryMessageStateReceived:
		return "RECEIVED"
	case repositoryMessageStateSent:
		return "SENT"
	case repositoryMessageStateToSend:
		return "TO_SEND"
	default:
		return "UNKNOWN"
	}
}

type repository struct {
	client dynamodbiface.DynamoDBAPI
	table  string
}

// seenBeforeOrStore decides whether a message is known to this repository.
func (r *repository) seenBeforeOrStore(m *message.Message) (bool, error) {
	item, err := r.getRecord(m.ID())
	if err != nil {
		return false, err
	}
	if item != nil {
		return true, nil
	}
	if err := r.putRecord(m); err != nil {
		return false, err
	}
	return false, nil
}

func (r *repository) getRecord(ID string) (*repositoryMessage, error) {
	output, err := r.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(ID)},
		},
	})
	if err != nil {
		return nil, err
	}
	if output.Item == nil {
		return nil, nil
	}
	msg := &repositoryMessage{}
	if err := dynamodbattribute.UnmarshalMap(output.Item, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (r *repository) putRecord(m *message.Message) error {
	rMsg, err := toRepoMessage(m)
	if err != nil {
		return err
	}
	item, err := dynamodbattribute.MarshalMap(rMsg)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	}
	_, err = r.client.PutItem(input)
	return err
}

func toRepoMessage(m *message.Message) (*repositoryMessage, error) {
	if m == nil {
		return nil, errors.New("message is nil")
	}
	rMsg := &repositoryMessage{}
	rMsg.MessageID = m.ID()
	return rMsg, nil
}
