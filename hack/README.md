## Vendoring

We're using `dep` which is not stable yet but it's good enough for what we
need. Run `dep status` to know more about our dependencies and constraints
established.

## Development dependencies

Please see the root `Makefile` or use `make tools` to download the tools we're
using in this project.

## Run Kinesis locally

Check out `minikine`, a small local Kinesis server based on kinesalite used for
testing purposes. You will NodeJS and run `npm install` inside the directory to
download its dependencies.

## Generate the gRPC protobuf stubs

By default gRPC uses protocol buffers. You will need the `protoc` compiler to
generate stub server and client code, which is listed as a development
dependency above. Use the [build-proto.sh](build-proto.sh) script to do it
automatically.

## Download the schema files

The schema files can be downloaded with the following command:

    $ ./download-schemas.sh $GH_API_TOKEN

You have to create your own
[personal API token](https://github.com/settings/tokens).

## Running tests

Use Go 1.9 or newer.

Run the following command in the root directory:

    $ go test -race ./...

Also from the root directory you can run the validator tests against the
fixtures as long as you have already [downloaded the schema files](#download-the-schema-files).
This is the command that you need to run from the root directory:

    $ go test ./broker/message -validator.enabled=true -validator.schemasDir=$PWD/hack/schemas
