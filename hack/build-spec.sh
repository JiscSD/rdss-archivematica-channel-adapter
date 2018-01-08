#!/usr/bin/env bash

# This script is used to generate Go code from the spec files.

set -o errexit
set -o pipefail
set -o nounset

if ! which go-bindata > /dev/null 2>&1 ; then
	echo >&2 "Cannot find go-bindata. Install with \"go get -u go get -u github.com/jteeuwen/go-bindata/...\""
	echo >&2 "Aborting.";
	exit 1;
fi

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly __root="$(cd "$(dirname "${__dir}")" && pwd)"
readonly __gopath="$(cd "$(dirname "${__root}/../../../")" && pwd)"

echo "Compiling..."
cd ${__root}

go-bindata \
	-o "./broker/message/specdata/specdata.go" \
	-nometadata \
	-pkg "specdata" \
	-prefix "./spec" \
		"./spec/schemas/..." \
		"./spec/messages/..."

go fmt ./broker/message/specdata/...
