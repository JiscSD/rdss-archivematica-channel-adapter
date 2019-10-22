VERSION := $(shell git describe --tags --always --dirty)
SCHEMA_SERVICE_ADDR := "https://messageschema.dev.jiscrepository.com"

default: testrace fmt lint

tools:
	# Install tools listed in tools.go.
	go install golang.org/x/tools/cmd/cover

	# Install golangci-lint.
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.21.0

build:
	@echo ${VERSION}
	@env CGO_ENABLED=0 go build -ldflags "-X github.com/JiscSD/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}" -a -o rdss-archivematica-channel-adapter

install:
	@echo ${VERSION}
	@env CGO_ENABLED=0 go install -ldflags "-X github.com/JiscSD/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}"

test:
	@go test -short ./...

testrace:
	@go test -short -race ./...

test-integration: install
	docker-compose --file ./integration/docker-compose.yml up -d --force-recreate
	docker-compose --file ./integration/docker-compose.yml ps
	integration/scripts/wait.sh
	integration/scripts/provision.sh
	go test -v ./integration/... -valsvc=$(SCHEMA_SERVICE_ADDR)

fmt:
	@test -z "$(shell gofmt -l -d -e . | tee /dev/stderr)"

lint:
	@golangci-lint run

cover:
	@hack/coverage.sh

bench:
	@for pkg in $(shell go list ./...); do \
		go test -bench=. $$pkg; \
	done

release:
	goreleaser --rm-dist

release-test:
	goreleaser --snapshot --skip-publish --rm-dist

.NOTPARALLEL:

.PHONY: default tools build install test testrace fmt lint cover spec bench release release-test
