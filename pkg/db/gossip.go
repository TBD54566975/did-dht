package db

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	gossipNamespace = "gossip"
)

type Message struct {
	ID          string       `json:"id,omitempty"`
	Topic       string       `json:"type,omitempty"`
	PublisherID string       `json:"publisherId,omitempty"`
	Record      SignedRecord `json:"record,omitempty"`
	ReceivedAt  string       `json:"receivedAt,omitempty"`
}

type SignedRecord struct {
	Payload map[string]any `json:"payload,omitempty"`
	JWS     string         `json:"jws,omitempty"`
}

type GossipStorage interface {
	WriteMessage(message Message) error
	ReadMessage(topic, id string) (*Message, error)
	ListMessages(topic string) ([]Message, error)
	DeleteMessage(topic, id string) error
}

func (s *Storage) WriteMessage(message Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal message")
	}
	return s.Write(namespace(message.Topic), message.ID, messageBytes)
}

func (s *Storage) ReadMessage(topic, id string) (*Message, error) {
	message, err := s.Read(namespace(topic), id)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read message")
	}
	if len(message) == 0 {
		return nil, nil
	}
	var m Message
	if err = json.Unmarshal(message, &m); err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal message")
	}
	return &m, nil
}

func (s *Storage) ListMessages(topic string) ([]Message, error) {
	records, err := s.ReadAll(namespace(topic))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to list messages")
	}
	var messages []Message
	for _, record := range records {
		var message Message
		if err = json.Unmarshal(record, &message); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal message")
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (s *Storage) DeleteMessage(topic, id string) error {
	return s.Delete(namespace(topic), id)
}

func namespace(topic string) string {
	return gossipNamespace + "-" + topic
}
