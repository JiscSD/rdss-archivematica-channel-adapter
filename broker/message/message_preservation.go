package message

import (
	"fmt"
)

// Preservation Event

// PreservationEventRequest represents the body of the message.
type PreservationEventRequest struct {
	InformationPackage
}

// PreservationEventRequest returns the body of the message.
func (m Message) PreservationEventRequest() (*PreservationEventRequest, error) {
	b, ok := m.MessageBody.(*PreservationEventRequest)
	if !ok {
		return nil, fmt.Errorf("PreservationEventRequest(): interface conversion error")
	}
	return b, nil
}
