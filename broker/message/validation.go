package message

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v3"
)

// Validator performs validation of incoming messages.
//
// Implementors must look up the version of the message first and perform
// transformation in case of version mismatch, or validation otherwise.
// ValidationError may be returned to share validation issues precisely.
type Validator interface {
	Validate(ctx context.Context, stream []byte) ([]byte, error)
}

type ValidationError struct {
	Errors []ValidationErrorDetail
}

type ValidationErrorDetail struct {
	Message string `json:"message"`
	Path    string `json:"path"`
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("validation issues: %+v", err.Errors)
}

// NoOpValidatorImpl is a no-op validator.
type NoOpValidatorImpl struct{}

var _ Validator = (*NoOpValidatorImpl)(nil)

func NewValidator() (*NoOpValidatorImpl, error) {
	return &NoOpValidatorImpl{}, nil
}

func (v *NoOpValidatorImpl) Validate(ctx context.Context, stream []byte) ([]byte, error) {
	return stream, nil
}

// jiscValidatorImpl is an implementation of Validator that uses Jisc's
// Message Schema Service to perform validation and conversion.
type jiscValidatorImpl struct {
	baseURL   *url.URL
	client    *http.Client
	userAgent string

	// Version of the specification that the adapter supports at this moment.
	version string
}

var _ Validator = (*jiscValidatorImpl)(nil)

func NewJiscValidator(baseURL, userAgent, specVersion string) (*jiscValidatorImpl, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error processing validator URL (%q): %w", baseURL, err)
	}

	validator := &jiscValidatorImpl{
		baseURL:   u,
		userAgent: userAgent,
		version:   specVersion,
	}

	// Custom HTTP client with sane defaults. The API looks pretty slow. I'm
	// seeing response times of ~500ms, so we're going to be forgiving.
	const (
		dialTimeout      = 5 * time.Second
		handshakeTimeout = 5 * time.Second
		timeout          = 10 * time.Second
	)
	validator.client = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: dialTimeout}).DialContext,
			TLSHandshakeTimeout: handshakeTimeout,
		},
	}

	return validator, nil
}

// Validate implements the Validator interface.
func (v *jiscValidatorImpl) Validate(ctx context.Context, stream []byte) ([]byte, error) {
	envelope, err := Open(stream)
	if err != nil {
		return nil, fmt.Errorf("error opening envelope: %v", err)
	}

	// Convert the message via the conversion API when we see a version mismatch.
	if envelope.Attributes.Version != v.version {
		stream, err := v.transformRequest(ctx, stream)
		if err != nil {
			return nil, fmt.Errorf("error transforming message: %w", err)
		}

		return stream, nil
	}

	// Otherwise, perform validation which expects us to provide the schema ID.
	schemaID := envelope.SchemaDefinition()
	if schemaID == "" {
		return nil, fmt.Errorf("error validating message: unexpected type %q", envelope.Attributes.MessageType)
	}
	if err := v.validateRequest(ctx, stream, schemaID); err != nil {
		return nil, fmt.Errorf("error validating message: %w", err)
	}

	return stream, nil
}

// request encodes and delivers the HTTP request with exponential backoff.
func (v *jiscValidatorImpl) request(ctx context.Context, method, urlStr string, requestPayload interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	if requestPayload != nil {
		if err := json.NewEncoder(buf).Encode(requestPayload); err != nil {
			return nil, fmt.Errorf("error encoding the request: %v", err)
		}
	}

	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing the URL string: %v", err)
	}
	dest := v.baseURL.ResolveReference(rel)

	req, err := http.NewRequestWithContext(ctx, method, dest.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	const mediaTypeJSON = "application/json"
	req.Header.Add("Content-Type", mediaTypeJSON)
	req.Header.Add("Accept", mediaTypeJSON)
	req.Header.Add("User-Agent", v.userAgent)

	// We want to retry in case of timeouts or a number of server errors.
	var backoffStrategy = backoff.WithContext(&backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.5,
		MaxInterval:         10 * time.Second,
		MaxElapsedTime:      2 * time.Minute,
		Clock:               backoff.SystemClock,
	}, ctx)

	var resp *http.Response
	err = backoff.Retry(
		func() error {
			resp, err = v.client.Do(req)
			if err != nil {
				return err
			}
			switch {
			// Give up rigth away on client errors.
			case resp.StatusCode >= 401 && resp.StatusCode < 500:
				return backoff.Permanent(
					fmt.Errorf("%s (client error)", http.StatusText(resp.StatusCode)),
				)
			// Retry on server errors.
			case resp.StatusCode >= 500:
				return errors.New("5xx (server error)")
			}
			return nil
		},
		backoffStrategy,
	)

	return resp, err
}

// do delivers the client request and decodes the returned response.
func (v *jiscValidatorImpl) decodeResponse(body io.ReadCloser, responsePayload interface{}) (err error) {
	defer func() {
		if rerr := body.Close(); rerr != nil {
			err = fmt.Errorf("error closing the response body: %v", rerr)
		}
	}()

	if responsePayload == nil {
		return nil
	}

	err = json.NewDecoder(body).Decode(responsePayload)
	if err != nil {
		err = fmt.Errorf("error decoding the response payload: %v", err)
	}

	return err
}

type jiscValidateRequest struct {
	SchemaID string          `json:"schema_id"`
	Element  json.RawMessage `json:"json_element"`
}

type jiscValidateResponse struct {
	Errors      []ValidationErrorDetail `json:"errorList"`
	MessageType string                  `json:"messageType"`
	SchemaID    string                  `json:"schemaId"`
	Valid       bool                    `json:"valid"`
	VersionTag  string                  `json:"versionTag"`
}

func (v *jiscValidatorImpl) validateRequest(ctx context.Context, stream []byte, schemaID string) error {
	path := fmt.Sprintf("schema_validation/%s/", v.version)

	requestPayload := jiscValidateRequest{
		SchemaID: schemaID,
		Element:  stream,
	}
	resp, err := v.request(ctx, "POST", path, &requestPayload)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}

	var responsePayload = &jiscValidateResponse{}
	if err = v.decodeResponse(resp.Body, responsePayload); err != nil {
		return err
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		break
	case resp.StatusCode == http.StatusBadRequest:
		err = ValidationError{Errors: responsePayload.Errors}
	default:
		err = fmt.Errorf("unexpected response status %d", resp.StatusCode)
	}

	return err
}

type jiscTransformRequest struct {
	Element   json.RawMessage `json:"json_element"`
	ToVersion string          `json:"to_version"`
}

type jiscTransformResponseSuccess struct {
	Element     json.RawMessage `json:"json_content"`
	FromVersion string          `json:"from_version_tag"`
	ToVersion   string          `json:"to_version_tag"`
}

type jiscTransformResponseError struct {
	Message string
}

func (v *jiscValidatorImpl) transformRequest(ctx context.Context, stream []byte) ([]byte, error) {
	const path = "schema_conversion/"

	requestPayload := jiscTransformRequest{
		Element:   stream,
		ToVersion: v.version,
	}
	resp, err := v.request(ctx, "POST", path, &requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		responsePayload := jiscTransformResponseError{}
		if err = v.decodeResponse(resp.Body, responsePayload); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error transforming message: %s", responsePayload.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	responsePayload := &jiscTransformResponseSuccess{}
	if err = v.decodeResponse(resp.Body, responsePayload); err != nil {
		return nil, err
	}

	return responsePayload.Element, nil
}
