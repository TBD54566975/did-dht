package dht

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/goccy/go-json"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/internal/record"
	"did-dht/pkg/db"
	"did-dht/pkg/service/gossip"
)

func (s *Service) Start(ctx context.Context, gossipSvc *gossip.Service) error {
	if err := gossipSvc.StartGossiper(ctx, s.cfg.DHTConfig.Topic); err != nil {
		return util.LoggingErrorMsg(err, "failed to start DHT gossiper")
	}
	s.gossipSvc = gossipSvc
	return nil
}

func (s *Service) Info() (string, string, []string, []peer.ID) {
	logrus.Infof("host peers: %v", s.host.Peerstore().Peers())

	// TODO(gabe): move this out
	return s.host.ID().String(), s.externalAddress, s.gossipSvc.GetGossipTopics(), s.host.Peerstore().Peers()
}

// PublishRecord publishes the given record to the DHT and gossip sub topic
func (s *Service) PublishRecord(ctx context.Context, r Record) error {
	if s.cfg.DHTConfig.EnforceSignedMessages && r.JWS == "" {
		return errors.New("message must be signed")
	}
	if s.storage == nil {
		return errors.New("storage not initialized")
	}
	if s.dht == nil {
		return errors.New("dht not initialized")
	}
	if s.gossipSvc == nil {
		return errors.New("gossip svc not started")
	}

	// if the record doesn't have a JWS, sign it with the service's key
	if r.JWS == "" {
		signedRecord, err := record.SignRecordJWS(s.signer, r)
		if err != nil {
			return errors.WithMessage(err, "failed to sign message")
		}
		r.JWS = signedRecord.JWS
	} else {
		// verify the record's signature is correct
		if err := VerifyRecord(ctx, s.resolver, r); err != nil {
			return errors.WithMessage(err, "failed to verify message")
		}
	}

	// put the record in our local storage
	if err := s.storage.WriteRecord(db.Record{
		DID:      r.DID,
		Endpoint: r.Endpoint,
		JWS:      r.JWS,
	}); err != nil {
		return util.LoggingErrorMsg(err, "failed to write record, not publishing to network...")
	}

	// put the record in the DHT
	recordBytes, err := json.Marshal(r)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal record")
	}
	if err = s.dht.PutValue(ctx, s.dhtKey(r.DID), recordBytes); err != nil {
		logrus.WithError(err).Error("failed to put record in DHT")
	}

	// broadcast via gossip sub
	msg.PublisherID = s.host.ID().String()
	if err = s.gossipSvc.Publish(ctx, s.cfg.DHTConfig.Topic, record); err != nil {
		logrus.WithError(err).Error("failed to publish record via gossip sub")
	}

	return nil
}

// QueryRecord returns the record for the given DID first from local storage, then from the DHT
func (s *Service) QueryRecord(ctx context.Context, did string) (*record.Message, error) {
	if s.storage == nil {
		return nil, errors.New("storage not initialized")
	}
	if s.dht == nil {
		return nil, errors.New("dht not initialized")
	}

	// attempt to read from local storage first
	r, err := s.storage.ReadRecord(did)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read record")
	}
	if r != nil {
		return &record.Message{
			ID:          "",
			PublisherID: "",
			Topic:       "",
			Record:      record.SignedRecord{},
			ReceivedAt:  "",
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
	return &record.Message{Record: r}, nil
}

// ListRecords returns all records stored locally
func (s *Service) ListRecords(_ context.Context) ([]Record, error) {
	if s.storage == nil {
		return nil, errors.New("storage not initialized")
	}

	records, err := s.storage.ListRecords()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to list records")
	}

	var messages []Record
	for _, r := range records {
		messages = append(messages, Record{
			DID:      r.DID,
			Endpoint: r.Endpoint,
			JWS:      r.JWS,
		})
	}
	return messages, nil
}

func (s *Service) RemoveRecord(_ context.Context, did string) error {
	if s.gossipSvc == nil {
		return errors.New("gossip service not started")
	}

	// TODO(gabe): when we don't have the record locally, query the DHT using our custom protocol to invalidate the record
	return s.storage.DeleteRecord(did)
}

func (s *Service) dhtKey(did string) string {
	return "/" + s.cfg.DHTConfig.Namespace + "/" + did
}
