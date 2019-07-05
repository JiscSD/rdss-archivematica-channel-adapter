package adapter

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Storage interface {
	AssociateResearchObject(ctx context.Context, objectUUID string, transferID string) error
	GetResearchObject(ctx context.Context, objectUUID string) (string, error)
}

type storageDynamoDBImpl struct {
	DynamoDB dynamodbiface.DynamoDBAPI
	Table    string
}

var _ Storage = (*storageDynamoDBImpl)(nil)

func NewStorageDynamoDB(client dynamodbiface.DynamoDBAPI, table string) *storageDynamoDBImpl {
	return &storageDynamoDBImpl{
		DynamoDB: client,
		Table:    table,
	}
}

type storageItem struct {
	ObjectUUID string `dynamodbav:"objectUUID"`
	TransferID string `dynamodbav:"transferID"`
}

func (s *storageDynamoDBImpl) AssociateResearchObject(ctx context.Context, objectUUID string, transferID string) error {
	si := &storageItem{
		ObjectUUID: objectUUID,
		TransferID: transferID,
	}
	item, err := dynamodbattribute.MarshalMap(si)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.Table),
		Item:      item,
	}
	_, err = s.DynamoDB.PutItemWithContext(ctx, input)
	return err
}

func (s *storageDynamoDBImpl) GetResearchObject(ctx context.Context, objectUUID string) (string, error) {
	var input = &dynamodb.GetItemInput{
		TableName: aws.String(s.Table),
		Key: map[string]*dynamodb.AttributeValue{
			"objectUUID": {S: aws.String(objectUUID)},
		},
	}
	output, err := s.DynamoDB.GetItemWithContext(ctx, input)
	if err != nil || output.Item == nil {
		return "", fmt.Errorf("not found")
	}
	si := &storageItem{}
	if err := dynamodbattribute.UnmarshalMap(output.Item, si); err != nil {
		return "", err
	}
	return si.TransferID, nil
}
