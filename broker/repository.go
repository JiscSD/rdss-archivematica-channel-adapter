package broker

import (
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

var ErrInvalidRepositoryConfig = errors.New("invalid repository configuration")

// RepositoryMessage is a minifed version of message.Message meant to be stored
// in the local data repository as specified in the RDSS API docs.
type RepositoryMessage struct {
	MessageId    string                 `dynamodbav:"ID"`
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
	Get(ID string) *RepositoryMessage
	Put(*message.Message) error
}

func NewRepository(config *RepositoryConfig) (Repository, error) {
	if config.Backend == "dynamodb" {
		return &RepositoryDynamoDBImpl{
			DynamoDB: getDynamoDBInstance(config),
			Table:    config.DynamoDBTable,
		}, nil
	} else if config.Backend == "builtin" {
		return &RepositoryBuiltinImpl{msgs: make(map[string]*RepositoryMessage)}, nil
	}
	return nil, ErrInvalidRepositoryConfig
}

// MustRepository is a helper that wraps a call to a function returning
// (Repository, error) and panics if the error is non-nil. It is intended for
// use in variable initializations such as
//	var t = template.Must(template.New("name").Parse("text"))
func MustRepository(r Repository, err error) Repository {
	if err != nil {
		panic(err)
	}
	return r
}

// RepositoryDynamoDBImpl implements Repository.
type RepositoryDynamoDBImpl struct {
	DynamoDB dynamodbiface.DynamoDBAPI
	Table    string
}

var _ Repository = (*RepositoryDynamoDBImpl)(nil)

func (r *RepositoryDynamoDBImpl) Get(ID string) *RepositoryMessage {
	var input = &dynamodb.GetItemInput{
		TableName: aws.String(r.Table),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(ID)},
		},
	}
	output, err := r.DynamoDB.GetItem(input)
	if err != nil || output.Item == nil {
		return nil
	}
	msg := &RepositoryMessage{}
	if err := dynamodbattribute.UnmarshalMap(output.Item, msg); err != nil {
		return nil
	}
	return msg
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
		TableName: aws.String(r.Table),
		Item:      item,
	}
	_, err = r.DynamoDB.PutItem(input)
	return err
}

func getDynamoDBInstance(config *RepositoryConfig) dynamodbiface.DynamoDBAPI {
	awsCfg := aws.NewConfig()
	if config.DynamoDBRegion != "" {
		awsCfg = awsCfg.WithRegion(config.DynamoDBRegion)
	}
	if config.DynamoDBEndpoint != "" {
		awsCfg = awsCfg.WithEndpoint(config.DynamoDBEndpoint)
	}
	awsCfg.DisableSSL = aws.Bool(!config.DynamoDBTLS)

	return dynamodb.New(session.Must(session.NewSession(awsCfg)))
}

func toRepoMessage(msg *message.Message) (*RepositoryMessage, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}
	rMsg := &RepositoryMessage{}
	rMsg.MessageId = msg.ID()
	return rMsg, nil
}

// RepositoryInMemoryImpl is a memory-based Repository.
type RepositoryBuiltinImpl struct {
	msgs map[string]*RepositoryMessage
	sync.RWMutex
}

var _ Repository = (*RepositoryBuiltinImpl)(nil)

func (r *RepositoryBuiltinImpl) Get(ID string) *RepositoryMessage {
	r.RLock()
	defer r.RUnlock()
	if rMsg, ok := r.msgs[ID]; ok {
		return rMsg
	}
	return nil
}

func (r *RepositoryBuiltinImpl) Put(msg *message.Message) error {
	r.Lock()
	defer r.Unlock()
	rMsg, err := toRepoMessage(msg)
	if err != nil {
		return err
	}
	r.msgs[msg.ID()] = rMsg
	return nil
}
