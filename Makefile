VERSION := $(shell git describe --tags --always --dirty)

default: testrace vet fmt

tools:
	# See also tools.go
	go install github.com/johnewart/io-bindata
	go install golang.org/x/tools/cmd/cover

build:
	@echo ${VERSION}
	@env CGO_ENABLED=0 go build -ldflags "-X github.com/JiscRDSS/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}" -a -o rdss-archivematica-channel-adapter

install:
	@echo ${VERSION}
	@env CGO_ENABLED=0 go install -ldflags "-X github.com/JiscRDSS/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}"

test:
	@go test ./...

testrace:
	@go test -race ./...

vet:
	@go vet ./...

fmt:
	@test -z "$(shell gofmt -l -d -e . | tee /dev/stderr)"

cover:
	@hack/coverage.sh

spec:
	hack/build-spec.sh

bench:
	@for pkg in $(shell go list ./...); do \
		go test -bench=. $$pkg; \
	done

release:
	goreleaser --rm-dist

release-test:
	goreleaser --snapshot --skip-publish --rm-dist

.NOTPARALLEL:

.PHONY: default tools build test testrace cover proto bench spec
