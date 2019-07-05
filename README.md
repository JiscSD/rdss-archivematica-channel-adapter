[![Travis CI](https://travis-ci.org/JiscRDSS/rdss-archivematica-channel-adapter.svg?branch=master)](https://travis-ci.org/JiscRDSS/rdss-archivematica-channel-adapter) [![GoDoc](https://godoc.org/github.com/JiscRDSS/rdss-archivematica-channel-adapter?status.svg)](https://godoc.org/github.com/JiscRDSS/rdss-archivematica-channel-adapter) [![Coverage Status](https://coveralls.io/repos/github/JiscRDSS/rdss-archivematica-channel-adapter/badge.svg?branch=master)](https://coveralls.io/github/JiscRDSS/rdss-archivematica-channel-adapter?branch=master) [![Go Report Card](https://goreportcard.com/badge/JiscRDSS/rdss-archivematica-channel-adapter)](https://goreportcard.com/report/JiscRDSS/rdss-archivematica-channel-adapter) [![Sourcegraph](https://sourcegraph.com/github.com/JiscRDSS/rdss-archivematica-channel-adapter/-/badge.svg)](https://sourcegraph.com/github.com/JiscRDSS/rdss-archivematica-channel-adapter?badge)

# RDSS Archivematica Channel Adapter

- [Introduction](#Introduction)
- [Installation](#Installation)
- [Configuration](#Configuration)
  - [Configuration file](#Configuration-file)
  - [Environment variables](#Environment-variables)
  - [AWS service client configuration](#AWS-service-client-configuration)
- [Metrics and runtime profiling data](#Metrics-and-runtime-profiling-data)
- [Contributing](#Contributing)

## Introduction

RDSS Archivematica Channel Adapter is an implementation of a channel adapter for [Archivematica](https://archivematica.org) following the [RDSS messaging API specification](https://github.com/JiscRDSS/rdss-message-api-specification).

## Installation

We're not releasing binaries yet but you can build a Docker image as follows:

    $ docker build --tag rdss-archivematica-channel-adapter .

Now you can run the adapter inside a container:

    $ docker run --tty --rm rdss-archivematica-channel-adapter

Typically, you will use the `server` subcommand and pass some configuration attributes via the environment, e.g.:

    $ docker run \
        --tty --rm \
        --env "RDSS_ARCHIVEMATICA_ADAPTER_LOGGING.LEVEL=WARNING" \
        rdss-archivematica-channel-adapter \
            server

## Configuration

Configuration defaults are included in the source code ([config.go](./app/config.go)). Use it as a reference since it lists all the attributes available including descriptions. We use the [TOML configuration file format](https://en.wikipedia.org/wiki/TOML).

Inject custom configuration attributes via a configuration file and/or environment variables.

### Configuration file

The configuration file can be indicated via the `--config` command-line argument. When undefined, the application attempts to read from one of the following locations:

- `$HOME/.rdss-archivematica-channel-adapter.toml`
- `/etc/archivematica/rdss-archivematica-channel-adapter.toml`

### Environment variables

Configuration from environment variables have precedence over file-based configuration. All environment variables follow the same naming scheme: `RDSS_ARCHIVEMATICA_ADAPTER_<SECTION>_<ATTRIBUTE>=<VALUE>`. Some valid examples are:

- `RDSS_ARCHIVEMATICA_ADAPTER_LOGGING.LEVEL=DEBUG`<br />
  (section: `LOGGING`, attribute: `LEVEL`, value: `DEBUG`)
- `RDSS_ARCHIVEMATICA_ADAPTER_MESSAGE.VALIDATION=FALSE`<br />
  (section: `MESSAGE`, attribute: `VALIDATION`, value: `FALSE`)

### AWS service client configuration

The AWS service client configuration rely on the [shared configuration functionality](https://docs.aws.amazon.com/sdk-for-go/api/aws/session/) which is similar to the [AWS CLI configuration system](https://docs.aws.amazon.com/cli/latest/topic/config-vars.html).

Additionally, you can override the name of the configuration profile used on each service client using the following attributes:

- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_PROFILE`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_PROFILE`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_PROFILE`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_PROFILE`

This can be useful under a variety of scenarios:

- Deployment of alternative services like LocalStack, Minio, etc...
- Applying different credentials, e.g. assuming a IAM role in the SQS/SNS clients.

## Metrics and runtime profiling data

`rdss-archivematica-channel-adapter server` runs a HTTP server that listens on `0.0.0.0:6060` with two purposes:

* `/metrics` serves metrics of the Go runtime and the application meant to be scraped by a Prometheus server.
* `/debug/pprof` serves runtime profiling data in the format expected by the pprof visualization tool. Visit [net/http/pprof docs](https:/golang.org/pkg/net/http/pprof/) for more.

## Contributing

* See [CONTRIBUTING.md][1] for information about setting up your environment and the workflow that we expect.
* Check out the [open issues][2].

[1]: /CONTRIBUTING.md
[2]: https://github.com/JiscRDSS/rdss-archivematica-channel-adapter/issues
