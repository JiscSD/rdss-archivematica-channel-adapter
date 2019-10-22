package message

import (
	"context"
)

// Validator performs validation of incoming messages.
//
// Implementors must look up the version of the message first and perform
// transformation in case of version mismatch, or validation otherwise.
// ValidationError may be returned to share validation issues precisely.
type Validator interface {
	Validate(ctx context.Context, stream []byte) ([]byte, error)
}

type ValidationError struct{}

func (err ValidationError) Error() string {
	return ""
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
