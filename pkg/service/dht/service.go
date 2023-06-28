package dht

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Service) Start(ctx context.Context) error {
	gossiper, err := StartGossiper(ctx, s.storage, s.gossipSub, s.host.ID(), s.cfg.DHTConfig.Name, s.cfg.DHTConfig.Topic)
	if err != nil {
		logrus.WithError(err).Error("failed to start gossiper")
		return err
	}
	s.gossiper = gossiper
	return nil
}

func (s *Service) Info() (string, string, []peer.ID) {
	return s.host.ID().String(), s.externalAddress, s.gossiper.ListPeers()
}

// PublishRecord publishes the given record to the DHT and gossip sub topic
func (s *Service) PublishRecord(ctx context.Context, msg DDTMessage) error {
	if s.gossiper == nil {
		return errors.New("gossiper not started")
	}
	msg.PublisherID = s.host.ID().String()

	// put the record in the DHT
	recordBytes, err := json.Marshal(msg.Record)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal record")
	}
	if err = s.dht.PutValue(ctx, s.dhtKey(msg.Record.DID), recordBytes); err != nil {
		return errors.WithMessage(err, "failed to put record in DHT")
	}

	// broadcast via gossip sub
	if err = s.gossiper.Publish(ctx, msg); err != nil {
		return errors.WithMessage(err, "failed to publish record via gossip sub")
	}

	return nil
}

// QueryRecord returns the record for the given DID first from local storage, then from the DHT
func (s *Service) QueryRecord(ctx context.Context, did string) (*DDTMessage, error) {
	if s.gossiper == nil {
		return nil, errors.New("gossiper not started")
	}

	// attempt to read from local storage first
	record, err := s.storage.ReadRecord(did)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read record")
	}
	if record != nil {
		return &DDTMessage{
			PublisherID: record.PublisherID,
			Record:      Record(record.Record),
		}, nil
	}

	logrus.Info("record not found locally, querying DHT")
	dhtRecord, err := s.dht.GetValue(ctx, s.dhtKey(did))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get record from DHT")
	}
	var r Record
	if err = json.Unmarshal(dhtRecord, &r); err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal record")
	}
	// TODO(gabe): add publisher info here
	return &DDTMessage{
		Record: r,
	}, nil
}

// ListRecords returns all records stored locally
func (s *Service) ListRecords(_ context.Context) ([]DDTMessage, error) {
	if s.gossiper == nil {
		return nil, errors.New("gossiper not started")
	}

	records, err := s.storage.ListRecords()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to list records")
	}

	var messages []DDTMessage
	for _, record := range records {
		messages = append(messages, DDTMessage{
			PublisherID: record.PublisherID,
			Record:      Record(record.Record),
		})
	}
	return messages, nil
}

func (s *Service) RemoveRecord(_ context.Context, did string) error {
	if s.gossiper == nil {
		return errors.New("gossiper not started")
	}

	// TODO(gabe): when we don't have the record locally, query the DHT using our custom protocol to invalidate the record
	return s.storage.DeleteRecord(did)
}

func (s *Service) dhtKey(did string) string {
	return s.cfg.DHTConfig.Namespace + "/" + did
}
