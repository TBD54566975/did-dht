package gossip

const (
	// TopicBufferSize is the number of incoming messages to buffer for each topic.
	TopicBufferSize = 128
)

type Message struct {
	ID          string         `json:"peerID,omitempty"`
	Topic       string         `json:"type,omitempty"`
	PublisherID string         `json:"publisherId,omitempty"`
	Record      map[string]any `json:"record,omitempty"`
	ReceivedAt  string         `json:"receivedAt,omitempty"`
}

type Record struct {
	Payload map[string]any `json:"payload,omitempty"`
	JWS     string         `json:"jws,omitempty"`
}
