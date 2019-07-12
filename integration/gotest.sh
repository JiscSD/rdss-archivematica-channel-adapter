#!/usr/bin/env bash

set -e

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd ${__dir}
go test -tags=integration -v .
