/*
Package message provide types and functions to work with the RDSS API.

In order to update to the latest release:

- Update the commit of the message-api-spec module so it points to the tag
  desired.
- Update the Version string in version.go.
- Build specdata/specdata.go using hack/build-spec.go. The outcome is a bucket
  containing all the files of the spec submodule so they can be embedded into
  the binary, e.g. for schema validation purposes, access to the schema files
  is required.
- When the API enumeration.json schema changes, update the corresponding
  shared_enumeration_gen.go using go generate. The generation happens in
  generator.go using text/template and go/format.
*/
package message
