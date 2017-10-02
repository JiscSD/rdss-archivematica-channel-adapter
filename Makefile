PKGS?=$$(go list ./... | grep -v '/vendor/')
FILES?=$$(find . -name '*.go' | grep -v vendor)

VERSION := $(shell git describe --tags --always --dirty)

default: testrace vet

tools:
	go get -u github.com/golang/dep/...
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/golang/protobuf/protoc-gen-go

build:
	@echo ${VERSION}
	@env CGO_ENABLED=0 go build -ldflags "-X github.com/JiscRDSS/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}" -a -o rdss-archivematica-channel-adapter .

install:
	@echo ${VERSION}
	@env CGO_ENABLED=0 go install -ldflags "-X github.com/JiscRDSS/rdss-archivematica-channel-adapter/version.VERSION=${VERSION}" $(PKGS)

test:
	@go test -i $(PKGS)
	@go test $(PKGS)

testrace:
	@go test -i -race $(PKGS)
	@go test -race $(PKGS)

vet:
	@go vet $(PKGS); if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

cover:
	@hack/coverage.sh

vendor-status:
	dep status

proto:
	hack/build-proto.sh

bench:
	@go test -v -run=^$ -bench=$(PKGS)

.NOTPARALLEL:

.PHONY: default tools build test testrace cover vendor-status proto bench
