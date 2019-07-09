/*
Package message provide types and functions to work with the RDSS API.

This package reflect what is in the specification repository:
https://github.com/JiscRDSS/rdss-message-api-specification.

These are the main steps that need to be taken after a new release is made in
the specification repository:

1) Update the commit of the message-api-spec git submodule so it points to the
git tag desired, e.g. "v4.0.0".

2) Update the Version string in version.go.

3) Build specdata/specdata.go using hack/build-spec.go. The outcome is a bucket
containing all the files of the spec submodule. This is so they can be embedded
into the binary, e.g. for schema validation purposes, access to the schema files
is required.

4) When the API enumeration.json schema changes, update the corresponding
shared_enumeration_gen.go using go generate. The generation happens in
generator.go using text/template and go/format. Run:

    go generate ./broker/message
*/
package message
