package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/twinj/uuid"

	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
)

const (
	metadataCreateRequest = `{
  "messageHeader": {
    "messageId": "e4bdaea0-4712-4682-8d4b-e6d92b1fe1ac",
    "messageClass": "Command",
    "messageType": "MetadataCreate",
    "messageTimings": {
      "publishedTimestamp": null,
      "expirationTimestamp": null
    },
    "messageSequence": {
      "sequence": "",
      "position": 0,
      "total": 0
    },
    "version": "1.2.1"
  },
  "messageBody": {
    "objectUuid": "06f75186-f9cb-4be3-8df7-64dd037eb54f",
    "objectTitle": "Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files",
    "objectPersonRole": [
      {
        "person": {
          "personUuid": "479dada4-8650-421e-8480-63d58107a998",
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
        "fileUuid": "42143745-dba9-9c98-cdd7-0c521ae55118",
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
        "fileUuid": "9102f719-2f78-d9c2-de12-15f27024f78b",
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
				t.Errorf("error unmarshalling the message: %s", err)
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
					ID:           "e4bdaea0-4712-4682-8d4b-e6d92b1fe1ac",
					MessageClass: MessageClassCommand,
					MessageType:  MessageTypeMetadataCreate,
					Version:      Version,
				},
				MessageBody: &MetadataCreateRequest{
					ResearchObject{
						ObjectUuid:  "06f75186-f9cb-4be3-8df7-64dd037eb54f",
						ObjectTitle: "Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files",
						ObjectPersonRole: []PersonRole{
							{
								Person: &Person{
									PersonUuid:      "479dada4-8650-421e-8480-63d58107a998",
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
								FileUUID:       "42143745-dba9-9c98-cdd7-0c521ae55118",
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
								FileUUID:       "9102f719-2f78-d9c2-de12-15f27024f78b",
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
		MessageHeader: MessageHeader{ID: "ID"},
		MessageBody:   typedBody(MessageTypeVocabularyRead, ""),
	}
	if id := m.ID(); id != "ID" {
		t.Errorf("unexpected ID: %v", id)
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
	{"MetadataCreateRequest", "body/metadata/create/request.json", MessageTypeMetadataCreate, MessageClassCommand, false},
	{"MetadataDeleteRequest", "body/metadata/delete/request.json", MessageTypeMetadataDelete, MessageClassCommand, false},
	{"MetadataReadRequest", "body/metadata/read/request.json", MessageTypeMetadataRead, MessageClassCommand, false},
	{"MetadataReadResponse", "body/metadata/read/response.json", MessageTypeMetadataRead, MessageClassCommand, true},
	{"MetadataUpdateRequest", "body/metadata/update/request.json", MessageTypeMetadataUpdate, MessageClassCommand, false},
	{"VocabularyPatchRequest", "body/vocabulary/patch/request.json", MessageTypeVocabularyPatch, MessageClassCommand, false},
	{"VocabularyReadRequest", "body/vocabulary/read/request.json", MessageTypeVocabularyRead, MessageClassCommand, false},
	{"VocabularyReadResponse", "body/vocabulary/read/response.json", MessageTypeVocabularyRead, MessageClassCommand, true},
}

func TestMessage_DecodeFixtures(t *testing.T) {
	var (
		fixturesDir = "../../fixtures/messages"
		wd, _       = os.Getwd()
	)
	for _, tt := range sharedTests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(wd, fixturesDir, tt.pathFixture)
			in, err := os.Open(path)
			if err != nil {
				t.Fatal("fixture could not be opened:", err)
			}
			dec := json.NewDecoder(in)

			var correlationId string
			if tt.isResponse {
				correlationId = "12345"
			}

			body := typedBody(tt.t, correlationId)
			if err = dec.Decode(body); err != nil {
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
		})
	}
}
