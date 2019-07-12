#!/usr/bin/env bash

set -e

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly __root="$(cd "$(dirname "${__dir}")" && pwd)"

function aws() {
	env \
		AWS_DEFAULT_REGION="us-east-1" \
		AWS_ACCESS_KEY_ID="1234" \
		AWS_SECRET_ACCESS_KEY="5678" \
			aws --endpoint-url="http://127.0.0.1:${1}" ${*:2}
}

function build() {
	pushd ${__root} 1>/dev/null
		go build -o /tmp/rdss-archivematica-channel-adapter
	popd 1>/dev/null
}

function provision() {
	# Create SQS queue
	aws 4576 sqs delete-queue --queue-url="http://127.0.0.1:4576/queue/main" || true 2>/dev/null
	aws 4576 sqs create-queue --queue-name="main" --attributes VisibilityTimeout=600

	# Create SNS topics
	aws 4575 sns delete-topic --topic-arn="arn:aws:sns:us-east-1:123456789012:main" || true 2>/dev/null
	aws 4575 sns create-topic --name="main"
	aws 4575 sns delete-topic --topic-arn="arn:aws:sns:us-east-1:123456789012:invalid" || true 2>/dev/null
	aws 4575 sns create-topic --name="invalid"
	aws 4575 sns delete-topic --topic-arn="arn:aws:sns:us-east-1:123456789012:error" || true 2>/dev/null
	aws 4575 sns create-topic --name="error"

	# Create DynamoDB tables
	aws 4569 dynamodb delete-table --table-name="rdss_archivematica_adapter_local_data_repository" || true 2>/dev/null
	aws 4569 dynamodb create-table --table-name="rdss_archivematica_adapter_local_data_repository" --attribute-definitions="AttributeName=ID,AttributeType=S" --key-schema="AttributeName=ID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"
	aws 4569 dynamodb delete-table --table-name="rdss_archivematica_adapter_processing_state" || true 2>/dev/null
	aws 4569 dynamodb create-table --table-name="rdss_archivematica_adapter_processing_state" --attribute-definitions="AttributeName=objectUUID,AttributeType=S" --key-schema="AttributeName=objectUUID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"
	aws 4569 dynamodb delete-table --table-name="rdss_archivematica_adapter_registry" || true 2>/dev/null
	aws 4569 dynamodb create-table --table-name="rdss_archivematica_adapter_registry" --attribute-definitions="AttributeName=tenantJiscID,AttributeType=S" --key-schema="AttributeName=tenantJiscID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"

	# Create entry in DynamoDB-based registry
	aws 4569 dynamodb put-item --table-name="rdss_archivematica_adapter_registry" --item "file://${__dir}/testdata/registry-item-tenant-1.json"
	aws 4569 dynamodb put-item --table-name="rdss_archivematica_adapter_registry" --item "file://${__dir}/testdata/registry-item-tenant-2.json"

	# Create S3 bucket
	aws 4572 s3api delete-bucket --bucket="bucket" || true 2>/dev/null
	aws 4572 s3api create-bucket --bucket="bucket"
}

function flush() {
	aws 4576 sqs delete-queue --queue-url="http://127.0.0.1:4576/queue/main" || true 2>/dev/null
	aws 4576 sqs create-queue --queue-name="main" --attributes VisibilityTimeout=600
	aws 4569 dynamodb delete-table --table-name="rdss_archivematica_adapter_local_data_repository" || true 2>/dev/null
	aws 4569 dynamodb create-table --table-name="rdss_archivematica_adapter_local_data_repository" --attribute-definitions="AttributeName=ID,AttributeType=S" --key-schema="AttributeName=ID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"
	aws 4569 dynamodb delete-table --table-name="rdss_archivematica_adapter_processing_state" || true 2>/dev/null
	aws 4569 dynamodb create-table --table-name="rdss_archivematica_adapter_processing_state" --attribute-definitions="AttributeName=objectUUID,AttributeType=S" --key-schema="AttributeName=objectUUID,KeyType=HASH" --billing-mode="PAY_PER_REQUEST"
}

# Parse command-line arguments
should_provision=false
should_flush=false
should_build=true
for i in "$@"; do
case $i in
    -p|--provision)
    should_provision=true
    shift
    ;;
	-f|--flush)
	should_flush=true
	shift
	;;
    --no-build)
    should_build=false
    shift
    ;;
esac
done


if [ "$should_provision" = "true" ]; then
	provision
fi

if [ "$should_flush" = "true" ]; then
	flush
fi

if [ "$should_build" = "true" ]; then
	build
fi

env \
	RDSS_ARCHIVEMATICA_ADAPTER_LOGGING.LEVEL="debug" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.PROCESSING_TABLE="rdss_archivematica_adapter_processing_state" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REPOSITORY_TABLE="rdss_archivematica_adapter_local_data_repository" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REGISTRY_TABLE="rdss_archivematica_adapter_registry" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.VALIDATION_MODE="strict" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_RECV_MAIN_ADDR="http://127.0.0.1:4576/queue/main" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_MAIN_ADDR="arn:aws:sns:us-east-1:123456789012:main" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_ERROR_ADDR="arn:aws:sns:us-east-1:123456789012:error" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_INVALID_ADDR="arn:aws:sns:us-east-1:123456789012:invalid" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_ENDPOINT="http://127.0.0.1:4572" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_ENDPOINT="http://127.0.0.1:4569" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_ENDPOINT="http://127.0.0.1:4576" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_ENDPOINT="http://127.0.0.1:4575" \
	AWS_REGION="us-east-1" \
	AWS_ACCESS_KEY="1234" \
	AWS_SECRET_KEY="5678" \
		/tmp/rdss-archivematica-channel-adapter server
