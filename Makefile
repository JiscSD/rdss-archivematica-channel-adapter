PKGS?=$$(go list ./... | grep -v '/vendor/')
FILES?=$$(find . -name '*.go' | grep -v vendor)

VERSION := $(shell git describe --tags --always --dirty)

default: testrace vet fmtcheck

tools:
	go get -u github.com/golang/dep/cmd/...
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/gogo/protobuf/proto
	go get -u github.com/gogo/protobuf/jsonpb
	go get -u github.com/gogo/protobuf/protoc-gen-gogo
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u github.com/jteeuwen/go-bindata/...

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

fmt:
	gofmt -w $(FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/hack/gofmtcheck.sh'"

cover:
	@hack/coverage.sh

vendor-status:
	dep status

proto:
	hack/build-proto.sh

spec:
	hack/build-spec.sh

bench:
	@go test -v -run=^$ -bench=$(PKGS)

.NOTPARALLEL:

.PHONY: default tools build test testrace cover vendor-status proto bench spec
