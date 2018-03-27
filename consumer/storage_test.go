package consumer

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

func TestStorageInMemoryImpl(t *testing.T) {
	ctx := context.Background()
	s := NewStorageInMemory()

	id, _ := s.GetResearchObject(ctx, "foo")
	if have, want := id, ""; have != want {
		t.Fatalf("GetResearchObject(); have `%s` want `%s`", have, want)
	}

	_ = s.AssociateResearchObject(ctx, "foo", "bar")

	id, _ = s.GetResearchObject(ctx, "foo")
	if have, want := id, "bar"; have != want {
		t.Fatalf("GetResearchObject(); have `%s` want `%s`", have, want)
	}
}

func TestStorageDynamoDBImpl(t *testing.T) {
	ctx := context.Background()
	dynamock := &mockDynamoDBClient{
		GetItemWantedItem: &storageItem{ObjectUUID: "1", TransferID: "2"},
	}
	s := NewStorageDynamoDB(dynamock, "table")

	id, _ := s.GetResearchObject(ctx, "foo")
	if have, want := *dynamock.GetItemInput.Key["objectUUID"].S, "foo"; have != want {
		t.Fatalf("GetResearchObject(); want %v, have %v", want, have)
	}
	if want, have := dynamock.GetItemWantedItem.(*storageItem).TransferID, id; want != have {
		t.Fatalf("GetResearchObject(); want %v, have %v", want, have)
	}

	_ = s.AssociateResearchObject(ctx, "foo", "bar")
	if have, want := *dynamock.PutItemInput.Item["objectUUID"].S, "foo"; have != want {
		t.Fatalf("GetResearchObject(); want %v, have %v", want, have)
	}
	if have, want := *dynamock.PutItemInput.Item["transferID"].S, "bar"; have != want {
		t.Fatalf("GetResearchObject(); want %v, have %v", want, have)
	}
}

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	GetItemWantedItem interface{}
	GetItemInput      *dynamodb.GetItemInput
	PutItemInput      *dynamodb.PutItemInput
}

func (m *mockDynamoDBClient) GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error) {
	m.GetItemInput = input
	item, err := dynamodbattribute.MarshalMap(m.GetItemWantedItem)
	if err != nil {
		return nil, err
	}
	return &dynamodb.GetItemOutput{Item: item}, nil
}

func (m *mockDynamoDBClient) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	m.PutItemInput = input
	return &dynamodb.PutItemOutput{}, nil
}
