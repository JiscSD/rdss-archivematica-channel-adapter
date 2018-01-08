package message

import (
	"bytes"
	"errors"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

func TestNewValidator(t *testing.T) {
	var schema = []byte(`{
		"id": "mySchema",
		"$schema": "http://json-schema.org/draft-04/schema#",
		"objectUuid": {
			"type": "string"
		},
		"required": [
			"objectUuid"
		]
	}`)
	DefaultSchemaDocFinder = func(string) ([]byte, error) { return schema, nil }
	validator, err := NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	if have, want := len(validator.Validators()), len(rdssSchemas); have != want {
		t.Fatalf("NewValidator did not load all the schemas expected; have %d, want %d", have, want)
	}

	// Invalid message.
	msg := New(MessageTypeMetadataDelete, MessageClassCommand)
	msg.body = []byte(`{"foo": "bar"}`)
	res, err := validator.Validate(msg)
	if err != nil {
		t.Fatal(err)
	}
	if res.Valid() {
		t.Fatal("message was invalid but the validator did not identify it as such")
	}

	// Valid message.
	msg.body = []byte(`{"objectUuid": "25d23261-37ec-4603-8c3d-e34715f19621"}`)
	res, err = validator.Validate(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Valid() {
		t.Fatal("message was valid but the validator did not identify it as such")
	}

	// Test missing validator.
	delete(rdssSchemas, "MetadataDeleteRequest")
	validator, err = NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	_, err = validator.Validate(msg)
	if err.Error() != "validator for MetadataDeleteRequest does not exist" {
		t.Fatalf("expected error did not occur; got %v", err)
	}
}

func TestNewValidator_WithError(t *testing.T) {
	var notFoundError = errors.New("schema not found")
	DefaultSchemaDocFinder = func(string) ([]byte, error) { return nil, notFoundError }
	validator, err := NewValidator()
	if validator != nil {
		t.Fatalf("NewValidator() returned a non-nil value, got %v", validator)
	}
	if err == nil {
		t.Fatalf("NewValidator() returned an unexpected error: %v", err)
	}
}

func Test_resolveSchemaRef(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{
			"https://www.jisc.ac.uk/rdss/schema/types.json",
			"schemas/types.json",
		},
		{
			"https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/",
			"schemas/intellectual_asset.json",
		},
		{
			"https://www.jisc.ac.uk/rdss/schema/messages/header/header_schema.json",
			"messages/header/header_schema.json",
		},
	}
	for _, tt := range tests {
		if got := resolveSchemaRef(tt.source); got != tt.want {
			t.Errorf("resolveSchemaRef() = %v, want %v", got, tt.want)
		}
	}
}

func mockSchemaDocFinder(blobs map[string][]byte) schemaDocFinder {
	return func(source string) ([]byte, error) {
		return blobs[source], nil
	}
}

func TestLocalSchemaLoaderFactory(t *testing.T) {
	// Create our own schema document finder that provides schema documents for
	// both https://foobar.tld/schema.json and https://foobar.tld/types.json.
	rdssPrefix = "https://foobar.tld/schemas/"
	DefaultSchemaDocFinder = mockSchemaDocFinder(map[string][]byte{
		"https://foobar.tld/schemas/schema.json": []byte(`{
			"id": "https://foobar.tld/schemas/schema.json/#",
			"$schema": "http://json-schema.org/draft-04/schema#",
			"definitions": {
				"shortString": {
					"type": "string",
					"maxLength": 6
				}
			},
			"properties": {
				"name": {
					"$ref": "#/definitions/shortString"
				},
				"age": {
					"$ref": "https://foobar.tld/schemas/types.json#/definitions/positiveInteger"
				}
			},
			"required": [
				"name",
				"age"
			]
		}`),
		"https://foobar.tld/schemas/types.json": []byte(`{
			"id": "https://foobar.tld/schemas/types.json#",
			"$schema": "http://json-schema.org/draft-04/schema#",
			"definitions": {
				"positiveInteger": {
					"type": "integer",
					"minimum": 1
				}
			}
		}`),
	})
	// NewLocalSchemaLoaderFactory relies on DefaultSchemaDocFinder.
	// So we should be able to load the schema and validate docs against it.
	factory := NewLocalSchemaLoaderFactory()
	schema, err := gojsonschema.NewSchema(factory.New("https://foobar.tld/schemas/schema.json"))
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		doc   []byte
		valid bool
	}{
		// Valid. All good!
		{[]byte(`{"name": "Foobar", "age": 21}`), true},
		// Invalid. Missing required properties.
		{[]byte(`{}`), false},
		// Invalid. The name is too long.
		{[]byte(`{"name": "Foobar Barfoo", "age": 21}`), false},
		// Invalid. The age should be greater or equal than 1.
		{[]byte(`{"name": "Foobar", "age": -4}`), false},
	}
	for i, tc := range testCases {
		res, err := schema.Validate(gojsonschema.NewBytesLoader(tc.doc))
		if err != nil {
			t.Fatal(err)
		}
		if tc.valid != res.Valid() {
			t.Errorf("test %d failed: unexpected validation result; have \"%t\", got \"%t\"", i, tc.valid, res.Valid())
			if !res.Valid() {
				for _, err := range res.Errors() {
					t.Log(err)
				}
			}
		}
	}
}

func TestLocalSchemaLoaderFactory_UnknownSchema(t *testing.T) {
	rdssPrefix = "https://foobar.tld/schemas/"
	DefaultSchemaDocFinder = mockSchemaDocFinder(map[string][]byte{
		"https://foobar.tld/schemas/schema.json": []byte(`{
			"id": "https://foobar.tld/schemas/schema.json/#",
			"$schema": "http://json-schema.org/draft-04/schema#",
			"properties": {
				"prop": {
					"$ref": "file:///unknown.json/#/definitions/positiveInteger"
				}
			}
		}`),
	})

	factory := NewLocalSchemaLoaderFactory()
	_, err := gojsonschema.NewSchema(factory.New("https://foobar.tld/schemas/schema.json"))
	if err.Error() != "open /unknown.json/: no such file or directory" {
		t.Fatalf("unexpected error; have %v", err)
	}
}

func Test_decodeJson(t *testing.T) {
	testCases := []struct {
		name     string
		blob     []byte
		wantFail bool
	}{
		{"Valid JSON", []byte("Not A JSON document"), true},
		{"Invalid JSON", []byte("{}"), false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewReader(tc.blob)
			_, err := decodeJson(reader)
			failed := err != nil
			if tc.wantFail != failed {
				t.Errorf("decodeJson() unexpected error: %v", err)
			}
		})
	}
}
