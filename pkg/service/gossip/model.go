package gossip

const (
	// TopicBufferSize is the number of incoming messages to buffer for each topic.
	TopicBufferSize = 128
)

type Message struct {
	ID          string         `json:"id,omitempty"`
	Topic       string         `json:"type,omitempty"`
	PublisherID string         `json:"publisherId,omitempty"`
	Record      map[string]any `json:"record,omitempty"`
	ReceivedAt  string         `json:"receivedAt,omitempty"`
}
