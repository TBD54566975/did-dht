package service

import (
	"context"
	"encoding/json"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"

	"did-dht/pkg/db"
)

const (
	// TopicBufferSize is the number of incoming messages to buffer for each topic.
	TopicBufferSize = 128
)

type Gossiper struct {
	Messages chan *DIDDHTMessage
	storage  db.DDTStorage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	topicName string

	// our id and name
	id   peer.ID
	name string
}

type DIDDHTMessage struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
}

func StartGossiper(ctx context.Context, storage db.DDTStorage, ps *pubsub.PubSub, id peer.ID, name, topic string) (*Gossiper, error) {
	// join the topic
	t, err := ps.Join(topic)
	if err != nil {
		return nil, err
	}

	// subscribe to it
	sub, err := t.Subscribe()
	if err != nil {
		return nil, err
	}

	ddt := &Gossiper{
		Messages: make(chan *DIDDHTMessage, TopicBufferSize),
		storage:  storage,

		ctx:   ctx,
		ps:    ps,
		topic: t,
		sub:   sub,

		topicName: topic,
		name:      name,
		id:        id,
	}

	// start reading messages from the topic in a loop
	go ddt.pullMessages()

	// start processing messages from the topic in a loop
	go ddt.processMessages()

	return ddt, nil
}

func (ddt *Gossiper) Close() error {
	ddt.sub.Cancel()
	return ddt.topic.Close()
}

func (ddt *Gossiper) ListPeers() []peer.ID {
	return ddt.ps.ListPeers(ddt.topicName)
}

func (ddt *Gossiper) Publish(ctx context.Context, msg string) error {
	m := &DIDDHTMessage{
		ID:      ddt.id.String(),
		Name:    ddt.name,
		Message: msg,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ddt.topic.Publish(ctx, msgBytes)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (ddt *Gossiper) pullMessages() {
	for {
		msg, err := ddt.sub.Next(ddt.ctx)
		if err != nil {
			logrus.WithError(err).Warn("failed to read message from topic, closing...")
			close(ddt.Messages)
			return
		}

		// make sure we're not the sender
		if msg.GetFrom() == ddt.id {
			continue
		}

		var m DIDDHTMessage
		if err = json.Unmarshal(msg.Data, &m); err != nil {
			logrus.WithError(err).Warn("failed to unmarshal message")
			continue
		}

		// send valid messages to the channel
		ddt.Messages <- &m
	}
}

func (ddt *Gossiper) processMessages() {
	for {
		select {
		case <-ddt.ctx.Done():
			logrus.Info("context cancelled, closing...")
			return
		case msg := <-ddt.Messages:
			logrus.Infof("Received message from %s: %s", msg.Name, msg.Message)
			if err := ddt.storage.WriteRecord(db.DDTRecord{
				ID:        msg.ID,
				Name:      msg.Name,
				Message:   msg.Message,
				CreatedAt: time.Now().Format(time.RFC3339),
			}); err != nil {
				logrus.WithError(err).Warn("failed to write record")
			}
		}
	}
}
