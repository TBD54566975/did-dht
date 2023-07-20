package record

type Message struct {
	ID          string       `json:"peerId,omitempty"`
	PublisherID string       `json:"publisherId,omitempty"`
	Topic       string       `json:"type,omitempty"`
	Record      SignedRecord `json:"record,omitempty"`
	ReceivedAt  string       `json:"receivedAt,omitempty"`
}

type SignedRecord struct {
	Payload map[string]any `json:"payload,omitempty"`
	JWS     string         `json:"jws,omitempty"`
}
