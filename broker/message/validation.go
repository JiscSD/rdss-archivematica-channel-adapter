package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message/specdata"
)

func init() {
	// The checker provider by the library isn't perfect.
	gojsonschema.FormatCheckers.Add("email", EmailFormatChecker{})
}

type Validator interface {
	// Validate returns the results of validating a document against its schema.
	// The error returned is non-nil when the validation process failed. If nil,
	// the caller still needs to check the validation state given by Result.
	Validate(msg *Message) (*gojsonschema.Result, error)

	// Validators returns the validators available.
	Validators() map[string]*gojsonschema.Schema
}

// rdssValidator implements Validator. It is the default validation solution for
// the RDSS API and it depends on its schema files.
type rdssValidator struct {
	validators map[string]*gojsonschema.Schema
}

var _ Validator = rdssValidator{}

// rdssPrefix is used by our schema loder so we can differentiate local
// references from external references. When rdssPrefix is matched the schema
// loader will load our internal schemas persisted in the specdata package.
// It is not a constant becuase we change it in our tests.
var rdssPrefix = "https://www.jisc.ac.uk/rdss/schema/"

var rdssSchemas = map[string]string{
	"header":                 "https://www.jisc.ac.uk/rdss/schema/messages/header/header_schema.json",
	"MetadataCreateRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/create/request_schema.json",
	"MetadataDeleteRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/delete/request_schema.json",
	"MetadataReadRequest":    "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/read/request_schema.json",
	"MetadataReadResponse":   "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/read/response_schema.json",
	"MetadataUpdateRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/update/request_schema.json",
	"VocabularyPatchRequest": "https://www.jisc.ac.uk/rdss/schema/messages/body/vocabulary/patch/request_schema.json",
	"VocabularyReadRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/vocabulary/read/request_schema.json",
	"VocabularyReadResponse": "https://www.jisc.ac.uk/rdss/schema/messages/body/vocabulary/read/response_schema.json",
}

// schemaDocFinder is a function that given a reference like the ones found in
// `rdssSchema` returns the corresponding stream of bytes of the schema document
// that corresponds.
type schemaDocFinder func(string) ([]byte, error)

// DefaultSchemaDocFinder is used by localSchemaLoader to read the schema
// documents from the local store. The default function depends on the
// `specdata` package which is populated at build time. The tests use a custom
// function so it can populate arbitrary documents.
var DefaultSchemaDocFinder schemaDocFinder = rdssSchemaDocFinder

// rdssSchemaDocFinder fetches the schema documents from the `specdata`
// package.
func rdssSchemaDocFinder(source string) ([]byte, error) {
	return specdata.Asset(resolveSchemaRef(source))
}

// resolveSchemaRef knows the path of the schema assets in the specdata pkg.
func resolveSchemaRef(source string) string {
	source = strings.TrimSuffix(source, "/")
	source = strings.TrimPrefix(source, rdssPrefix)

	// This is what makes possible to have references like:
	//   https://www.jisc.ac.uk/rdss/schema/types.json
	// In addition to:
	//   https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/create/request_schema.json
	if !strings.HasPrefix(source, "messages/") {
		source = fmt.Sprintf("schemas/%s", source)
	}

	return source
}

// NewValidator returns a Validator with all the RDSS API schemas loaded.
func NewValidator() (Validator, error) {
	v := &rdssValidator{
		validators: make(map[string]*gojsonschema.Schema),
	}

	// Initialize schema validators.
	loader := NewLocalSchemaLoaderFactory()
	for name, path := range rdssSchemas {
		schema, err := gojsonschema.NewSchema(loader.New(path))
		if err != nil {
			return nil, errors.Wrapf(err, "schema loader could not create a schema object from %s", path)
		}
		v.validators[name] = schema
	}

	return v, nil
}

// Validate implementes Validator. It aggregates the results of both the header
// and the body validation results.
func (v rdssValidator) Validate(msg *Message) (*gojsonschema.Result, error) {
	res, err := v.validateBody(msg.body, msg.Type())
	if err != nil {
		return nil, err
	}
	if len(msg.header) == 0 {
		return res, nil
	}
	hr, err := v.validateHeader(msg.header)
	if err != nil {
		return nil, err
	}
	if !hr.Valid() {
		for _, item := range hr.Errors() {
			res.AddError(item, item.Details())
		}
	}
	return res, err
}

func (v rdssValidator) validateHeader(data []byte) (*gojsonschema.Result, error) {
	const schema = "header"
	loader := gojsonschema.NewBytesLoader(data)
	val, ok := v.validators[schema]
	if !ok {
		return nil, fmt.Errorf("validator for %s does not exist", schema)
	}
	return val.Validate(loader)
}

func (v rdssValidator) validateBody(data []byte, mType string) (*gojsonschema.Result, error) {
	loader := gojsonschema.NewBytesLoader(data)
	val, ok := v.validators[mType]
	if !ok {
		return nil, fmt.Errorf("validator for %s does not exist", mType)
	}
	return val.Validate(loader)
}

// Validators implementes Validator.
func (v rdssValidator) Validators() map[string]*gojsonschema.Schema {
	return v.validators
}

type localSchemaLoaderFactory struct {
	gojsonschema.JSONLoaderFactory
}

type localSchemaLoader struct {
	gojsonschema.JSONLoader
	source  string
	factory localSchemaLoaderFactory
}

func NewLocalSchemaLoaderFactory() localSchemaLoaderFactory {
	return localSchemaLoaderFactory{}
}

func (f localSchemaLoaderFactory) New(source string) gojsonschema.JSONLoader {
	return &localSchemaLoader{
		JSONLoader: gojsonschema.NewReferenceLoader(source),
		source:     source,
		factory:    f,
	}
}

func (l *localSchemaLoader) JsonSource() interface{} {
	return l.source
}

func (l *localSchemaLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference(l.source)
}

func (l *localSchemaLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return l.factory
}

func (l *localSchemaLoader) LoadJSON() (interface{}, error) {
	ref, err := l.JsonReference()
	if err != nil {
		return nil, err
	}
	url := ref.GetUrl().String()
	if !strings.HasPrefix(url, rdssPrefix) {
		return l.JSONLoader.LoadJSON()
	}
	blob, err := DefaultSchemaDocFinder(url)
	if err != nil {
		return nil, err
	}
	return decodeJson(bytes.NewReader(blob))
}

func decodeJson(r io.Reader) (interface{}, error) {
	var document interface{}
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	return document, nil
}

// NoOpValidator is a validator that recognizes all messages as valid.
type NoOpValidator struct{}

var _ Validator = NoOpValidator{}

// Validate implementes Validator.
func (v NoOpValidator) Validate(msg *Message) (*gojsonschema.Result, error) {
	return &gojsonschema.Result{}, nil
}

// Validators implementes Validator.
func (v NoOpValidator) Validators() map[string]*gojsonschema.Schema {
	return map[string]*gojsonschema.Schema{}
}

// EmailFormatChecker is a custom emailFormatChecker. The one provided by the
// gojsonschema library is perfect but we want to support the edge case
// "person@net" which is used in the spec. It's considered valid by in the spec
// test suite (Python's jsonschema package).
type EmailFormatChecker struct{}

// IsFormat implements gojsonschema.FormatChecker.
func (f EmailFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if ok == false {
		return false
	}
	const addressUsedInSpecExamples = "person@net"
	if asString == addressUsedInSpecExamples {
		return true
	}
	emailFC := &gojsonschema.EmailFormatChecker{}
	return emailFC.IsFormat(asString)
}

// See https://github.com/JiscRDSS/rdss-message-api-specification/commit/81af7c27c4adb10bba05ced436347789e67d6a14.
const regex = "^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(-([0-9A-Za-z-]+\\.)*[0-9A-Za-z-]+)?(\\+([0-9A-Za-z-]+\\.)*[0-9A-Za-z-]+)?$"

var versionRegexp = regexp.MustCompile(regex)

type VersionInvalidError struct {
	gojsonschema.ResultErrorFields
}

// ValidateVersion validates a version string.
//
// Our jsonschema library doesn't seem to support patterns. It could be
// possible to have the spec changed so it uses the "regex" format instead but
// for now we're just going to be validating this attribute manually.
func ValidateVersion(ver string, result *gojsonschema.Result) {
	if versionRegexp.MatchString(ver) {
		// Stop here if we have a match.
		return
	}
	details := gojsonschema.ErrorDetails{}
	err := &VersionInvalidError{}
	err.SetContext(gojsonschema.NewJsonContext("version", nil))
	err.SetType("invalid_version")
	err.SetValue(ver)
	err.SetDetails(details)
	err.SetDescriptionFormat("Invalid version format")
	result.AddError(err, details)
}
