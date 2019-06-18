package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/twinj/uuid"
	"github.com/xeipuuv/gojsonschema"

	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message/specdata"
)

// Test if we can recreate `message.json` from Go and test if the result is the
// same byte to byte. `message.json` is a full message (including headers) that
// we can find in the API repository.
func TestMessage_ToJSON(t *testing.T) {
	// Load fixture.
	fixture := specdata.MustAsset("messages/example_message.json")

	t.Skip("skipping test because example_message.json has typos")

	// Our message.
	message := &Message{
		MessageHeader: MessageHeader{
			ID:            MustUUID("e3a18f48-9ccf-456b-96c5-784ae8eee63d"),
			MessageClass:  MessageClassEnum_Command,
			MessageType:   MessageTypeEnum_MetadataCreate,
			ReturnAddress: "A free text string",
			MessageTimings: MessageTimings{
				PublishedTimestamp:  Timestamp(time.Date(2004, time.August, 1, 10, 0, 0, 0, time.UTC)),
				ExpirationTimestamp: Timestamp(time.Date(2004, time.August, 1, 10, 0, 0, 0, time.UTC)),
			},
			MessageSequence: MessageSequence{
				Sequence: MustUUID("b66be1c2-e610-461e-bc49-14a42c0b5d24"),
				Position: 1,
				Total:    1,
			},
			MessageHistory: []MessageHistory{
				MessageHistory{
					MachineID:      "A free text string",
					MachineAddress: "machine.example.com",
					Timestamp:      Timestamp(time.Date(2004, time.August, 1, 10, 0, 0, 0, time.UTC)),
				},
			},
			Version:      "4.0.0",
			Generator:    "A free text string",
			TenantJiscID: 2,
		},
		MessageBody: &MetadataCreateRequest{
			ResearchObjectBase{
				ResearchObject: &ResearchObject{
					ObjectUUID:  MustUUID("5680e8e0-28a5-4b20-948e-fd0d08781e0b"),
					ObjectTitle: "A free text string",
					ObjectPersonRole: []PersonRole{
						PersonRole{
							Person: Person{
								PersonUUID: MustUUID("d191abec-71fd-410e-9929-3b18d93587bc"),
								PersonIdentifier: []PersonIdentifier{
									PersonIdentifier{
										PersonIdentifierValue: "A free text string",
										PersonIdentifierType:  PersonIdentifierTypeEnum_researcherID,
									},
								},
								PersonHonorificPrefix: "A free text string",
								PersonGivenNames:      "A free text string",
								PersonFamilyNames:     "A free text string",
								PersonHonorificSuffix: "A free text string",
								PersonMail:            "email_address@jisc.ac.uk",
							},
							Role: PersonRoleEnum_author,
						},
						PersonRole{
							Person: Person{
								PersonUUID: MustUUID("27811a4c-9cb5-4e6d-a069-5c19288fae58"),
								PersonIdentifier: []PersonIdentifier{
									PersonIdentifier{
										PersonIdentifierValue: "A free text string",
										PersonIdentifierType:  PersonIdentifierTypeEnum_researcherID,
									},
								},
								PersonHonorificPrefix: "A free text string",
								PersonGivenNames:      "A free text string",
								PersonFamilyNames:     "A free text string",
								PersonHonorificSuffix: "A free text string",
								PersonMail:            "a_legitimate_email_address@jisc.ac.uk",
								PersonOrganisationUnit: &OrganisationUnit{
									OrganisationUnitUUID: MustUUID("28be7f16-0e70-461f-a2db-d9d7c64a8f17"),
									OrganisationUnitName: "A free text string",
									Organisation: Organisation{
										OrganisationJiscId:  1,
										OrganisationName:    "A free text string",
										OrganisationType:    OrganisationTypeEnum_professionalBody,
										OrganisationAddress: "A free text string",
									},
								},
							},
							Role: PersonRoleEnum_editor,
						},
					},
					ObjectDescription: []ObjectDescription{
						ObjectDescription{
							DescriptionValue: "A free text string",
							DescriptionType:  DescriptionTypeEnum_description,
						},
					},
					ObjectRights: Rights{
						RightsStatement: []string{"A free text string"},
						RightsHolder:    []string{"A free text string"},
						Licence: []Licence{
							Licence{
								LicenceName:       "A free text string",
								LicenceIdentifier: "A free text string",
								LicenseStartDate:  Timestamp(time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)),
								LicenseEndDate:    Timestamp(time.Date(2018, time.December, 31, 23, 59, 59, 0, time.UTC)),
							},
						},
						Access: []Access{
							Access{
								AccessType:      AccessTypeEnum_open,
								AccessStatement: "A free text string",
							},
						},
					},
					ObjectDate: []Date{
						Date{
							DateValue: Timestamp(time.Date(2002, time.October, 2, 10, 0, 0, 0, time.FixedZone("", -18000))),
							DateType:  DateTypeEnum_created,
						},
					},
					ObjectKeyword: []string{
						"A free text string",
					},
					ObjectCategory: []string{
						"A free text string",
					},
					ObjectResourceType: ResourceTypeEnum_text,
					ObjectValue:        ObjectValueEnum_high,
					ObjectIdentifier: []Identifier{
						Identifier{
							IdentifierValue: "A free text string",
							IdentifierType:  IdentifierTypeEnum_URL,
						},
					},
					ObjectRelatedIdentifier: []IdentifierRelationship{
						IdentifierRelationship{
							Identifier: Identifier{
								IdentifierValue: "A free text string",
								IdentifierType:  IdentifierTypeEnum_DOI,
							},
							RelationType: RelationTypeEnum_cites,
						},
					},
					ObjectOrganisationRole: []OrganisationRole{
						OrganisationRole{
							Organisation: Organisation{
								OrganisationJiscId: 1,
								OrganisationName:   "A free text string",
							},
							Role: OrganisationRoleEnum_sponsor,
						},
					},
					ObjectFile: []File{
						File{
							FileUUID:       MustUUID("e150c4ab-0370-4e5a-8722-7fb3369b7017"),
							FileIdentifier: "A free text string",
							FileName:       "A free text string",
							FileSize:       789132,
							FileChecksum: []Checksum{
								Checksum{
									ChecksumUUID:  MustUUID("df23b46b-6b64-4a40-842f-5ad363bb6e11"),
									ChecksumType:  ChecksumTypeEnum_md5,
									ChecksumValue: "A free text string",
								},
							},
							FileCompositionLevel: "A free text string",
							FileDateModified: []Timestamp{
								Timestamp(time.Date(2002, time.October, 2, 10, 0, 0, 0, time.FixedZone("", -18000))),
							},
							FileUse:             FileUseEnum_serviceFile,
							FileUploadStatus:    UploadStatusEnum_uploadStarted,
							FileStorageStatus:   StorageStatusEnum_online,
							FileStorageLocation: "https://tools.ietf.org/html/rfc3986",
							FileStoragePlatform: FileStoragePlatform{
								StoragePlatformUUID: MustUUID("f2939501-2b2d-4e5c-9197-0daa57ccb621"),
								StoragePlatformName: "A free text string",
								StoragePlatformType: StorageTypeEnum_HTTP,
								StoragePlatformCost: "A free text string",
							},
						},
					},
				},
			},
		},
	}

	// Encode our message.
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}

	// Indent the document and append the line break.
	var out bytes.Buffer
	json.Indent(&out, data, "", "  ")
	have := out.Bytes()
	have = append(have, byte('\n'))

	if !bytes.Equal(have, fixture) {
		t.Errorf("Unexpected result:\nHAVE: `%s`\nEXPECTED: `%s`", have, fixture)
		// ioutil.WriteFile("/tmp/test_message_1.txt", have, 0644)
		// ioutil.WriteFile("/tmp/test_messgae_2", fixture, 0644)
	}
}

func TestMessage_New(t *testing.T) {
	msg := New(MessageTypeEnum_MetadataCreate, MessageClassEnum_Command)
	if !reflect.DeepEqual(msg.MessageBody, new(MetadataCreateRequest)) {
		t.Error("Unexexpected type of message body")
	}
	if id, err := uuid.Parse(msg.ID()); err != nil {
		t.Errorf("ID generated is not a UUID: %v", id)
	}
}

func TestMessage_ID(t *testing.T) {
	m := &Message{
		MessageHeader: MessageHeader{ID: NewUUID()},
		MessageBody:   typedBody(MessageTypeEnum_MetadataRead, nil),
	}
	if have, want := m.ID(), m.MessageHeader.ID.String(); have != want {
		t.Errorf("Unexpected ID; have %v, want %v", have, want)
	}
}

func TestMessage_TagError(t *testing.T) {
	m := New(MessageTypeEnum_MetadataCreate, MessageClassEnum_Command)
	if m.TagError(nil); m.MessageHeader.ErrorCode != "" || m.MessageHeader.ErrorDescription != "" {
		t.Error("m.TagError(nil): unexpected error headers")
	}

	m = New(MessageTypeEnum_MetadataCreate, MessageClassEnum_Command)
	if m.TagError(errors.New("foobar")); m.MessageHeader.ErrorCode != "Unknown" || m.MessageHeader.ErrorDescription != "foobar" {
		t.Error("m.TagError(errors.New('foobar')): unexpected error headers")
	}

	m = New(MessageTypeEnum_MetadataCreate, MessageClassEnum_Command)
	if m.TagError(bErrors.New(bErrors.GENERR001, "foobar")); m.MessageHeader.ErrorCode != "GENERR001" || m.MessageHeader.ErrorDescription != "foobar" {
		t.Error("m.TagError(errors.New('foobar')): unexpected error headers")
	}
}

func TestMessage_typedBody(t *testing.T) {
	tests := []struct {
		t             MessageTypeEnum
		correlationID *UUID
		want          interface{}
	}{
		{MessageTypeEnum_MetadataCreate, nil, new(MetadataCreateRequest)},
		{MessageTypeEnum_MetadataRead, nil, new(MetadataReadRequest)},
		{MessageTypeEnum_MetadataRead, NewUUID(), new(MetadataReadResponse)},
		{MessageTypeEnum_MetadataUpdate, nil, new(MetadataUpdateRequest)},
		{MessageTypeEnum_MetadataDelete, nil, new(MetadataDeleteRequest)},
		{MessageTypeEnum(-1), nil, nil},
	}
	for _, tt := range tests {
		if got := typedBody(tt.t, tt.correlationID); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("typedBody() = %v, want %v", got, tt.want)
		}
	}
}

var sharedTests = []struct {
	name        string
	pathFixture string
	t           MessageTypeEnum
	c           MessageClassEnum
	isResponse  bool
}{
	{"MetadataCreateRequest", "messages/body/metadata/create/article_create_request.json", MessageTypeEnum_MetadataCreate, MessageClassEnum_Command, false},
	{"MetadataCreateRequest", "messages/body/metadata/create/dataset_create_request.json", MessageTypeEnum_MetadataCreate, MessageClassEnum_Command, false},
	{"MetadataCreateRequest", "messages/body/metadata/create/research_object_create_request.json", MessageTypeEnum_MetadataCreate, MessageClassEnum_Command, false},
	{"MetadataCreateRequest", "messages/body/metadata/create/thesis_dissertation_create_request.json", MessageTypeEnum_MetadataCreate, MessageClassEnum_Command, false},
	{"MetadataDeleteRequest", "messages/body/metadata/delete/research_object_delete_request.json", MessageTypeEnum_MetadataDelete, MessageClassEnum_Command, false},
	{"MetadataReadRequest", "messages/body/metadata/read/research_object_read_request.json", MessageTypeEnum_MetadataRead, MessageClassEnum_Command, false},
	{"MetadataReadResponse", "messages/body/metadata/read/research_object_read_response.json", MessageTypeEnum_MetadataRead, MessageClassEnum_Command, true},
	// {"MetadataUpdateRequest", "messages/body/metadata/update/research_object_update_request.json", MessageTypeMetadataUpdate, MessageClassCommand, false},
	{"PreservationEventRequest", "messages/body/preservation/preservation_event_request.json", MessageTypeEnum_PreservationEvent, MessageClassEnum_Event, false},
}

func TestMessage_DecodeFixtures(t *testing.T) {
	var validator = getValidator(t)
	for _, tt := range sharedTests {
		t.Run(tt.name, func(t *testing.T) {
			blob := specdata.MustAsset(tt.pathFixture)
			dec := json.NewDecoder(bytes.NewReader(blob))

			var correlationID *UUID
			if tt.isResponse {
				correlationID = MustUUID("bddccd20-f548-11e7-be52-730af1229478")
			}

			// Validation test
			msg := New(tt.t, tt.c)
			msg.MessageHeader.CorrelationID = correlationID
			msg.MessageBody = typedBody(tt.t, correlationID)
			msg.body = blob
			res, err := validator.Validate(msg)
			if err != nil {
				t.Fatal("validator failed:", err)
			}
			if !res.Valid() {
				for _, err := range res.Errors() {
					t.Log("validation error:", err)
				}
				t.Error("validator reported that the message is not valid")
			}

			// Test that decoding works.
			if err := dec.Decode(msg.MessageBody); err != nil {
				t.Fatal("decoding failed:", err)
			}

			// Test the getter.
			if ret := reflect.ValueOf(msg).MethodByName(tt.name).Call([]reflect.Value{}); !ret[1].IsNil() {
				err := ret[1].Interface().(error)
				t.Fatal("returned unexpected error:", err)
			}

			// Same with invalid type
			msg = &Message{MessageBody: struct{}{}}
			if ret := reflect.ValueOf(msg).MethodByName(tt.name).Call([]reflect.Value{}); ret[1].IsNil() {
				t.Fatal("expected interface conversion error wasn't returned")
			}
		})
	}
}

func TestMessage_OtherFixtures(t *testing.T) {
	testCases := []struct {
		name        string
		pathFixture string
		pathSchema  string
		value       interface{}
	}{
		{
			"Message sample",
			"messages/example_message.json",
			"schemas/message/metadata/create_request.json",
			New(MessageTypeEnum_MetadataCreate, MessageClassEnum_Command),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blob := specdata.MustAsset(tc.pathFixture)
			if err := json.Unmarshal(blob, &tc.value); err != nil {
				t.Fatal(err)
			}
			msg, ok := tc.value.(*Message)
			if !ok {
				t.Fatalf("value has not the expected type")
			}
			res, err := getValidator(t).Validate(msg)
			assertResults(t, res, err)
		})
	}
}

func TestMessage_OtherFixtures_Header(t *testing.T) {
	testCases := []struct {
		pathFixtureHeader string
		pathFixtureBody   string
	}{
		{
			"messages/header/metadata_create_error_header.json",
			"messages/body/metadata/create/article_create_request.json",
		},
		{
			"messages/header/metadata_create_header.json",
			"messages/body/metadata/create/article_create_request.json",
		},
		{
			"messages/header/metadata_delete_header.json",
			"messages/body/metadata/delete/research_object_delete_request.json",
		},
		{
			"messages/header/metadata_read_request_header.json",
			"messages/body/metadata/read/research_object_read_request.json",
		},
		{
			"messages/header/metadata_read_response_header.json",
			"messages/body/metadata/read/research_object_read_response.json",
		},
		/*
			{
				"messages/header/metadata_update_header.json",
				"messages/body/metadata/update/research_object_update_request.json",
			},
		*/
		{
			"messages/header/preservation_event_header.json",
			"messages/body/preservation/preservation_event_request.json",
		},
	}
	for _, tt := range testCases {
		name := filepath.Base(tt.pathFixtureHeader)
		t.Run(name, func(t *testing.T) {
			header := specdata.MustAsset(tt.pathFixtureHeader)
			body := specdata.MustAsset(tt.pathFixtureBody)
			blob := []byte(`{
			"messageHeader": ` + string(header) + `,
			"messageBody": ` + string(body) + `
		  }`)
			msg := &Message{}
			if err := json.Unmarshal(blob, msg); err != nil {
				t.Fatal(err)
			}
			res, err := getValidator(t).Validate(msg)
			assertResults(t, res, err)
		})
	}
}

func getValidator(t *testing.T) Validator {
	validator, err := NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	return validator
}

func assertResults(t *testing.T, res *gojsonschema.Result, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !res.Valid() {
		for _, err := range res.Errors() {
			t.Log("validation error:", err)
		}
		t.Error("validator reported that the message is not valid")
	}
}
