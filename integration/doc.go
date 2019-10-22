// `make test-integration` is the simplest way to run these tests.
//
// `go test` flags supported:
//
//   -debug
//
//    Enable debug mode.
//
//   -valsvc="http://..."
//
//    Define the schema servide address for validation.
//    If undefined, validation/conversion state is not executed.
//
// Example: go test -v ./integration/... -valsvc="http://...""
//
package integration
