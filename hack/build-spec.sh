#!/usr/bin/env bash

# This script is used to generate Go code from the spec files.

set -o errexit
set -o pipefail
set -o nounset

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly __root="$(cd "$(dirname "${__dir}")" && pwd)"
readonly __gopath="$(cd "$(dirname "${__root}/../../../")" && pwd)"

echo "Compiling..."
cd ${__root}

genfile="./broker/message/specdata/specdata.go"
tmpfile=$(mktemp /tmp/abc-script.XXXXXX)

io-bindata \
	-o ${genfile} \
	-nometadata \
	-pkg "specdata" \
	-prefix "./message-api-spec" \
		"./message-api-spec/schemas/..." \
		"./message-api-spec/messages/..."

go fmt ./broker/message/specdata/...

echo "//lint:file-ignore ST1005 Generated code." | cat - ${genfile} > ${tmpfile} && mv ${tmpfile} ${genfile}
