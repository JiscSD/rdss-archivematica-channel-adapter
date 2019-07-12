#!/usr/bin/env bash

set -e

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function aws() {
	env \
		AWS_DEFAULT_REGION="us-east-1" \
		AWS_ACCESS_KEY_ID="1234" \
		AWS_SECRET_ACCESS_KEY="5678" \
			aws --endpoint-url="http://localhost:${1}" ${*:2}
}

aws 4569 dynamodb put-item \
    --table-name="rdss_archivematica_adapter_registry" \
    --item "file://${__dir}/../testdata/registry-item-tenant-1.json"

aws 4569 dynamodb put-item \
    --table-name="rdss_archivematica_adapter_registry" \
    --item "file://${__dir}/../testdata/registry-item-tenant-2.json"
