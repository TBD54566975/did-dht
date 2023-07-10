package gossip

import (
	"context"
	"encoding/json"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"

	"did-dht/pkg/db"
)

// Gossiper is a wrapper around a pubsub topic that reads messages from the topic and pushes them onto a channel.
type Gossiper struct {
	Messages chan *Message
	storage  db.GossipStorage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	topicName string

	// our peer's id and name
	peerID   peer.ID
	peerName string
}

// Close closes the gossiper by cancelling the context and closing the topic.
func (g *Gossiper) Close() error {
	g.sub.Cancel()
	return g.topic.Close()
}

// Publish publishes a message to the topic.
func (g *Gossiper) Publish(ctx context.Context, msg []byte) error {
	return g.topic.Publish(ctx, msg)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (g *Gossiper) pullMessages() {
	for {
		msg, err := g.sub.Next(g.ctx)
		if err != nil {
			logrus.WithError(err).Warnf("failed to read message from topic<%s>, closing...", g.topicName)
			close(g.Messages)
			return
		}

		// make sure we're not the sender
		from := msg.GetFrom()
		if from == g.peerID {
			continue
		}

		m := Message{PublisherID: from.String(), Topic: msg.GetTopic(), ReceivedAt: time.Now().Format(time.RFC3339)}
		if err = json.Unmarshal(msg.Data, &m.Record); err != nil {
			logrus.WithError(err).Warn("failed to unmarshal message")
			continue
		}

		// send valid messages to the channel
		g.Messages <- &m
	}
}

// processMessages reads messages from the Messages channel and writes them to the database.
func (g *Gossiper) processMessages() {
	for {
		select {
		case <-g.ctx.Done():
			logrus.Infof("context cancelled, closing message processor for topic<%s>...", g.topicName)
			return
		case msg := <-g.Messages:
			logrus.Infof("Received message from %q: %q", msg.PublisherID, msg.Record)
			if err := g.storage.WriteMessage(db.Message{
				ID:          msg.ID,
				Topic:       msg.Topic,
				PublisherID: msg.PublisherID,
				Record:      msg.Record,
				ReceivedAt:  time.Now().Format(time.RFC3339),
			}); err != nil {
				logrus.WithError(err).Warn("failed to write record")
			}
		}
	}
}
