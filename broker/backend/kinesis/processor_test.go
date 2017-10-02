// +build !race

// Race detector disabled explicitly because MockDynamo is not safe.

package kinesis

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/sirupsen/logrus"
	"github.com/twitchscience/kinsumer/mocks"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
)

func Test_processor_route_ErrorHandling(t *testing.T) {
	tc := []struct {
		handlers []backend.Handler
		expErr   bool
	}{
		{
			handlers: []backend.Handler{
				func([]byte) error { return nil },
				func([]byte) error { return nil },
			},
			expErr: false,
		},
		{
			handlers: []backend.Handler{
				func([]byte) error { return errors.New("something really bad happened") },
				func([]byte) error { return nil },
			},
			expErr: true,
		},
	}
	backend := getBackend(t)
	for _, tt := range tc {
		p, err := newProcessor(backend, "stream")
		if err != nil {
			t.Fatal(err)
		}
		p.handlers = tt.handlers
		err = p.route([]byte("{}"))
		if tt.expErr && err == nil {
			t.Fatalf("processor.route() returned unexpected nil error value")
		}
	}
}

func getBackend(t *testing.T) *BackendImpl {
	// Create backend with non-default app name and table names
	backend := &BackendImpl{
		logger:  logrus.New(),
		Kinesis: myKinesis{},
		appName: "test_app_0",
		DynamoDB: myDynamo{mocks.NewMockDynamo([]string{
			"test_app_0_clients",
			"test_app_0_checkpoints",
			"test_app_0_metadata",
		})},
	}
	return backend
}

type myKinesis struct {
	kinesisiface.KinesisAPI
}

func (d myKinesis) DescribeStream(input *kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error) {
	output := &kinesis.DescribeStreamOutput{}
	output.StreamDescription = &kinesis.StreamDescription{StreamStatus: aws.String("ACTIVE")}
	return output, nil
}

func (d myKinesis) DescribeStreamPages(*kinesis.DescribeStreamInput, func(*kinesis.DescribeStreamOutput, bool) bool) error {
	return nil
}

type myDynamo struct {
	dynamodbiface.DynamoDBAPI
}

func (d myDynamo) DescribeTable(input *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
	output := &dynamodb.DescribeTableOutput{}
	output.Table = &dynamodb.TableDescription{TableStatus: aws.String("ACTIVE")}
	return output, nil
}

func (d myDynamo) DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	output := &dynamodb.DeleteItemOutput{}
	return output, nil
}
