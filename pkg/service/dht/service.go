package dht

import (
	"context"
	"encoding/json"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/pkg/db"
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

func (s *Service) Info() (string, []string, []string, []peer.ID) {
	return s.host.ID().String(), s.externalAddresses, s.gossiper.GetTopics(), s.host.Network().Peers()
}

// PublishRecord publishes the given record to the DHT and gossip sub topic
func (s *Service) PublishRecord(ctx context.Context, msg DDTMessage) error {
	if s.cfg.DHTConfig.EnforceSignedMessages && msg.Record.JWS == "" {
		return errors.New("message must be signed")
	}
	if s.storage == nil {
		return errors.New("storage not initialized")
	}
	if s.dht == nil {
		return errors.New("dht not initialized")
	}
	if s.gossiper == nil {
		return errors.New("gossiper not started")
	}
	msg.PublisherID = s.host.ID().String()

	// if the record doesn't have a JWS, sign it with the service's key
	if msg.Record.JWS == "" {
		signedRecord, err := SignRecordJWS(s.signer, msg.Record)
		if err != nil {
			return errors.WithMessage(err, "failed to sign message")
		}
		msg.Record.JWS = signedRecord.JWS
	}

	// verify the record's signature is correct
	if err := VerifyRecord(ctx, s.resolver, msg.Record); err != nil {
		return errors.WithMessage(err, "failed to verify message")
	}

	// put the record in our local storage
	if err := s.storage.WriteRecord(db.DDTRecord{
		PublisherID: s.host.ID().String(),
		Record: db.Record(Record{
			DID:      msg.Record.DID,
			Endpoint: msg.Record.Endpoint,
			JWS:      msg.Record.JWS,
		}),
	}); err != nil {
		return util.LoggingErrorMsg(err, "failed to write record, not publishing to network...")
	}

	// put the record in the DHT
	recordBytes, err := json.Marshal(msg.Record)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal record")
	}
	if err = s.dht.PutValue(ctx, s.dhtKey(msg.Record.DID), recordBytes); err != nil {
		return errors.WithMessage(err, "failed to put record in DHT")
	}

	// broadcast via gossip sub
	if err = s.gossiper.Publish(ctx, recordBytes); err != nil {
		return errors.WithMessage(err, "failed to publish record via gossip sub")
	}

	return nil
}

// QueryRecord returns the record for the given DID first from local storage, then from the DHT
func (s *Service) QueryRecord(ctx context.Context, did string) (*DDTMessage, error) {
	if s.storage == nil {
		return nil, errors.New("storage not initialized")
	}
	if s.dht == nil {
		return nil, errors.New("dht not initialized")
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
	if s.storage == nil {
		return nil, errors.New("storage not initialized")
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
	return "/" + s.cfg.DHTConfig.Namespace + "/" + did
}
