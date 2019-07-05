# Contributing

## Building from source

This section describes how to build the application from source.

### Prerequisites

Install [Go 1.12][1] or newer.

### Fetch the source

We use [`Go modules`][2] for dependency management.

Clone the repository:

    git clone https://github.com/JiscRDSS/rdss-archivematica-channel-adapter

### Building

To build the application, run:

    go install

The binary `rdss-archivematica-channel-adapter` should be in your `$GOPATH/bin`
directory, likely `/home/user/go/bin/rdss-archivematica-channel-adapter`.

### Run the tests

Simply run:

    go test ./...

The root `Makefile` includes other options.

[1]: https://golang.org/doc/install
[2]: https://github.com/golang/go/wiki/Modules
