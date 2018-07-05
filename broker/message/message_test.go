package message

import (
	"bytes"
	"encoding/json"
	"errors"
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
	fixture := specdata.MustAsset("messages/message.json")

	// Our message.
	message := &Message{
		MessageHeader: MessageHeader{
			ID:            MustUUID("efac164a-c9bd-45e0-8991-1c505e4f45c2"),
			CorrelationID: MustUUID("4501437a-ce95-4372-be5c-277cb6a826eb"),
			MessageClass:  MessageClassCommand,
			MessageType:   MessageTypeMetadataCreate,
			ReturnAddress: "string",
			MessageTimings: MessageTimings{
				PublishedTimestamp:  Timestamp(time.Date(2004, time.August, 1, 10, 0, 0, 0, time.UTC)),
				ExpirationTimestamp: Timestamp(time.Date(2004, time.August, 1, 10, 0, 0, 0, time.UTC)),
			},
			MessageSequence: MessageSequence{
				Sequence: MustUUID("6ad8194d-d1d0-4389-a64d-c73d761463c9"),
				Position: 1,
				Total:    1,
			},
			MessageHistory: []MessageHistory{
				MessageHistory{
					MachineId:      "string",
					MachineAddress: "machine.example.com",
					Timestamp:      Timestamp(time.Date(2004, time.August, 1, 10, 0, 0, 0, time.UTC)),
				},
			},
			Version:          "1.2.3",
			ErrorCode:        "GENERR001",
			ErrorDescription: "string",
			Generator:        "string",
		},
		MessageBody: &MetadataCreateRequest{
			ResearchObject{
				ObjectUuid:  MustUUID("5680e8e0-28a5-4b20-948e-fd0d08781e0b"),
				ObjectTitle: "string",
				ObjectPersonRole: []PersonRole{
					PersonRole{
						Person: Person{
							PersonUuid: MustUUID("27811a4c-9cb5-4e6d-a069-5c19288fae58"),
							PersonIdentifier: []PersonIdentifier{
								PersonIdentifier{
									PersonIdentifierValue: "string",
									PersonIdentifierType:  PersonIdentifierTypeEnum_ORCID,
								},
							},
							PersonHonorificPrefix: "string",
							PersonGivenNames:      "string",
							PersonFamilyNames:     "string",
							PersonHonorificSuffix: "string",
							PersonMail:            "person@net",
							PersonOrganisationUnit: OrganisationUnit{
								OrganisationUnitUuid: MustUUID("28be7f16-0e70-461f-a2db-d9d7c64a8f17"),
								OrganisationUuidName: "string",
								Organisation: Organisation{
									OrganisationJiscId:  1,
									OrganisationName:    "string",
									OrganisationType:    OrganisationTypeEnum_charity,
									OrganisationAddress: "string",
								},
							},
						},
						Role: PersonRoleEnum_administrator,
					},
				},
				ObjectDescription: "string",
				ObjectRights: Rights{
					RightsStatement: []string{"string"},
					RightsHolder:    []string{"string"},
					Licence: []Licence{
						Licence{
							LicenceName:       "string",
							LicenceIdentifier: "string",
							LicenseStartDate:  Timestamp(time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)),
							LicenseEndDate:    Timestamp(time.Date(2018, time.December, 31, 23, 59, 59, 0, time.UTC)),
						},
					},
					Access: []Access{
						Access{
							AccessType:      AccessTypeEnum_open,
							AccessStatement: "string",
						},
					},
				},
				ObjectDate: []Date{
					Date{
						DateValue: "2002-10-02T10:00:00-05:00",
						DateType:  DateTypeEnum_accepted,
					},
				},
				ObjectKeywords:     []string{"string"},
				ObjectCategory:     []string{"string"},
				ObjectResourceType: ResourceTypeEnum_artDesignItem,
				ObjectValue:        ObjectValueEnum_normal,
				ObjectIdentifier: []Identifier{
					Identifier{
						IdentifierValue: "string",
						IdentifierType:  1,
					},
				},
				ObjectRelatedIdentifier: []IdentifierRelationship{
					IdentifierRelationship{
						Identifier: Identifier{
							IdentifierValue: "string",
							IdentifierType:  IdentifierTypeEnum_ARK,
						},
						RelationType: RelationTypeEnum_cites,
					},
				},
				ObjectOrganisationRole: []OrganisationRole{
					OrganisationRole{
						Organisation: Organisation{
							OrganisationJiscId:  1,
							OrganisationName:    "string",
							OrganisationType:    OrganisationTypeEnum_charity,
							OrganisationAddress: "string",
						},
						Role: OrganisationRoleEnum_funder,
					},
				},
				ObjectPreservationEvent: []PreservationEvent{
					PreservationEvent{
						PreservationEventValue:  "string",
						PreservationEventType:   PreservationEventTypeEnum_capture,
						PreservationEventDetail: "string",
					},
				},
				ObjectFile: []File{
					File{
						FileUUID:        MustUUID("e150c4ab-0370-4e5a-8722-7fb3369b7017"),
						FileIdentifier:  "string",
						FileName:        "string",
						FileSize:        1,
						FileLabel:       "string",
						FileDateCreated: Timestamp(time.Date(2002, time.October, 2, 10, 0, 0, 0, time.FixedZone("", -18000))),
						FileRights: Rights{
							RightsStatement: []string{"string"},
							RightsHolder:    []string{"string"},
							Licence: []Licence{
								Licence{
									LicenceName:       "string",
									LicenceIdentifier: "string",
									LicenseStartDate:  Timestamp(time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)),
									LicenseEndDate:    Timestamp(time.Date(2018, time.December, 31, 23, 59, 59, 0, time.UTC)),
								},
							},
							Access: []Access{
								Access{
									AccessType:      AccessTypeEnum_open,
									AccessStatement: "string",
								},
							},
						},
						FileChecksum: []Checksum{
							Checksum{
								ChecksumUuid:  MustUUID("df23b46b-6b64-4a40-842f-5ad363bb6e11"),
								ChecksumType:  ChecksumTypeEnum_md5,
								ChecksumValue: "string",
							},
						},
						FileFormatType:       "string",
						FileCompositionLevel: "string",
						FileHasMimeType:      true,
						FileDateModified: []Timestamp{
							Timestamp(time.Date(2002, time.October, 2, 10, 0, 0, 0, time.FixedZone("", -18000))),
						},
						FilePuid: []string{"string"},
						FileUse:  FileUseEnum_originalFile,
						FilePreservationEvent: []PreservationEvent{
							PreservationEvent{
								PreservationEventValue:  "string",
								PreservationEventType:   PreservationEventTypeEnum_capture,
								PreservationEventDetail: "string",
							},
						},
						FileUploadStatus:        UploadStatusEnum_uploadStarted,
						FileStorageStatus:       StorageStatusEnum_online,
						FileLastDownload:        Timestamp(time.Date(2002, time.October, 2, 10, 0, 0, 0, time.FixedZone("", -18000))),
						FileTechnicalAttributes: []string{"string"},
						FileStorageLocation:     "https://tools.ietf.org/html/rfc3986",
						FileStoragePlatform: FileStoragePlatform{
							StoragePlatformUuid: MustUUID("f2939501-2b2d-4e5c-9197-0daa57ccb621"),
							StoragePlatformName: "string",
							StoragePlatformType: StorageTypeEnum_S3,
							StoragePlatformCost: "string",
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
	}
}

func TestMessage_New(t *testing.T) {
	msg := New(MessageTypeMetadataCreate, MessageClassCommand)
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
		MessageBody:   typedBody(MessageTypeVocabularyRead, nil),
	}
	if have, want := m.ID(), m.MessageHeader.ID.String(); have != want {
		t.Errorf("Unexpected ID; have %v, want %v", have, want)
	}
}

func TestMessage_TagError(t *testing.T) {
	m := New(MessageTypeMetadataCreate, MessageClassCommand)
	if m.TagError(nil); m.MessageHeader.ErrorCode != "" || m.MessageHeader.ErrorDescription != "" {
		t.Error("m.TagError(nil): unexpected error headers")
	}

	m = New(MessageTypeMetadataCreate, MessageClassCommand)
	if m.TagError(errors.New("foobar")); m.MessageHeader.ErrorCode != "Unknown" || m.MessageHeader.ErrorDescription != "foobar" {
		t.Error("m.TagError(errors.New('foobar')): unexpected error headers")
	}

	m = New(MessageTypeMetadataCreate, MessageClassCommand)
	if m.TagError(bErrors.New(bErrors.GENERR001, "foobar")); m.MessageHeader.ErrorCode != "GENERR001" || m.MessageHeader.ErrorDescription != "foobar" {
		t.Error("m.TagError(errors.New('foobar')): unexpected error headers")
	}
}

func TestMessage_typedBody(t *testing.T) {
	tests := []struct {
		t             MessageType
		correlationID *UUID
		want          interface{}
	}{
		{MessageTypeMetadataCreate, nil, new(MetadataCreateRequest)},
		{MessageTypeMetadataRead, nil, new(MetadataReadRequest)},
		{MessageTypeMetadataRead, NewUUID(), new(MetadataReadResponse)},
		{MessageTypeMetadataUpdate, nil, new(MetadataUpdateRequest)},
		{MessageTypeMetadataDelete, nil, new(MetadataDeleteRequest)},
		{MessageTypeVocabularyRead, nil, new(VocabularyReadRequest)},
		{MessageTypeVocabularyRead, NewUUID(), new(VocabularyReadResponse)},
		{MessageTypeVocabularyPatch, nil, new(VocabularyPatchRequest)},
		{MessageType(-1), nil, nil},
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
	t           MessageType
	c           MessageClass
	isResponse  bool
}{
	{"MetadataCreateRequest", "messages/body/metadata/create/request.json", MessageTypeMetadataCreate, MessageClassCommand, false},
	{"MetadataDeleteRequest", "messages/body/metadata/delete/request.json", MessageTypeMetadataDelete, MessageClassCommand, false},
	{"MetadataReadRequest", "messages/body/metadata/read/request.json", MessageTypeMetadataRead, MessageClassCommand, false},
	{"MetadataReadResponse", "messages/body/metadata/read/response.json", MessageTypeMetadataRead, MessageClassCommand, true},
	{"MetadataUpdateRequest", "messages/body/metadata/update/request.json", MessageTypeMetadataUpdate, MessageClassCommand, false},
	{"VocabularyPatchRequest", "messages/body/vocabulary/patch/request.json", MessageTypeVocabularyPatch, MessageClassCommand, false},
	{"VocabularyReadRequest", "messages/body/vocabulary/read/request.json", MessageTypeVocabularyRead, MessageClassCommand, false},
	{"VocabularyReadResponse", "messages/body/vocabulary/read/response.json", MessageTypeVocabularyRead, MessageClassCommand, true},
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
			"messages/message.json",
			"messages/message_schema.json",
			New(MessageTypeMetadataCreate, MessageClassCommand),
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
	header := specdata.MustAsset("messages/header/header.json")
	body := specdata.MustAsset("messages/body/metadata/create/request.json")
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
