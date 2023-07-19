package dht

import (
	"did-dht/internal/record"
)

// Record is a DHT record, which can also be used in other contexts like for gossiping
type Record struct {
	DID      string `json:"did,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	JWS      string `json:"jws,omitempty"`
}

func (r Record) ToMap() map[string]any {
	return map[string]any{
		"did":      r.DID,
		"endpoint": r.Endpoint,
		"jws":      r.JWS,
	}
}

// ToGossipMessage converts a DHT record to a record message for use in gossiping
func (r Record) ToGossipMessage(id, publisherID, topic, receivedAt string) record.Message {
	return record.Message{
		ID:          id,
		PublisherID: publisherID,
		Topic:       topic,
		Record:      r.ToGossipRecord(),
		ReceivedAt:  receivedAt,
	}
}

// ToGossipRecord converts a DHT record to a gossip record
func (r Record) ToGossipRecord() record.SignedRecord {
	return record.SignedRecord{
		Payload: map[string]interface{}{
			"did":      r.DID,
			"endpoint": r.Endpoint,
		},
		JWS: r.JWS,
	}
}
