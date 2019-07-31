#!/usr/bin/env bash

set -e

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

sample="${__dir}/../testdata/message-metadata-create.json"
uuid=$(uuidgen)
tmpfile=$(mktemp /tmp/test-send-message.XXXXXX)
title="test-$(date +"%Y%m%d%H%M%S")"

cat ${sample} | jq '.messageBody.objectTitle = "'${title}'" | .messageHeader.messageId = "'${uuid}'"' > ${tmpfile}

env \
	AWS_SECRET_ACCESS_KEY=1234 \
	AWS_ACCESS_KEY_ID=5678 \
	AWS_DEFAULT_REGION=us-east-1 \
		aws --endpoint-url=http://localhost:4576 \
			sqs send-message \
				--delay-seconds 2 \
				--queue-url http://localhost:4576/queue/main \
				--message-body file://${tmpfile}
