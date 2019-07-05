#!/usr/bin/env bash

set -e

function aws() {
	env \
		AWS_DEFAULT_REGION="us-east-1" \
		AWS_ACCESS_KEY_ID="1234" \
		AWS_SECRET_ACCESS_KEY="5678" \
			aws --endpoint-url="http://127.0.0.1:${1}" ${*:2}
}

function provision() {
	# Build binary
	go build -o /tmp/rdss-archivematica-channel-adapter

	# Create SQS queue
	aws 4576 sqs delete-queue --queue-url="http://127.0.0.1:4576/queue/main" || true 2>/dev/null
	aws 4576 sqs create-queue --queue-name="main"

	# Create SNS topics
	aws 4575 sns delete-topic --topic-arn="arn:aws:sns:us-east-1:123456789012:main" || true 2>/dev/null
	aws 4575 sns create-topic --name="main"
	aws 4575 sns delete-topic --topic-arn="arn:aws:sns:us-east-1:123456789012:invalid" || true 2>/dev/null
	aws 4575 sns create-topic --name="invalid"
	aws 4575 sns delete-topic --topic-arn="arn:aws:sns:us-east-1:123456789012:error" || true 2>/dev/null
	aws 4575 sns create-topic --name="error"

	# Create DynamoDB tables
	aws 4569 dynamodb delete-table --table-name="rdss_archivematica_adapter_local_data_repository" || true 2>/dev/null
	aws 4569 dynamodb create-table --table-name="rdss_archivematica_adapter_local_data_repository" --attribute-definitions="AttributeName=ID,AttributeType=S" --key-schema="AttributeName=ID,KeyType=HASH" --provisioned-throughput="ReadCapacityUnits=10,WriteCapacityUnits=10"
	aws 4569 dynamodb delete-table --table-name="rdss_archivematica_adapter_processing_state" || true 2>/dev/null
	aws 4569 dynamodb create-table --table-name="rdss_archivematica_adapter_processing_state" --attribute-definitions="AttributeName=objectUUID,AttributeType=S" --key-schema="AttributeName=objectUUID,KeyType=HASH" --provisioned-throughput="ReadCapacityUnits=10,WriteCapacityUnits=10"

	# Create S3 bucket
	aws 4572 s3api delete-bucket --bucket="bucket" || true 2>/dev/null
	aws 4572 s3api create-bucket --bucket="bucket"
}

provision

env \
	RDSS_ARCHIVEMATICA_ADAPTER_LOGGING.LEVEL="debug" \
	RDSS_ARCHIVEMATICA_ADAPTER_AMCLIENT.URL="http://127.0.0.1:62080" \
	RDSS_ARCHIVEMATICA_ADAPTER_AMCLIENT.USER="test" \
	RDSS_ARCHIVEMATICA_ADAPTER_AMCLIENT.KEY="test" \
	RDSS_ARCHIVEMATICA_ADAPTER_AMCLIENT.TRANSFER_DIR="/tmp" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.PROCESSING_TABLE="rdss_archivematica_adapter_processing_state" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REPOSITORY_TABLE="rdss_archivematica_adapter_local_data_repository" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.VALIDATION_MODE="strict" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_RECV_MAIN_ADDR="http://127.0.0.1:4576/queue/main" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_MAIN_ADDR="arn:aws:sns:us-east-1:123456789012:main" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_ERROR_ADDR="arn:aws:sns:us-east-1:123456789012:error" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_INVALID_ADDR="arn:aws:sns:us-east-1:123456789012:invalid" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_PROFILE="" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_ENDPOINT="http://127.0.0.1:4572" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_PROFILE="" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_ENDPOINT="http://127.0.0.1:4569" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_PROFILE="" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_ENDPOINT="http://127.0.0.1:4576" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_PROFILE="" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_ENDPOINT="http://127.0.0.1:4575" \
	AWS_SDK_LOAD_CONFIG="1" \
	AWS_DEFAULT_REGION="us-east-1" \
	AWS_ACCESS_KEY="1234" \
	AWS_SECRET_KEY="5678" \
		/tmp/rdss-archivematica-channel-adapter server
