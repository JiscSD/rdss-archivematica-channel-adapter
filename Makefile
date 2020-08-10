VERSION := $(shell git describe --tags --always --dirty)
SCHEMA_SERVICE_ADDR := "https://messageschema.dev.jiscrepository.com"
GOCMD = go
GOBUILD = $(GOCMD) build
GOINSTALL = $(GOCMD) install
GOTEST = $(GOCMD) test

default: testrace fmt lint

tools:
	# Install tools listed in tools.go.
	$(GOINSTALL) golang.org/x/tools/cmd/cover

	# Install golangci-lint (using sudo in Travis CI).
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b $(shell go env GOPATH)/bin v1.30.0

build:
	@echo ${VERSION}
	@env CGO_ENABLED=0 $(GOBUILD) -ldflags "-X github.com/JiscSD/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}" -a -o rdss-archivematica-channel-adapter

install:
	@echo ${VERSION}
	@env CGO_ENABLED=0 $(GOINSTALL) -ldflags "-X github.com/JiscSD/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}"

test:
	@$(GOTEST) -short ./...

testrace:
	@$(GOTEST) -short -race ./...

test-integration: install
	$(GOTEST) -v ./integration/... -valsvc=$(SCHEMA_SERVICE_ADDR)

fmt:
	@test -z "$(shell gofmt -l -d -e . | tee /dev/stderr)"

lint:
	@golangci-lint run

cover:
	@hack/coverage.sh

bench:
	@for pkg in $(shell go list ./...); do \
		$(GOTEST) -bench=. $$pkg; \
	done

release:
	goreleaser --rm-dist

release-test:
	goreleaser --snapshot --skip-publish --rm-dist

.NOTPARALLEL:

.PHONY: default tools build install test testrace fmt lint cover spec bench release release-test
