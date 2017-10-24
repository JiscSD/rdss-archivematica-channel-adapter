#!/usr/bin/env bash

# This script is used to compile the protocol buffers.

set -o errexit
set -o pipefail
set -o nounset

if ! which godepgraph > /dev/null 2>&1 ; then
	echo >&2 "Cannot find godepgraph. Install with \"go get -u github.com/kisielk/godepgraph\""
	echo >&2 "See https://github.com/kisielk/godepgraph for more instructions.";
	echo >&2 "Aborting.";
	exit 1;
fi

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

godepgraph -s -p github.com/JiscRDSS/rdss-archivematica-channel-adapter/vendor github.com/JiscRDSS/rdss-archivematica-channel-adapter | dot -Tpng -o ${__dir}/deps-01-simple.png
godepgraph -horizontal github.com/JiscRDSS/rdss-archivematica-channel-adapter | dot -Tpng -o ${__dir}/deps-02-full.png
