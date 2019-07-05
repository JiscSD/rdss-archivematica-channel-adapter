package broker

import (
	"errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// RepositoryMessage is a minifed version of message.Message meant to be stored
// in the local data repository as specified in the RDSS API docs.
type RepositoryMessage struct {
	MessageID    string                 `dynamodbav:"ID"`
	MessageClass string                 `dynamodbav:"messageClass"`
	MessageType  string                 `dynamodbav:"messageType"`
	Sequence     string                 `dynamodbav:"sequence"`
	Position     int                    `dynamodbav:"position"`
	Status       RepositoryMessageState `dynamodbav:"status"`
}

type RepositoryMessageState int

const (
	_                              RepositoryMessageState = iota
	RepositoryMessageStateReceived RepositoryMessageState = iota
	RepositoryMessageStateSent
	RepositoryMessageStateToSend
)

func (s RepositoryMessageState) String() string {
	switch s {
	case RepositoryMessageStateReceived:
		return "RECEIVED"
	case RepositoryMessageStateSent:
		return "SENT"
	case RepositoryMessageStateToSend:
		return "TO_SEND"
	default:
		return "UNKNOWN"
	}
}

type Repository interface {
	Get(ID string) (*RepositoryMessage, error)
	Put(*message.Message) error
}

func NewRepository(client dynamodbiface.DynamoDBAPI, table string) *RepositoryDynamoDBImpl {
	return &RepositoryDynamoDBImpl{
		client: client,
		table:  table,
	}
}

// RepositoryDynamoDBImpl implements Repository.
type RepositoryDynamoDBImpl struct {
	client dynamodbiface.DynamoDBAPI
	table  string
}

var _ Repository = (*RepositoryDynamoDBImpl)(nil)

func (r *RepositoryDynamoDBImpl) Get(ID string) (*RepositoryMessage, error) {
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
	msg := &RepositoryMessage{}
	if err := dynamodbattribute.UnmarshalMap(output.Item, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (r *RepositoryDynamoDBImpl) Put(msg *message.Message) error {
	rMsg, err := toRepoMessage(msg)
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

func toRepoMessage(msg *message.Message) (*RepositoryMessage, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}
	rMsg := &RepositoryMessage{}
	rMsg.MessageID = msg.ID()
	return rMsg, nil
}
