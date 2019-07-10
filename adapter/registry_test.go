package adapter

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type dynamock struct {
	mock.Mock
	dynamodbiface.DynamoDBAPI
}

func (m *dynamock) ScanWithContext(ctx aws.Context, input *dynamodb.ScanInput, opts ...request.Option) (*dynamodb.ScanOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.ScanOutput), args.Error(1)
}

func TestRegistry(t *testing.T) {
	m := &dynamock{}
	reloadFrequency = 10 * time.Millisecond

	m.On(
		"ScanWithContext",
		mock.AnythingOfType("*context.cancelCtx"),
		&dynamodb.ScanInput{
			ConsistentRead: aws.Bool(true),
			TableName:      aws.String("mockTable"),
		},
	).Return(&dynamodb.ScanOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			map[string]*dynamodb.AttributeValue{
				"tenantJiscID": &dynamodb.AttributeValue{S: aws.String("1")},
				"url":          &dynamodb.AttributeValue{S: aws.String("http://192.168.1.1")},
				"user":         &dynamodb.AttributeValue{S: aws.String("test1")},
				"key":          &dynamodb.AttributeValue{S: aws.String("test1")},
				"transferDir":  &dynamodb.AttributeValue{S: aws.String("/home/jisc/tenant1")},
			},
			map[string]*dynamodb.AttributeValue{
				"tenantJiscID": &dynamodb.AttributeValue{S: aws.String("2")},
				"url":          &dynamodb.AttributeValue{S: aws.String("http://192.168.1.2")},
				"user":         &dynamodb.AttributeValue{S: aws.String("test2")},
				"key":          &dynamodb.AttributeValue{S: aws.String("test2")},
				"transferDir":  &dynamodb.AttributeValue{S: aws.String("/home/jisc/tenant2")},
			},
		},
	}, nil)

	r, err := NewRegistry(logrus.StandardLogger(), m, "mockTable")
	r.Reload()
	defer r.Stop()
	m.AssertNumberOfCalls(t, "ScanWithContext", 1)

	c := r.Get(1)
	assert.NotNil(t, c)
	assert.Equal(t, c.BaseURL.Hostname(), "192.168.1.1")
	assert.Equal(t, c.User, "test1")
	assert.Equal(t, c.Key, "test1")
	assert.NoError(t, err)

	c = r.Get(2)
	assert.NotNil(t, c)
	assert.Equal(t, c.BaseURL.Hostname(), "192.168.1.2")
	assert.Equal(t, c.User, "test2")
	assert.Equal(t, c.Key, "test2")
	assert.NoError(t, err)

	assert.Nil(t, r.Get(3))
}
