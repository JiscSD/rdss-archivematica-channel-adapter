# Contributing

- [Building from source](#building-from-source)
  - [Prerequisites](#prerequisites)
  - [Fetch the source](#fetch-the-source)
  - [Building](#building)
- [Development](#development)
  - [Run unit tests](#run-unit-tests)
  - [Run integration tests](#run-integration-tests)
  - [Makefile](#makefile)
- [Running the adapter locally](#running-the-adapter-locally)

## Building from source

This section describes how to build the application from source.

### Prerequisites

Install [Go 1.12][1] or newer.

### Fetch the source

We use [`Go modules`][2] for dependency management.

Clone the repository:

    git clone --recurse-submodules https://github.com/JiscSD/rdss-archivematica-channel-adapter

### Building

To build the application, run:

    go install

The binary `rdss-archivematica-channel-adapter` should be in your `$GOPATH/bin`
directory. Run `go env GOPATH` to know where your `GOPATH` is.

## Development

### Run unit tests

    make test

### Run integration tests

The integration test suite can be executed with:

    make test-integration

This is the high-level view of what's happening:

- Compile the application binary.
- Run LocalStack (local AWS cloud stack) with Docker Compose.
- Provision required resources, e.g. DynamoDB tables, SQS/SNS topics...
- Run tests in the `integration` package.

### Makefile

The root `Makefile` includes rules relevant to development and continuous
integration workflows.

## Running the adapter locally

Install the binary:

    make install

Run LocalStack:

    docker-compose -f integration/docker-compose.yml up -d

Run the adapter:

    ./integration/scripts/example_run.sh

You may see a warning: `Registry has been loaded but it is empty`. The adapter
is telling us that the registry of Archivematica pipelines is empty. We need to
populate the registry with entries that associate Jisc tenants to Archivematica
pipelines.

Let's start creating a new registry item to associate the tenant with ID `2` to
our local Archivematica pipeline. Let's store it in disk,
e.g. `/tmp/entry.json`, with the following contents which you can tweak
according to the Archivematica connection details that you are running locally:

```json
{
    "tenantJiscID": {"S": "2"},
    "url": {"S": "http://localhost:62080/api"},
    "user": {"S": "test"},
    "key": {"S": "test"},
    "transferDir": {"S": "/home/username/.am/ss-location-data/amclient-tenant-1"}
}
```

Now we're going to load the entry into the DynamoDB table.  With AWS CLI
installed (`awscli` in [pipy][3]), run:

    env \
        AWS_DEFAULT_REGION="us-east-1" \
        AWS_ACCESS_KEY_ID="1234" \
        AWS_SECRET_ACCESS_KEY="5678" \
            aws --endpoint-url="http://localhost:4569" \
                dynamodb put-item \
                    --table-name="rdss_archivematica_adapter_registry" \
                    --item "file:///tmp/item.json"

The adapter looks up the registry every 10 seconds so chances are that it
already knows about your new configuration when you're done reading this. You
can also force the reload sending a `SIGUSR1` signal to the adapter process, or
a `SIGUSR2` signal to log the existing entries identified.

With our new configuration, a `MetadataCreate` message coming from RDSS with
the tenant ID `2` will be transformed into a new transfer request to
Archivematica with the contents described in the message. The connection details
found in the entry we added to the registry will be used to access the pipeline.

The value that we've used in `transferDir` works well in the default
Archivematica development environment, but make sure that the directory is
created:

    mkdir /home/username/.am/ss-location-data/amclient-tenant-1

There should also be a Transfer Source Location (marked as *default*) pointing
to that directory. Inside your Storage Service container, that's probably
`/home/amclient-tenant-1`.

> The directory created by the adapter under `transferDir` has permissions bits
> `0700` before umask is applied. One simple way to work around permission
> issues is to run the adapter with a local user with uid `333` and gid `333`.

At this point, we can send messages to the adapter. The `integration` package
contains a helper to send a `MetadataCreate` message for tenant `2`. Run:

    ./integration/scripts/example_send.sh

You can also open the [message][4] and edit it. Once the message is delivered
to the adapter, it starts the processing according to the specification. A valid
`MetadataCreate` message should result in a new transfer in Archivematica and
publishing a message back to the corresponding SNS topic.

There are a total of three SNS topics where the adapter is going to be
publishing messages according to the specification. They're hard to observe in
isolation. The integration test suite sets up SQS queues subscribed to the SNS
topics- see [server_test.go](./integration/server_test.go) for more. It is
possible to do [similarly](https://gugsrs.com/localstack-sqs-sns/) using AWS
CLI.

[1]: https://golang.org/doc/install
[2]: https://github.com/golang/go/wiki/Modules
[3]: https://pypi.org/project/awscli/
[4]: ./integration/testdata/message-metadata-create.json
