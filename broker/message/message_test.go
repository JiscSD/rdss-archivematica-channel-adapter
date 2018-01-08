package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/twinj/uuid"

	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message/specdata"
)

const (
	metadataCreateRequest = `{
  "messageHeader": {
    "messageId": "4e5bef43-21b1-4ef4-b850-dbae05b4882d",
    "messageClass": "Command",
    "messageType": "MetadataCreate",
    "messageTimings": {
      "publishedTimestamp": null,
      "expirationTimestamp": null
    },
    "messageSequence": {
      "sequence": "6ad8194d-d1d0-4389-a64d-c73d761463c9",
      "position": 0,
      "total": 0
    },
    "messageHistory": [
      {
        "machineId": "foo",
        "machineAddress": "bar",
        "timestamp": "1997-07-16T19:20:00+01:00"
      }
    ],
    "version": "1.3.0"
  },
  "messageBody": {
    "objectUuid": "be8eff14-a92b-429e-80b8-0ec4594d72c0",
    "objectTitle": "Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files",
    "objectPersonRole": [
      {
        "person": {
          "personUuid": "8468f86b-a936-41b3-a8a7-ef37e3008ba8",
          "personIdentifier": null,
          "personEntitlement": null,
          "personOrganisation": null,
          "personAffiliation": null,
          "personGivenName": "Zhang, Shiyu",
          "personCn": "",
          "personSn": "",
          "personTelephoneNumber": "",
          "personMail": "",
          "personOu": ""
        },
        "role": 5
      }
    ],
    "objectDescription": "The fileset contains simulation files of the non-uniform meshed patch antennas (using FDTD Empire XCcel). The conductor thickness and conductivity can be adjusted according to the conductive threads.",
    "objectRights": null,
    "objectDate": [
      {
        "dateValue": "2017-03-17",
        "dateType": 10
      }
    ],
    "objectResourceType": 7,
    "objectValue": 0,
    "objectIdentifier": [
      {
        "identifierValue": "10.17028/rd.lboro.4665448.v1",
        "identifierType": 4,
        "relationType": 8
      }
    ],
    "objectFile": [
      {
        "fileUuid": "f8351e4f-66cc-4434-b0f1-54e7038c031a",
        "fileIdentifier": "1",
        "fileName": "woodpigeon-pic.jpg",
        "fileSize": 147004,
        "fileDateCreated": {
          "dateValue": "",
          "dateType": 0
        },
        "fileRights": {
          "rightsStatement": null,
          "rightsHolder": null,
          "licence": null,
          "access": null
        },
        "fileChecksum": [
          {
            "checksumType": 1,
            "checksumValue": "53a64110e067b14394c142c09571bea0"
          }
        ],
        "fileCompositionLevel": "",
        "fileDateModified": null,
        "fileUse": 0,
        "filePreservationEvent": null,
        "fileUploadStatus": 0,
        "fileStorageStatus": 0,
        "fileLastDownloaded": {
          "dateValue": "",
          "dateType": 0
        },
        "fileStorageLocation": "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg",
        "fileStorageType": 1
      },
      {
        "fileUuid": "c23d70ee-cc6b-4698-8d4c-9dcaefb40672",
        "fileIdentifier": "2",
        "fileName": "bird-sounds.mp3",
        "fileSize": 910616,
        "fileDateCreated": {
          "dateValue": "",
          "dateType": 0
        },
        "fileRights": {
          "rightsStatement": null,
          "rightsHolder": null,
          "licence": null,
          "access": null
        },
        "fileChecksum": [
          {
            "checksumType": 1,
            "checksumValue": "92c8ab01cecceb3bf0789c2cd8c7415a"
          }
        ],
        "fileCompositionLevel": "",
        "fileDateModified": null,
        "fileUse": 0,
        "filePreservationEvent": null,
        "fileUploadStatus": 0,
        "fileStorageStatus": 0,
        "fileLastDownloaded": {
          "dateValue": "",
          "dateType": 0
        },
        "fileStorageLocation": "s3://rdss-prod-figshare-0132/bird-sounds.mp3",
        "fileStorageType": 1
      }
    ]
  }
}`
)

func TestMessage_FromJSON(t *testing.T) {
	testCases := []struct {
		data    []byte       // Message
		et      MessageType  // Expected message type
		ec      MessageClass // Expected message class
		ebt     string       // Expected underlying type of body
		wantErr bool
	}{
		{[]byte(metadataCreateRequest), MessageTypeMetadataCreate, MessageClassCommand, "*message.MetadataCreateRequest", false},
		{[]byte(`{"messageHeader": {"messageID": 12345}, "messageBody": {}}`), -1, -1, "", true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.et, tc.ec), func(t *testing.T) {
			msg := &Message{}
			err := json.Unmarshal(tc.data, msg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("json.Unmarshal() did not return an error but one was expected")
				}
				return
			}
			if err != nil {
				t.Fatalf("error unmarshalling the message: %s", err)
			}
			if msg.MessageHeader.MessageType != tc.et {
				t.Errorf("expected=%s received=%s", tc.et, msg.MessageHeader.MessageType)
			}
			if msg.MessageHeader.MessageClass != tc.ec {
				t.Errorf("expected=%s received=%s", tc.ec, msg.MessageHeader.MessageClass)
			}
			it := reflect.TypeOf(msg.MessageBody).String()
			if it != tc.ebt {
				t.Errorf("expected=%s received=%s", it, tc.ebt)
			}
		})
	}
}

func TestMessage_ToJSON(t *testing.T) {
	testCases := []struct {
		data    []byte   // Expected JSON-encoded message
		message *Message // The message we're encoding
		wantErr bool
	}{
		{
			[]byte(metadataCreateRequest),
			&Message{
				MessageHeader: MessageHeader{
					ID:           MustUUID("4e5bef43-21b1-4ef4-b850-dbae05b4882d"),
					MessageClass: MessageClassCommand,
					MessageType:  MessageTypeMetadataCreate,
					MessageHistory: []MessageHistory{
						MessageHistory{
							MachineId:      "foo",
							MachineAddress: "bar",
							Timestamp:      Timestamp(time.Date(1997, time.July, 16, 19, 20, 0, 0, time.FixedZone("+0100", 3600))),
						},
					},
					MessageSequence: MessageSequence{
						Sequence: MustUUID("6ad8194d-d1d0-4389-a64d-c73d761463c9"),
						Position: 0,
						Total:    0,
					},
					Version: Version,
				},
				MessageBody: &MetadataCreateRequest{
					ResearchObject{
						ObjectUuid:  MustUUID("be8eff14-a92b-429e-80b8-0ec4594d72c0"),
						ObjectTitle: "Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files",
						ObjectPersonRole: []PersonRole{
							{
								Person: &Person{
									PersonUuid:      MustUUID("8468f86b-a936-41b3-a8a7-ef37e3008ba8"),
									PersonGivenName: "Zhang, Shiyu",
								},
								Role: PersonRoleEnum_dataCreator,
							},
						},
						ObjectDescription: "The fileset contains simulation files of the non-uniform meshed patch antennas (using FDTD Empire XCcel). The conductor thickness and conductivity can be adjusted according to the conductive threads.",
						ObjectDate: []Date{
							{
								DateType:  10,
								DateValue: "2017-03-17",
							},
						},
						ObjectResourceType: 7,
						ObjectIdentifier: []Identifier{
							{
								IdentifierValue: "10.17028/rd.lboro.4665448.v1",
								IdentifierType:  4,
								RelationType:    8,
							},
						},
						ObjectFile: []File{
							{
								FileUUID:       MustUUID("f8351e4f-66cc-4434-b0f1-54e7038c031a"),
								FileIdentifier: "1",
								FileName:       "woodpigeon-pic.jpg",
								FileSize:       147004,
								FileChecksum: []Checksum{
									{
										ChecksumType:  1,
										ChecksumValue: "53a64110e067b14394c142c09571bea0",
									},
								},
								FileStorageLocation: "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg",
								FileStorageType:     1,
							},
							{
								FileUUID:       MustUUID("c23d70ee-cc6b-4698-8d4c-9dcaefb40672"),
								FileIdentifier: "2",
								FileName:       "bird-sounds.mp3",
								FileSize:       910616,
								FileChecksum: []Checksum{
									{
										ChecksumType:  1,
										ChecksumValue: "92c8ab01cecceb3bf0789c2cd8c7415a",
									},
								},
								FileStorageLocation: "s3://rdss-prod-figshare-0132/bird-sounds.mp3",
								FileStorageType:     1,
							},
						},
					},
				},
			},
			false,
		},
		{nil, &Message{MessageBody: make(chan int)}, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.message.MessageHeader.MessageType, tc.message.MessageHeader.MessageClass), func(t *testing.T) {
			// Encode it and indent it
			data, err := json.Marshal(tc.message)
			if tc.wantErr {
				if err == nil {
					t.Error("json.Marshal() did not return an error but one was expected")
				}
				return
			}
			if err != nil {
				t.Errorf("ToJSON failed %s", err)
			}
			var out bytes.Buffer
			json.Indent(&out, data, "", "  ")

			// Compare them
			have := out.Bytes()
			if !bytes.Equal(have, tc.data) {
				t.Errorf("Unexpected result:\nHAVE: `%s`\nEXPECTED: `%s`", have, tc.data)
			}
		})
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
		MessageBody:   typedBody(MessageTypeVocabularyRead, ""),
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
		correlationID string
		want          interface{}
	}{
		{MessageTypeMetadataCreate, "", new(MetadataCreateRequest)},
		{MessageTypeMetadataRead, "", new(MetadataReadRequest)},
		{MessageTypeMetadataRead, "ID", new(MetadataReadResponse)},
		{MessageTypeMetadataUpdate, "", new(MetadataUpdateRequest)},
		{MessageTypeMetadataDelete, "", new(MetadataDeleteRequest)},
		{MessageTypeVocabularyRead, "", new(VocabularyReadRequest)},
		{MessageTypeVocabularyRead, "ID", new(VocabularyReadResponse)},
		{MessageTypeVocabularyPatch, "", new(VocabularyPatchRequest)},
		{MessageType(-1), "", nil},
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

			var correlationId string
			if tt.isResponse {
				correlationId = "12345"
			}

			body := typedBody(tt.t, correlationId)
			if err := dec.Decode(body); err != nil {
				t.Fatal("decoding failed:", err)
			}

			// Test typed body getter
			msg := &Message{MessageBody: body}
			if ret := reflect.ValueOf(msg).MethodByName(tt.name).Call([]reflect.Value{}); !ret[1].IsNil() {
				err := ret[1].Interface().(error)
				t.Fatal("returned unexpected error:", err)
			}

			// Same with invalid type
			msg = &Message{MessageBody: struct{}{}}
			if ret := reflect.ValueOf(msg).MethodByName(tt.name).Call([]reflect.Value{}); ret[1].IsNil() {
				t.Fatal("expected interface conversion error wasn't returned")
			}

			// Validation test
			t.Skip("See https://github.com/JiscRDSS/rdss-message-api-specification/pull/67")
			{
				msg = New(tt.t, tt.c)
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
			}
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
