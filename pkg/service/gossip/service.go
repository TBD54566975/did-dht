package gossip

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/util"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"did-dht/config"
	"did-dht/pkg/db"
)

type Service struct {
	cfg       *config.Config
	peerID    peer.ID
	storage   db.GossipStorage
	ps        *pubsub.PubSub
	gossipers map[string]Gossiper
}

// NewGossipService creates a new gossip service. Able to create multiple gossipers for different topics.
func NewGossipService(ctx context.Context, cfg *config.Config, storage db.GossipStorage, host host.Host) (*Service, error) {
	opts := []pubsub.Option{
		pubsub.WithMessageAuthor(host.ID()),
		pubsub.WithPeerExchange(true),
	}
	ps, err := pubsub.NewGossipSub(ctx, host, opts...)
	if err != nil {
		return nil, util.LoggingErrorMsgf(err, "failed to instantiate pubsub service")
	}
	return &Service{
		cfg:       cfg,
		peerID:    host.ID(),
		storage:   storage,
		ps:        ps,
		gossipers: make(map[string]Gossiper),
	}, nil
}

// StartGossiper creates a new gossiper for the given topic and starts listening for messages
func (s *Service) StartGossiper(ctx context.Context, topic string) error {
	if _, ok := s.gossipers[topic]; ok {
		return util.LoggingNewErrorf("gossiper<%s> already exists", topic)
	}

	// join the topic
	t, err := s.ps.Join(topic)
	if err != nil {
		return util.LoggingErrorMsgf(err, "failed to join topic: %s", topic)
	}

	// subscribe to it
	sub, err := t.Subscribe()
	if err != nil {
		return util.LoggingErrorMsgf(err, "failed to subscribe to topic: %s", topic)
	}

	ddt := &Gossiper{
		Messages: make(chan *Message, TopicBufferSize),
		storage:  s.storage,

		ctx:   ctx,
		ps:    s.ps,
		topic: t,
		sub:   sub,

		topicName: topic,

		peerID:   s.peerID,
		peerName: s.cfg.DHTConfig.Name,
	}

	// start reading messages from the topic in a loop
	go ddt.pullMessages()

	// start processing messages from the topic in a loop
	go ddt.processMessages()

	// add gossiper to the map
	s.gossipers[topic] = *ddt
	return nil
}

// StopGossiper stops the gossiper for the given topic
func (s *Service) StopGossiper(topic string) error {
	g, ok := s.gossipers[topic]
	if !ok {
		return util.LoggingNewErrorf("gossiper<%s> does not exist", topic)
	}
	if err := g.Close(); err != nil {
		return util.LoggingErrorMsgf(err, "failed to close gossiper<%s>", topic)
	}
	delete(s.gossipers, topic)
	return nil
}

// Publish publishes the given message to the given topic
func (s *Service) Publish(ctx context.Context, topic string, msg []byte) error {
	g, ok := s.gossipers[topic]
	if !ok {
		return util.LoggingNewErrorf("gossiper<%s> does not exist", topic)
	}
	return g.Publish(ctx, msg)
}

// GetGossipTopics returns the list of topics that the service is currently gossiping on
func (s *Service) GetGossipTopics() []string {
	var topics []string
	for _, g := range s.gossipers {
		topics = append(topics, g.topicName)
	}
	return topics
}

// ListMessagesForTopic returns the list of messages for the given topic
// TODO(gabe): pagination
func (s *Service) ListMessagesForTopic(topic string) ([]Message, error) {
	g, ok := s.gossipers[topic]
	if !ok {
		return nil, util.LoggingNewErrorf("gossiper<%s> does not exist", topic)
	}
	messages, err := g.storage.ListMessages(topic)
	if err != nil {
		return nil, util.LoggingErrorMsgf(err, "failed to list messages for topic: %s", topic)
	}

	msgs := make([]Message, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, Message{
			ID:          m.ID,
			Topic:       m.Topic,
			PublisherID: m.PublisherID,
			Record:      m.Record,
			ReceivedAt:  m.ReceivedAt,
		})
	}
	return msgs, nil
}
