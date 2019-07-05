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

func TestRepositoryMessageStateString(t *testing.T) {
	tests := []struct {
		s    repositoryMessageState
		want string
	}{
		{repositoryMessageStateReceived, "RECEIVED"},
		{repositoryMessageStateSent, "SENT"},
		{repositoryMessageStateToSend, "TO_SEND"},
		{0, "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("RepositoryMessageState.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestRepositoryGetRecord(t *testing.T) {
	tests := []struct {
		name   string
		want   *repositoryMessage
		client *mockDynamoDBClient
	}{
		{
			"Get item", &repositoryMessage{},
			&mockDynamoDBClient{
				getItemWantedMsg: &repositoryMessage{},
				getItemWantedErr: nil,
			},
		},
		{
			"dynamodb.get fails", nil,
			&mockDynamoDBClient{
				getItemWantedMsg: nil,
				getItemWantedErr: errors.New("error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repository{client: tt.client, table: ""}
			if got, _ := r.getRecord("foo"); !reflect.DeepEqual(tt.want, got) {
				t.Errorf("RepositoryDynamoDBImpl.Get(); want %v, got %v", tt.want, got)
			}
		})
	}
}
func TestRepositoryPutRecord(t *testing.T) {
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
			&mockDynamoDBClient{putItemWantedErr: errors.New("error")},
		},
		{nil, true, nil},
		// TODO: check case with dynamodbattribute.MarshalMap(rMsg) returning error
	}
	for _, tt := range tests {
		r := repository{client: tt.client, table: ""}
		err := r.putRecord(tt.arg)
		if tt.wantErr && err == nil {
			t.Error("RepositoryDynamoDBImpl.Get(); error expected but none returned")
		} else if !tt.wantErr && err != nil {
			t.Errorf("RepositoryDynamoDBImpl.Get(); error not expected but one returned: %v", err)
		}
	}
}

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	getItemWantedMsg interface{}
	getItemWantedErr error
	putItemWantedErr error
}

func (m *mockDynamoDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	item, err := dynamodbattribute.MarshalMap(m.getItemWantedMsg)
	if err != nil {
		return nil, err
	}
	return &dynamodb.GetItemOutput{Item: item}, m.getItemWantedErr
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, m.putItemWantedErr
}

func TestToRepoMessage(t *testing.T) {
	tests := []struct {
		arg     *message.Message
		want    *repositoryMessage
		wantErr bool
	}{
		{nil, nil, true},
		{
			&message.Message{MessageHeader: message.MessageHeader{ID: message.MustUUID("ab0f8186-4b68-430e-a07e-b517300e6f9f")}},
			&repositoryMessage{MessageID: "ab0f8186-4b68-430e-a07e-b517300e6f9f"},
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
