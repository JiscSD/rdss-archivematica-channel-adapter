#!/usr/bin/env bash

# This script attempts to remove the resources beforehand so it can be executed
# over an existing stack. But if you want to ensure a clean state for testing
# it's probably best to recreate the LocalStack container.

set -e

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function aws() {
	env \
		AWS_DEFAULT_REGION="us-east-1" \
		AWS_ACCESS_KEY_ID="1234" \
		AWS_SECRET_ACCESS_KEY="5678" \
			aws --endpoint-url="http://localhost:${1}" ${*:2}
}

# Create SQS queue
aws 4566 sqs delete-queue \
	--queue-url="http://localhost:4576/queue/main" || true 2>/dev/null
aws 4566 sqs create-queue \
	--queue-name="main" --attributes "VisibilityTimeout=600"

# Create SNS topics
aws 4566 sns delete-topic \
	--topic-arn="arn:aws:sns:us-east-1:000000000000:main" || true 2>/dev/null
aws 4566 sns create-topic \
	--name="main"
aws 4566 sns delete-topic \
	--topic-arn="arn:aws:sns:us-east-1:000000000000:invalid" || true 2>/dev/null
aws 4566 sns create-topic \
	--name="invalid"
aws 4566 sns delete-topic \
	--topic-arn="arn:aws:sns:us-east-1:000000000000:error" || true 2>/dev/null
aws 4566 sns create-topic \
	--name="error"

# Create DynamoDB tables
aws 4566 dynamodb delete-table \
	--table-name="rdss_archivematica_adapter_local_data_repository" || true 2>/dev/null
aws 4566 dynamodb create-table \
	--table-name="rdss_archivematica_adapter_local_data_repository" --attribute-definitions="AttributeName=ID,AttributeType=S" --key-schema="AttributeName=ID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"
aws 4566 dynamodb delete-table \
	--table-name="rdss_archivematica_adapter_processing_state" || true 2>/dev/null
aws 4566 dynamodb create-table \
	--table-name="rdss_archivematica_adapter_processing_state" --attribute-definitions="AttributeName=objectUUID,AttributeType=S" --key-schema="AttributeName=objectUUID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"
aws 4566 dynamodb delete-table \
	--table-name="rdss_archivematica_adapter_registry" || true 2>/dev/null
aws 4566 dynamodb create-table \
	--table-name="rdss_archivematica_adapter_registry" --attribute-definitions="AttributeName=tenantJiscID,AttributeType=S" --key-schema="AttributeName=tenantJiscID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"

# Create S3 bucket
aws 4566 s3 rm \
	s3://bucket --recursive || true 2>/dev/null
aws 4566 s3api create-bucket \
	--bucket="bucket"
