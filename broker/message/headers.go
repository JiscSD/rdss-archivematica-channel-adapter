package message

// MessageHeader contains important metadata describing the Message itself,
// including the type of Message, routing information, timings, sequencing, and
// so forth.
type MessageHeader struct {
	ID               *UUID            `json:"messageId"`
	CorrelationID    *UUID            `json:"correlationId,omitempty"`
	MessageClass     MessageClassEnum `json:"messageClass"`
	MessageType      MessageTypeEnum  `json:"messageType"`
	ReturnAddress    string           `json:"returnAddress,omitempty"`
	MessageTimings   MessageTimings   `json:"messageTimings"`
	MessageSequence  MessageSequence  `json:"messageSequence"`
	MessageHistory   []MessageHistory `json:"messageHistory,omitempty"`
	Version          string           `json:"version"`
	ErrorCode        string           `json:"errorCode,omitempty"`
	ErrorDescription string           `json:"errorDescription,omitempty"`
	Generator        string           `json:"generator"`
	TenantJiscID     uint64           `json:"tenantJiscID"`
}

type MessageTimings struct {
	PublishedTimestamp  Timestamp `json:"publishedTimestamp"`
	ExpirationTimestamp Timestamp `json:"expirationTimestamp,omitempty"`
}

type MessageSequence struct {
	Sequence *UUID `json:"sequence"`
	Position int   `json:"position"`
	Total    int   `json:"total"`
}

type MessageHistory struct {
	MachineID      string    `json:"machineId"`
	MachineAddress string    `json:"machineAddress"`
	Timestamp      Timestamp `json:"timestamp"`
}
