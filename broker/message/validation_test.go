package message_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	userAgent = "test"
)

func TestValidatorErrorsWithMalformedEnvelope(t *testing.T) {
	t.Parallel()

	svc, err := message.NewJiscValidator("", userAgent, "4.0.0")
	require.NoError(t, err)

	res, err := svc.Validate(context.Background(), []byte(`{true: false}`))
	assert.Empty(t, res)
	require.EqualError(t, err, "error opening envelope: error decoding header/body streams: invalid character 't' looking for beginning of object key string")

	res, err = svc.Validate(context.Background(), []byte(`{"messageHeader": {}}`))
	assert.Empty(t, res)
	require.EqualError(t, err, "error opening envelope: error decoding envelope attributes: version header is empty or missing")

	res, err = svc.Validate(context.Background(), []byte(`{"messageHeader": {"version": "5.0.0"}}`))
	assert.Empty(t, res)
	require.EqualError(t, err, "error opening envelope: error decoding envelope attributes: message type header is empty or missing")
}

func TestValidation(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		// To provision the test.
		messageType    string
		messageVersion string
		correlationID  string
		respStatus     int
		respPayload    string

		// To evalualuate the results.
		wantedMethod  string
		wantedPath    string
		wantedPayload string
		wantedErr     string
		wantedResult  []byte
	}{
		"Unknown type causes an error": {
			messageType:    "UnknownType",
			messageVersion: "4.0.0",
			wantedErr:      `error validating message: unexpected type "UnknownType"`,
			wantedResult:   nil,
		},
		"4xx status returns error": {
			messageType:    "MetadataCreate",
			messageVersion: "4.0.0",
			respStatus:     http.StatusForbidden,
			respPayload:    "{}",
			wantedMethod:   "POST",
			wantedPath:     "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/metadata/create_request.json/#/definitions/MetadataCreateRequest",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "MetadataCreate"
					},
					"messageBody": {}
				}
			}`,
			wantedErr:    "error validating message: error sending request: Forbidden (client error)",
			wantedResult: nil,
		},
		"Validation issues are returned": {
			messageType:    "MetadataCreate",
			messageVersion: "4.0.0",
			respStatus:     http.StatusBadRequest,
			respPayload: `
			{
				"errorList": [
					{
						"message": "'foobar' is a required property",
						"path": "messageHeader"
					}
				],
				"messageType": "MetadataCreate",
				"schemaId": "https://www.jisc.ac.uk/rdss/schema/message/metadata/create_request.json/#/definitions/MetadataCreateRequest",
				"valid": false,
				"versionTag": "4.0.0"
			}`,
			wantedMethod: "POST",
			wantedPath:   "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/metadata/create_request.json/#/definitions/MetadataCreateRequest",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "MetadataCreate"
					},
					"messageBody": {}
				}
			}`,
			wantedErr:    "error validating message: validation issues: [{Message:'foobar' is a required property Path:messageHeader}]",
			wantedResult: nil,
		},
		"Known type MetadataCreate validates": {
			messageType:    "MetadataCreate",
			messageVersion: "4.0.0",
			respStatus:     http.StatusOK,
			respPayload:    "{}",
			wantedMethod:   "POST",
			wantedPath:     "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/metadata/create_request.json/#/definitions/MetadataCreateRequest",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "MetadataCreate"
					},
					"messageBody": {}
				}
			}`,
			wantedResult: genMessageReqResp("4.0.0", "MetadataCreate", ""),
		},
		"Known type MetadataUpdate validates": {
			messageType:    "MetadataUpdate",
			messageVersion: "4.0.0",
			respStatus:     http.StatusOK,
			respPayload:    "{}",
			wantedMethod:   "POST",
			wantedPath:     "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/metadata/update_request.json/#/definitions/MetadataUpdateRequest",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "MetadataUpdate"
					},
					"messageBody": {}
				}
			}`,
			wantedResult: genMessageReqResp("4.0.0", "MetadataUpdate", ""),
		},
		"Known type MetadataDelete validates": {
			messageType:    "MetadataDelete",
			messageVersion: "4.0.0",
			respStatus:     http.StatusOK,
			respPayload:    "{}",
			wantedMethod:   "POST",
			wantedPath:     "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/metadata/delete_request.json/#/definitions/MetadataDeleteRequest",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "MetadataDelete"
					},
					"messageBody": {}
				}
			}`,
			wantedResult: genMessageReqResp("4.0.0", "MetadataDelete", ""),
		},
		"Known type MetadataRead validates": {
			messageType:    "MetadataRead",
			messageVersion: "4.0.0",
			respStatus:     http.StatusOK,
			respPayload:    "{}",
			wantedMethod:   "POST",
			wantedPath:     "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/metadata/read_request.json/#/definitions/MetadataReadRequest",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "MetadataRead"
					},
					"messageBody": {}
				}
			}`,
			wantedResult: genMessageReqResp("4.0.0", "MetadataRead", ""),
		},
		"Known type MetadataRead (Response) validates": {
			messageType:    "MetadataRead",
			correlationID:  "12345",
			messageVersion: "4.0.0",
			respStatus:     http.StatusOK,
			respPayload:    "{}",
			wantedMethod:   "POST",
			wantedPath:     "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/metadata/read_response.json/#/definitions/MetadataReadResponse",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"correlationId": "12345",
						"messageType": "MetadataRead"
					},
					"messageBody": {}
				}
			}`,
			wantedResult: genMessageReqResp("4.0.0", "MetadataRead", "12345"),
		},
		"Known type PreservationEvent validates": {
			messageType:    "PreservationEvent",
			messageVersion: "4.0.0",
			respStatus:     http.StatusOK,
			respPayload:    "{}",
			wantedMethod:   "POST",
			wantedPath:     "/schema_validation/4.0.0/",
			wantedPayload: `
			{
				"schema_id": "https://www.jisc.ac.uk/rdss/schema/message/preservation/preservation_event_request.json/#/definitions/PreservationEventRequest",
				"json_element": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "PreservationEvent"
					},
					"messageBody": {}
				}
			}`,
			wantedResult: genMessageReqResp("4.0.0", "PreservationEvent", ""),
		},
		"Transformation performed when version is different": {
			messageType:    "MetadataCreate",
			messageVersion: "3.0.0", // Old message!
			respStatus:     http.StatusOK,
			respPayload: `
			{
				"from_version_tag": "3.0.0",
				"to_version_tag": "4.0.0",
				"json_content": {
					"messageHeader": {
						"version": "4.0.0",
						"messageType": "MetadataCreate"
					},
					"messageBody": {
						"converted": true
					}
				}
			}`,
			wantedMethod: "POST",
			wantedPath:   "/schema_conversion/",
			wantedPayload: `
			{
				"to_version": "4.0.0",
				"json_element": {
					"messageHeader": {
						"version": "3.0.0",
						"messageType": "MetadataCreate"
					},
					"messageBody": {}
				}
			}`,
			wantedResult: []byte(`
			{
				"messageHeader": {
					"version": "4.0.0",
					"messageType": "MetadataCreate"
				},
				"messageBody": {
					"converted": true
				}
			}`),
		},
	}
	for name, tc := range tests {
		tc := tc // Capture range variable.
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Server for end-to-end HTTP tests.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, tc.wantedMethod)
				assert.Equal(t, r.URL.Path, tc.wantedPath)

				payload, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)
				r.Body.Close()
				assert.JSONEq(t, string(payload), tc.wantedPayload)

				w.WriteHeader(tc.respStatus)
				fmt.Fprintf(w, tc.respPayload)
			}))
			defer server.Close()

			// Set up validator.
			const versionSupported = "4.0.0"
			addr, _ := url.Parse(server.URL)
			valsvc, err := message.NewJiscValidator(addr.String(), userAgent, versionSupported)
			require.NoError(t, err)

			// Perform validation/transformation.
			result, err := valsvc.Validate(
				context.Background(),
				genMessageReqResp(tc.messageVersion, tc.messageType, tc.correlationID),
			)

			if tc.wantedResult != nil {
				assert.JSONEq(t, string(result), string(tc.wantedResult))
			} else {
				assert.Nil(t, result)
			}

			if tc.wantedErr != "" {
				assert.EqualError(t, err, tc.wantedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func genMessageReqResp(version, messageType, correlationId string) []byte {
	var correlationProp string
	if correlationId != "" {
		correlationProp = fmt.Sprintf(",\n\"correlationId\": \"%s\"", correlationId)
	}

	var message = fmt.Sprintf(`{
	"messageHeader": {
		"version": "%s",
		"messageType": "%s"%s
	},
	"messageBody": {}
}`, version, messageType, correlationProp)

	return []byte(message)
}
