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
			"https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/create/request_schema.json",
			"messages/body/metadata/create/request_schema.json",
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

func TestNoOpValidator(t *testing.T) {
	invalidMessage := New(MessageTypeMetadataDelete, MessageClassCommand)
	invalidMessage.body = []byte(`{}`) // Invalid because missing required `objectUuid`.
	validator := &NoOpValidator{}
	res, err := validator.Validate(invalidMessage)
	if !res.Valid() || err != nil {
		t.Fatalf("NoOpValidator.Validate() returned unexpected values")
	}
}

var (
	benchRes          *gojsonschema.Result
	benchErr          error
	benchValidator, _ = NewValidator()
)

func benchmarkValidation(msg *Message, validator Validator, b *testing.B) {
	var (
		res *gojsonschema.Result
		err error
	)
	for n := 0; n < b.N; n++ {
		// Record the result to prevent the compiler eliminating the function
		// call.
		res, err = validator.Validate(msg)
	}
	// Store the results to package level variables so the compiler cannot
	// eliminate the Benchmark itself.
	benchRes, benchErr = res, err
}

func BenchmarkValidationNoOp(b *testing.B) {
	msg := New(MessageTypeMetadataDelete, MessageClassCommand)
	msg.body = []byte(`{}`)
	benchmarkValidation(msg, &NoOpValidator{}, b)
}

func BenchmarkValidationSimple(b *testing.B) {
	msg := New(MessageTypeMetadataDelete, MessageClassCommand)
	msg.body = []byte(`{"objectUuid": 2}`)
	benchmarkValidation(msg, benchValidator, b)
}

func BenchmarkValidationComplex(b *testing.B) {
	msg := New(MessageTypeMetadataDelete, MessageClassCommand)
	msg.body = []byte(`{"title":"Person","type":"object","properties":{"firstName":{"type":"string"},"lastName":{"type":"string"},"age":{"description":"Age in years","type":"integer","minimum":0}},"required":["firstName","lastName"]}`)
	benchmarkValidation(msg, benchValidator, b)
}

func BenchmarkValidationVeryComplex(b *testing.B) {
	msg := New(MessageTypeMetadataDelete, MessageClassCommand)
	msg.body = []byte(`[{"_id":"5a54fa5ce685338d09be0c6c","index":0,"guid":"ff10ae5f-46f9-4db6-a538-471356b84858","isActive":true,"balance":"$1,034.42","picture":"http://placehold.it/32x32","age":38,"eyeColor":"brown","name":"Natalia Nicholson","gender":"female","company":"UXMOX","email":"natalianicholson@uxmox.com","phone":"+1 (987) 501-3782","address":"939 Barwell Terrace, Efland, Oregon, 2426","about":"Nisi laboris ipsum velit sint. Nulla in reprehenderit eiusmod esse do. Fugiat consectetur aliquip cupidatat magna.\r\n","registered":"2018-01-01T02:36:09 +08:00","latitude":-52.446068,"longitude":-98.758356,"tags":["exercitation","elit","nostrud","ad","do","magna","incididunt"],"friends":[{"id":0,"name":"Landry Potter"},{"id":1,"name":"Kirk Oconnor"},{"id":2,"name":"Tyson Donovan"}],"greeting":"Hello, Natalia Nicholson! You have 3 unread messages.","favoriteFruit":"strawberry"},{"_id":"5a54fa5cfea1daec5928fbe5","index":1,"guid":"199b2a3e-828f-4209-ae3e-876fa2bc75f1","isActive":true,"balance":"$1,993.16","picture":"http://placehold.it/32x32","age":37,"eyeColor":"blue","name":"Marguerite Harrison","gender":"female","company":"VINCH","email":"margueriteharrison@vinch.com","phone":"+1 (815) 453-2061","address":"204 Oriental Boulevard, Ironton, Federated States Of Micronesia, 2676","about":"Irure ullamco et sunt eiusmod incididunt exercitation aute amet quis voluptate ut adipisicing consequat. Velit amet sunt reprehenderit id reprehenderit incididunt minim ea excepteur in nulla non. Dolore duis exercitation exercitation esse aute est nulla aute.\r\n","registered":"2017-03-30T02:35:01 +07:00","latitude":-54.794495,"longitude":177.485519,"tags":["nostrud","quis","incididunt","duis","cupidatat","Lorem","anim"],"friends":[{"id":0,"name":"Luisa Baird"},{"id":1,"name":"Parsons Jordan"},{"id":2,"name":"Inez Gutierrez"}],"greeting":"Hello, Marguerite Harrison! You have 9 unread messages.","favoriteFruit":"banana"},{"_id":"5a54fa5c7a705e00944d2967","index":2,"guid":"799ca210-5eb0-41be-92bd-23d3cac9a281","isActive":false,"balance":"$1,564.40","picture":"http://placehold.it/32x32","age":39,"eyeColor":"brown","name":"Velma Wells","gender":"female","company":"STEELTAB","email":"velmawells@steeltab.com","phone":"+1 (862) 417-3951","address":"134 Polhemus Place, Villarreal, New Hampshire, 9068","about":"Adipisicing Lorem quis occaecat est voluptate aliqua irure veniam consequat aliqua laborum irure. Irure non mollit minim ad sit dolor ullamco qui ex irure irure irure esse laboris. Deserunt excepteur velit consectetur cillum laborum est minim laboris in. Esse et ipsum et sunt ea laboris laboris eiusmod reprehenderit. In voluptate elit commodo ea laborum ipsum.\r\n","registered":"2014-07-05T07:14:38 +07:00","latitude":-15.689386,"longitude":-33.225634,"tags":["commodo","minim","ex","ipsum","dolor","consectetur","velit"],"friends":[{"id":0,"name":"Farrell Garrett"},{"id":1,"name":"Lindsey French"},{"id":2,"name":"Cassie Pace"}],"greeting":"Hello, Velma Wells! You have 8 unread messages.","favoriteFruit":"strawberry"},{"_id":"5a54fa5c4f25f1195525f00f","index":3,"guid":"f052e5d4-6e58-46c5-8987-c91f3efe9fb5","isActive":false,"balance":"$2,762.30","picture":"http://placehold.it/32x32","age":23,"eyeColor":"brown","name":"Debbie Conway","gender":"female","company":"EVIDENDS","email":"debbieconway@evidends.com","phone":"+1 (950) 567-3142","address":"121 Woodside Avenue, Ivanhoe, Palau, 7927","about":"Amet sit ipsum incididunt ad Lorem fugiat quis pariatur. Esse amet cupidatat duis dolore magna duis eiusmod officia pariatur. Sint nisi consectetur culpa ex officia anim laboris quis mollit eu et nisi occaecat. Pariatur id in incididunt pariatur commodo fugiat anim eu ex.\r\n","registered":"2016-01-10T03:23:06 +08:00","latitude":-23.31763,"longitude":-100.161341,"tags":["aliquip","labore","nisi","exercitation","reprehenderit","eu","elit"],"friends":[{"id":0,"name":"Geraldine Newman"},{"id":1,"name":"Janna Wiley"},{"id":2,"name":"Faulkner Gross"}],"greeting":"Hello, Debbie Conway! You have 9 unread messages.","favoriteFruit":"strawberry"},{"_id":"5a54fa5c153867a8d054f67b","index":4,"guid":"6ce26b41-cfeb-4a12-9d43-a72c63790e70","isActive":false,"balance":"$1,999.70","picture":"http://placehold.it/32x32","age":33,"eyeColor":"green","name":"Carolyn Beach","gender":"female","company":"INSECTUS","email":"carolynbeach@insectus.com","phone":"+1 (948) 549-3967","address":"928 Grattan Street, Alden, Missouri, 5582","about":"Veniam anim qui esse ea dolor cupidatat duis nulla minim do ipsum. Dolor veniam magna non ut. Laboris qui Lorem aute ea nisi enim nisi aliqua deserunt cupidatat voluptate laborum. Aliquip ipsum cillum est reprehenderit dolore qui est proident tempor sit magna est tempor. Sint ad aliquip occaecat Lorem ipsum commodo.\r\n","registered":"2015-02-28T10:04:32 +08:00","latitude":-67.857079,"longitude":128.484301,"tags":["proident","occaecat","Lorem","mollit","eu","dolor","quis"],"friends":[{"id":0,"name":"Marianne Kline"},{"id":1,"name":"Gates Chavez"},{"id":2,"name":"Moreno Dawson"}],"greeting":"Hello, Carolyn Beach! You have 6 unread messages.","favoriteFruit":"strawberry"},{"_id":"5a54fa5c7936c3032f1cbdc5","index":5,"guid":"51c53dd2-b7f5-49f3-9bf8-fb8d17a098cf","isActive":false,"balance":"$1,577.53","picture":"http://placehold.it/32x32","age":36,"eyeColor":"blue","name":"Mallory Beck","gender":"female","company":"PROVIDCO","email":"mallorybeck@providco.com","phone":"+1 (964) 473-2251","address":"152 Norfolk Street, Orviston, Louisiana, 6420","about":"Amet voluptate adipisicing culpa cupidatat mollit aute deserunt reprehenderit nisi. Aliquip exercitation enim ullamco magna magna quis consectetur ullamco irure Lorem eu mollit aliquip nisi. Dolor ipsum commodo eu esse Lorem non Lorem ipsum ea. Incididunt reprehenderit deserunt laboris labore sit velit cupidatat ex cillum non. Proident cupidatat culpa duis ad id ipsum commodo culpa commodo consectetur consequat ea elit ullamco.\r\n","registered":"2017-01-02T11:36:56 +08:00","latitude":69.579569,"longitude":86.592765,"tags":["id","aute","velit","reprehenderit","nisi","aliqua","et"],"friends":[{"id":0,"name":"Beasley Austin"},{"id":1,"name":"Fleming Miller"},{"id":2,"name":"Jodi Meyer"}],"greeting":"Hello, Mallory Beck! You have 9 unread messages.","favoriteFruit":"banana"},{"_id":"5a54fa5c06081c418c65cf40","index":6,"guid":"49221710-6453-4ff3-b350-9c496f94e31b","isActive":false,"balance":"$2,527.48","picture":"http://placehold.it/32x32","age":29,"eyeColor":"brown","name":"Bernard Mckenzie","gender":"male","company":"NIPAZ","email":"bernardmckenzie@nipaz.com","phone":"+1 (862) 484-2710","address":"343 Bainbridge Street, Cecilia, North Dakota, 2089","about":"Velit pariatur nulla dolor commodo minim occaecat. Elit voluptate dolor mollit consectetur cillum proident. Proident id culpa excepteur reprehenderit consectetur eu veniam enim. Tempor pariatur officia eiusmod laborum consequat laborum fugiat velit non cupidatat duis id elit nisi. Aute nisi irure ad aliquip pariatur commodo fugiat reprehenderit pariatur Lorem.\r\n","registered":"2017-07-31T06:06:45 +07:00","latitude":76.738851,"longitude":-19.850087,"tags":["anim","quis","nisi","consequat","officia","dolore","irure"],"friends":[{"id":0,"name":"Daphne Glenn"},{"id":1,"name":"Haley Ramsey"},{"id":2,"name":"Ester Ferrell"}],"greeting":"Hello, Bernard Mckenzie! You have 1 unread messages.","favoriteFruit":"banana"}]`)
	benchmarkValidation(msg, benchValidator, b)
}
