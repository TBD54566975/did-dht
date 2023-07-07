package db

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	gossipNamespace = "gossip"
)

type GossipMessage interface {
	ID() string
	Type() string
}

type GossipStorage interface {
	WriteMessage(message GossipMessage, topic string) error
	ReadMessage(topic, id string) (GossipMessage, error)
	ListMessages(topic string) ([]GossipMessage, error)
	DeleteMessage(topic, id string) error
}

func (s *Storage) WriteMessage(message GossipMessage) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal message")
	}
	return s.Write(dhtNamespace, message.ID(), messageBytes)
}

func (s *Storage) ReadMessage(topic, id string) (GossipMessage, error) {
	message, err := s.Read(namespace(topic), id)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read message")
	}
	if len(message) == 0 {
		return nil, nil
	}
	var messageResult GossipMessage
	if err = json.Unmarshal(message, &messageResult); err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal message")
	}
	return messageResult, nil
}

func (s *Storage) ListMessages(topic string) ([]GossipMessage, error) {
	records, err := s.ReadAll(namespace(topic))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to list messages")
	}
	var messages []GossipMessage
	for _, record := range records {
		var message GossipMessage
		if err = json.Unmarshal(record, &message); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal message")
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func namespace(topic string) string {
	return gossipNamespace + "-" + topic
}
