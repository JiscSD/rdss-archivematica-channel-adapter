#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

covermode=${COVERMODE:-atomic}
coverprofile=$(mktemp /tmp/coverage.XXXXXXXXXX)

hash goveralls 2>/dev/null || go get github.com/mattn/goveralls

generate_cover_data() {
    go test -short -coverprofile="${coverprofile}" -covermode="${covermode}" $(go list ./...)
}

push_to_coveralls() {
    goveralls -coverprofile="${coverprofile}" -service=travis-ci
}

generate_cover_data
go tool cover -func "${coverprofile}"

case "${1-}" in
    --html)
        go tool cover -html "${coverprofile}"
        ;;
    --coveralls)
        push_to_coveralls
        ;;
esac
