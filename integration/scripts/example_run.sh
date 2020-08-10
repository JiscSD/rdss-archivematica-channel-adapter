#!/usr/bin/env bash

set -e

hash rdss-archivematica-channel-adapter 2>/dev/null || (cd ../.. && make install)

env \
	RDSS_ARCHIVEMATICA_ADAPTER_LOGGING.LEVEL="info" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.PROCESSING_TABLE="rdss_archivematica_adapter_processing_state" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REPOSITORY_TABLE="rdss_archivematica_adapter_local_data_repository" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.REGISTRY_TABLE="rdss_archivematica_adapter_registry" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_RECV_MAIN_ADDR="http://localhost:4566/queue/main" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_MAIN_ADDR="arn:aws:sns:us-east-1:000000000000:main" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_ERROR_ADDR="arn:aws:sns:us-east-1:000000000000:error" \
	RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_INVALID_ADDR="arn:aws:sns:us-east-1:000000000000:invalid" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_ENDPOINT="http://localhost:4566" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_ENDPOINT="http://localhost:4566" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_ENDPOINT="http://localhost:4566" \
	RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_ENDPOINT="http://localhost:4566" \
	AWS_REGION="us-east-1" \
	AWS_ACCESS_KEY="1234" \
	AWS_SECRET_KEY="5678" \
		rdss-archivematica-channel-adapter server
