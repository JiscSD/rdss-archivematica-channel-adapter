package broker

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

func TestRepositoryMessageState_String(t *testing.T) {
	tests := []struct {
		s    RepositoryMessageState
		want string
	}{
		{RepositoryMessageStateReceived, "RECEIVED"},
		{RepositoryMessageStateSent, "SENT"},
		{RepositoryMessageStateToSend, "TO_SEND"},
		{0, "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("RepositoryMessageState.String() = %v, want %v", got, tt.want)
		}
	}
}

func Test_toRepoMessage(t *testing.T) {
	tests := []struct {
		arg     *message.Message
		want    *RepositoryMessage
		wantErr bool
	}{
		{nil, nil, true},
		{
			&message.Message{MessageHeader: message.MessageHeader{ID: message.MustUUID("ab0f8186-4b68-430e-a07e-b517300e6f9f")}},
			&RepositoryMessage{MessageID: "ab0f8186-4b68-430e-a07e-b517300e6f9f"},
			false,
		},
	}
	for _, tt := range tests {
		msg, err := toRepoMessage(tt.arg)
		if tt.wantErr {
			if err == nil {
				t.Fatal()
			}
			return
		}
		if err != nil || msg == nil {
			t.Fatal()
		}
		if !reflect.DeepEqual(tt.want, msg) {
			t.Fatal()
		}
	}

}

func TestRepositoryDynamoDBImpl_Get(t *testing.T) {
	tests := []struct {
		name   string
		want   *RepositoryMessage
		client *mockDynamoDBClient
	}{
		{
			"Get item", &RepositoryMessage{},
			&mockDynamoDBClient{
				GetItem_WantedMsg: &RepositoryMessage{},
				GetItem_WantedErr: nil,
			},
		},
		{
			"dynamodb.get fails", nil,
			&mockDynamoDBClient{
				GetItem_WantedMsg: nil,
				GetItem_WantedErr: errors.New("error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepository(tt.client, "")
			if got, _ := r.Get("foo"); !reflect.DeepEqual(tt.want, got) {
				t.Errorf("RepositoryDynamoDBImpl.Get(); want %v, got %v", tt.want, got)
			}
		})
	}
}
func TestRepositoryDynamoDBImpl_Put(t *testing.T) {
	tests := []struct {
		arg     *message.Message
		wantErr bool
		client  *mockDynamoDBClient
	}{
		{
			message.New(message.MessageTypeEnum_MetadataCreate, message.MessageClassEnum_Command),
			false,
			&mockDynamoDBClient{},
		},
		{
			message.New(message.MessageTypeEnum_MetadataCreate, message.MessageClassEnum_Command),
			true,
			&mockDynamoDBClient{PutItem_WantedErr: errors.New("error")},
		},
		{nil, true, nil},
		// TODO: check case with dynamodbattribute.MarshalMap(rMsg) returning error
	}
	for _, tt := range tests {
		r := NewRepository(tt.client, "")
		err := r.Put(tt.arg)
		if tt.wantErr && err == nil {
			t.Error("RepositoryDynamoDBImpl.Get(); error expected but none returned")
		} else if !tt.wantErr && err != nil {
			t.Errorf("RepositoryDynamoDBImpl.Get(); error not expected but one returned: %v", err)
		}
	}
}

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	GetItem_WantedMsg interface{}
	GetItem_WantedErr error
	PutItem_WantedErr error
}

func (m *mockDynamoDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	item, err := dynamodbattribute.MarshalMap(m.GetItem_WantedMsg)
	if err != nil {
		return nil, err
	}
	return &dynamodb.GetItemOutput{Item: item}, m.GetItem_WantedErr
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, m.PutItem_WantedErr
}
