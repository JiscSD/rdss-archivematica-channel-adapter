#!/usr/bin/env bash

env AWS_SECRET_ACCESS_KEY=1234 \
    AWS_ACCESS_KEY_ID=5678 \
    AWS_DEFAULT_REGION=us-east-1 \
	aws --endpoint-url=http://localhost:4576 \
	sqs send-message --queue-url http://127.0.0.1:4576/queue/main --message-body file://message-api-spec/messages/example_message.json
