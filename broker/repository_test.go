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

func TestNewRepository(t *testing.T) {
	testCases := []struct {
		name                 string
		backend              string
		expectedRetRepoIsNil bool
		expectedRetErr       error
	}{
		{"Unsupported backend", "unknown", true, ErrInvalidRepositoryConfig},
		{"Supported backend dynamodb", "dynamodb", false, nil},
		{"Supported backend builtin", "builtin", false, nil},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := NewRepository(&RepositoryConfig{Backend: tc.backend})
			if (tc.expectedRetRepoIsNil && repo != nil) || (!tc.expectedRetRepoIsNil && repo == nil) {
				t.Error("unexpected repo returned")
			}
			if err != tc.expectedRetErr {
				t.Error("unexpected error returned")
			}
		})
	}
}

func TestMustRepository(t *testing.T) {
	testCases := []struct {
		name    string
		backend string
		panic   bool
	}{
		{"Unsupported backend", "unknown", true},
		{"Supported backend", "dynamodb", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tc.panic && r == nil {
					t.Error("It did not panic")
				}
				if !tc.panic && r != nil {
					t.Error("It did panic")
				}
			}()
			MustRepository(NewRepository(&RepositoryConfig{Backend: tc.backend}))
		})
	}
}

func Test_getDynamoDBInstance(t *testing.T) {
	config := &RepositoryConfig{DynamoDBRegion: "region", DynamoDBEndpoint: "endpoint", DynamoDBTLS: true}
	dbiface := getDynamoDBInstance(config)
	db, _ := dbiface.(*dynamodb.DynamoDB)

	if want, got := config.DynamoDBRegion, *db.Config.Region; want != got {
		t.Errorf("getDynamoDBInstance(); want = %v, got = %v", want, got)
	}
	if want, got := config.DynamoDBEndpoint, *db.Config.Endpoint; want != got {
		t.Errorf("getDynamoDBInstance(); want = %v, got = %v", want, got)
	}
	if want, got := config.DynamoDBTLS, !*db.Config.DisableSSL; want != got {
		t.Errorf("getDynamoDBInstance(); want = %v, got = %v", want, got)
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
			&RepositoryMessage{MessageId: "ab0f8186-4b68-430e-a07e-b517300e6f9f"},
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
func TestRepositoryBuiltinImpl_Get(t *testing.T) {
	tests := []struct {
		msgs map[string]*RepositoryMessage
		key  string
		want *RepositoryMessage
	}{
		{nil, "foobar", nil},
		{map[string]*RepositoryMessage{}, "foobar", nil},
		{
			map[string]*RepositoryMessage{"this": &RepositoryMessage{MessageId: "ID"}},
			"this",
			&RepositoryMessage{MessageId: "ID"},
		},
		{
			map[string]*RepositoryMessage{
				"uno": &RepositoryMessage{MessageId: "1"},
				"dos": &RepositoryMessage{MessageId: "2"},
			},
			"dos",
			&RepositoryMessage{MessageId: "2"},
		},
	}
	for _, tt := range tests {
		repo := &RepositoryBuiltinImpl{msgs: tt.msgs}
		ret := repo.Get(tt.key)
		if !reflect.DeepEqual(tt.want, ret) {
			t.Error()
		}
	}
}

func TestRepositoryBuiltinImpl_Put(t *testing.T) {
	tests := []struct {
		key       string
		msg       *message.Message
		want      string
		found     bool
		errWanted bool
	}{
		{
			"ab0f8186-4b68-430e-a07e-b517300e6f9f",
			&message.Message{MessageHeader: message.MessageHeader{ID: message.MustUUID("ab0f8186-4b68-430e-a07e-b517300e6f9f")}},
			"ab0f8186-4b68-430e-a07e-b517300e6f9f",
			true,
			false,
		},
		{
			"my-id",
			&message.Message{MessageHeader: message.MessageHeader{ID: message.MustUUID("ab0f8186-4b68-430e-a07e-b517300e6f9f")}},
			"ab0f8186-4b68-430e-a07e-b517300e6f9f",
			false,
			false,
		},
		{
			"mi-id",
			nil,
			"12345",
			false,
			true,
		},
	}
	for _, tt := range tests {
		repo := &RepositoryBuiltinImpl{msgs: make(map[string]*RepositoryMessage)}
		err := repo.Put(tt.msg)
		if tt.errWanted && err == nil {
			t.Fatal("error expected but returned nil", err)
		} else if !tt.errWanted && err != nil {
			t.Fatal("error not expected but one was returned", err)
		}
		msg, ok := repo.msgs[tt.key]
		if tt.found && !ok {
			t.Fatalf("expected value with key %v wasn't found", tt.key)
		} else if !tt.found && ok {
			t.Fatalf("unexpected value with key %v found", tt.key)
		}
		if ok && msg.MessageId != tt.want {
			t.Fatalf("MessageId has an unexpected value: %v; expected: %v", msg.MessageId, tt.want)
		}
	}
}

func TestRepositoryDynamoDBImpl_Get(t *testing.T) {
	repo, _ := NewRepository(&RepositoryConfig{Backend: "dynamodb"})
	r, _ := repo.(*RepositoryDynamoDBImpl)
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
			r.DynamoDB = tt.client
			if got := r.Get("foo"); !reflect.DeepEqual(tt.want, got) {
				t.Errorf("RepositoryDynamoDBImpl.Get(); want %v, got %v", tt.want, got)
			}
		})
	}
}
func TestRepositoryDynamoDBImpl_Put(t *testing.T) {
	repo, _ := NewRepository(&RepositoryConfig{Backend: "dynamodb"})
	r, _ := repo.(*RepositoryDynamoDBImpl)
	tests := []struct {
		arg     *message.Message
		wantErr bool
		client  *mockDynamoDBClient
	}{
		{
			message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand),
			false,
			&mockDynamoDBClient{},
		},
		{
			message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand),
			true,
			&mockDynamoDBClient{PutItem_WantedErr: errors.New("error")},
		},
		{nil, true, nil},
		// TODO: check case with dynamodbattribute.MarshalMap(rMsg) returning error
	}
	for _, tt := range tests {
		r.DynamoDB = tt.client
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
