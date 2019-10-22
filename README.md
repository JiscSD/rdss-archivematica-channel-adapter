[![Travis CI](https://travis-ci.org/JiscSD/rdss-archivematica-channel-adapter.svg?branch=master)](https://travis-ci.org/JiscSD/rdss-archivematica-channel-adapter) [![GoDoc](https://godoc.org/github.com/JiscSD/rdss-archivematica-channel-adapter?status.svg)](https://godoc.org/github.com/JiscSD/rdss-archivematica-channel-adapter) [![Coverage Status](https://coveralls.io/repos/github/JiscSD/rdss-archivematica-channel-adapter/badge.svg?branch=master)](https://coveralls.io/github/JiscSD/rdss-archivematica-channel-adapter?branch=master) [![Go Report Card](https://goreportcard.com/badge/JiscSD/rdss-archivematica-channel-adapter)](https://goreportcard.com/report/JiscSD/rdss-archivematica-channel-adapter) [![Sourcegraph](https://sourcegraph.com/github.com/JiscSD/rdss-archivematica-channel-adapter/-/badge.svg)](https://sourcegraph.com/github.com/JiscSD/rdss-archivematica-channel-adapter?badge)

# RDSS Archivematica Channel Adapter

- [Introduction](#introduction)
- [Installation](#installation)
- [Configuration](#configuration)
  - [Configuration file](#configuration-file)
  - [Environment variables](#environment-variables)
  - [Service dependencies](#service-dependencies)
  - [AWS service client configuration](#aws-service-client-configuration)
  - [Registry of Archivematica pipelines](#registry-of-archivematica-pipelines)
- [Metrics and runtime profiling data](#metrics-and-runtime-profiling-data)
- [Contributing](#contributing)

## Introduction

RDSS Archivematica Channel Adapter is an implementation of a channel adapter for [Archivematica](https://archivematica.org) following the [RDSS messaging API specification](https://github.com/JiscSD/rdss-message-api-specification).

## Installation

This application is distributed as a single static binary file that you can download from the [Releases](https://github.com/JiscSD/rdss-archivematica-channel-adapter/releases) page. You can use a process manager such [systemd](https://www.linode.com/docs/quick-answers/linux/start-service-at-boot/) to run it.

The following example runs the application server using the Docker image.

    $ docker run \
        --tty --rm \
        --env "RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_RECV_MAIN_ADDR=https://queue.amazonaws.com/444455556666/recv" \
        --env "RDSS_ARCHIVEMATICA_ADAPTER_ADAPTER.QUEUE_SEND_MAIN_ADDR=arn:aws:sqs:us-east-2:444455556666:send" \
        --env "AWS_REGION=us-east-1" \
        --env "AWS_ACCESS_KEY=1234" \
        --env "AWS_SECRET_KEY=5678" \
        artefactual/rdss-archivematica-channel-adapter:latest \
            server

Read the [configuration](#Configuration) section before you proceed with the deployment.

## Configuration

All configuration attributes are described in the source code. See [config.go](./app/config.go) for more.

There are sensible defaults in place. You need to pay special attention to the attributes below and tweak them according to your environment:

- `adapter.processing_table`
- `adapter.repository_table`
- `adapter.registry_table`
- `adapter.queue_recv_main_addr`
- `adapter.queue_send_main_addr`

### Configuration file

We use the [TOML configuration file format](https://en.wikipedia.org/wiki/TOML). The configuration file can be indicated via the `--config` command-line argument. When undefined, the application attempts to read from one of the following locations:

- `$HOME/.config/rdss-archivematica-channel-adapter.toml`
- `/etc/archivematica/rdss-archivematica-channel-adapter.toml`

This is a minimal configuration example:

```toml
[adapter]
processing_table = "adapter_processing"
repository_table = "repository_processing"
registry_table = "registry_processing"
queue_recv_main_addr = "https://queue.amazonaws.com/444455556666/recv"
queue_send_main_addr = "arn:aws:sqs:us-east-2:444455556666:send"
```

### Environment variables

Configuration from environment variables have precedence over file-based configuration. All environment variables follow the same naming scheme: `RDSS_ARCHIVEMATICA_ADAPTER_<SECTION>_<ATTRIBUTE>=<VALUE>`. The following is a valid example:

- `RDSS_ARCHIVEMATICA_ADAPTER_LOGGING.LEVEL=DEBUG`<br />
  (section: `LOGGING`, attribute: `LEVEL`, value: `DEBUG`)

### Service dependencies

This application sits between multiple services and assumes access to the following resources and actions.

| Resource      | API action                                              | Configuration                                                                                                                                                     |
|---------------|---------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AWS SQS       | sqs:ReceiveMessage                                      | adapter.queue_recv_main_addr<br/>aws.sqs_profile (optional)<br/>aws.sqs_endpoint (optional)                                                                       |
| AWS SNS       | sns:Publish                                             | adapter.queue_send_main_addr<br/>adapter.queue_send_invalid_addr<br/>adapter.queue_send_error_addr<br/>aws.sns_profile (optional)<br/>aws.sns_endpoint (optional) |
| AWS DynamoDB  | dynamodb:GetItem<br/>dynamodb:PutItem<br/>dynamodb:Scan | adapter.processing_table<br/>adapter.repository_table<br/>adapter.registry_table<br/>aws.dynamodb_profile (optional)<br/>aws.dynamodb_endpoint (optional)         |
| AWS S3        | s3:GetObject                                            | adapter.s3_profile<br/>adapter.s3_endpoint<br/><small>*(only needed when preservation requests point to S3 buckets.)*</small>                                     |
| Archivematica | N/A                                                     | *(configured via the adapter.registry_table)*                                                                                                                     |

SQS/SNS resources are expected to be provisioned by RDSS. The DynamoDB tables are local to the adapter and need to be created by the user. For example, they can be created using the AWS CLI as in the following example:

```
aws dynamodb create-table \
    --table-name="rdss_archivematica_adapter_local_data_repository" \
    --attribute-definitions="AttributeName=ID,AttributeType=S" \
    --key-schema="AttributeName=ID,KeyType=HASH" \
    --billing-mode="PAY_PER_REQUEST"

aws dynamodb create-table \
    --table-name="rdss_archivematica_adapter_processing_state" \
    --attribute-definitions="AttributeName=objectUUID,AttributeType=S" \
    --key-schema="AttributeName=objectUUID,KeyType=HASH" \
    --billing-mode="PAY_PER_REQUEST"

aws dynamodb create-table \
    --table-name="rdss_archivematica_adapter_registry" \
    --attribute-definitions="AttributeName=tenantJiscID,AttributeType=S" \
    --key-schema="AttributeName=tenantJiscID,KeyType=HASH" \
    --billing-mode="PAY_PER_REQUEST"
```

### AWS service client configuration

The AWS service client configuration rely on the [shared configuration functionality](https://docs.aws.amazon.com/sdk-for-go/api/aws/session/) which is similar to the [AWS CLI configuration system](https://docs.aws.amazon.com/cli/latest/topic/config-vars.html).

Additionally, you can override the configuration profile on each client as well as the endpoint using the following environment strings:

- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_PROFILE`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.S3_ENDPOINT`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_PROFILE`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.DYNAMODB_ENDPOINT`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_PROFILE`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.SQS_ENDPOINT`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_PROFILE`
- `RDSS_ARCHIVEMATICA_ADAPTER_AWS.SNS_ENDPOINT`

This can be useful under a variety of scenarios:

- Deployment of alternative services like LocalStack, Minio, etc...
- Applying different credentials, e.g. assuming a IAM role in the SQS/SNS clients.

### Registry of Archivematica pipelines

The adapter uses a registry of Archivematica pipelines stored in DynamoDB (table `adapter.repository_table`) that looks like the following:

| tenantJiscID | url                    | user | key      | transferDir        |
|--------------|------------------------|------|----------|--------------------|
| 1            | http://192.168.1.1/api | user | juoCah3o | /mnt/share/tenant1 |
| 2            | http://192.168.1.2/api | user | Ixie9aid | /mnt/share/tenant2 |

It is possible to create, delete and scan items in [various ways](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GettingStartedDynamoDB.html), including the AWS Management Console. The folowing is an example of item creation using the AWS CLI:

```
env \
    AWS_DEFAULT_REGION="us-east-1" \
    AWS_ACCESS_KEY_ID="1234" \
    AWS_SECRET_ACCESS_KEY="5678" \
        aws dynamodb put-item \
            --table-name="rdss_archivematica_adapter_registry"
            --item "file:///tmp/test-registry-item.json"
```

The previous command loads the record in `/tmp/test-registry-item.json`:

```json
{
    "tenantJiscID": {"S": "3"},
    "url": {"S": "http://192.168.1.3/api"},
    "user": {"S": "user"},
    "key": {"S": "eh6eeDuu"},
    "transferDir": {"S": "/mnt/share/tenant3"}
}
```

The adapter loads the registry in three cases:

- When the application starts.
- Every 10 seconds once the application has been initialized properly.
- When a `USR1` signal is received, e.g.:

      killall -s SIGUSR1 rdss-archivematica-channel-adapter

Send the `USR2` signal to log the current instances loaded:

    killall -s SIGUSR2 rdss-archivematica-channel-adapter

## Metrics and runtime profiling data

`rdss-archivematica-channel-adapter server` runs a HTTP server that listens on `0.0.0.0:6060` with two purposes:

* `/metrics` serves metrics of the Go runtime and the application meant to be scraped by a Prometheus server.
* `/debug/pprof` serves runtime profiling data in the format expected by the pprof visualization tool. Visit [net/http/pprof docs](https:/golang.org/pkg/net/http/pprof/) for more.

## Contributing

* See [CONTRIBUTING.md][1] for information about setting up your environment and the workflow that we expect.
* Check out the [open issues][2].

Also, the [broker][3] package can be used to implement your own RDSS adapter using the Go programming language. The linked docs include documentation and examples. The API stability is not guaranteed.

[1]: /CONTRIBUTING.md
[2]: https://github.com/JiscSD/rdss-archivematica-channel-adapter/issues
[3]: https://godoc.org/github.com/JiscSD/rdss-archivematica-channel-adapter/broker
