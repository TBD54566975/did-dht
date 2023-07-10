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

type Gossiper struct {
	Messages chan *Message
	storage  db.GossipStorage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	topicName string

	// our peer's id and name
	id   peer.ID
	name string
}

func (g *Gossiper) Close() error {
	g.sub.Cancel()
	return g.topic.Close()
}

func (g *Gossiper) GetTopics() []string {
	return g.ps.GetTopics()
}

func (g *Gossiper) Publish(ctx context.Context, msg []byte) error {
	return g.topic.Publish(ctx, msg)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (g *Gossiper) pullMessages() {
	for {
		msg, err := g.sub.Next(g.ctx)
		if err != nil {
			logrus.WithError(err).Warn("failed to read message from topic<%s>, closing...", g.topicName)
			close(g.Messages)
			return
		}

		// make sure we're not the sender
		from := msg.GetFrom()
		if from == g.id {
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

func (g *Gossiper) processMessages() {
	for {
		select {
		case <-g.ctx.Done():
			logrus.Info("context cancelled, closing...")
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
