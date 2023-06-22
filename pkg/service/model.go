package service

import (
	"context"
	"encoding/json"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
)

const (
	// TopicBufferSize is the number of incoming messages to buffer for each topic.
	TopicBufferSize = 128
)

type DIDDHTTopic struct {
	Messages chan *DIDDHTMessage

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
	Message    string
	SenderID   string
	SenderName string
}

func JoinTopic(ctx context.Context, ps *pubsub.PubSub, id peer.ID, name, topic string) (*DIDDHTTopic, error) {
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

	ddt := &DIDDHTTopic{
		Messages: make(chan *DIDDHTMessage, TopicBufferSize),

		ctx:   ctx,
		ps:    ps,
		topic: t,
		sub:   sub,

		topicName: topic,
		name:      name,
		id:        id,
	}

	// start reading messages from the topic in a loop
	go ddt.readLoop()

	return ddt, nil
}

func (ddt *DIDDHTTopic) Close() {
	ddt.sub.Cancel()
}

func (ddt *DIDDHTTopic) ListPeers() []peer.ID {
	return ddt.ps.ListPeers(ddt.topicName)
}

func (ddt *DIDDHTTopic) Publish(msg string) error {
	m := &DIDDHTMessage{
		Message:    msg,
		SenderID:   ddt.id.String(),
		SenderName: ddt.name,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ddt.topic.Publish(ddt.ctx, msgBytes)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (ddt *DIDDHTTopic) readLoop() {
	for {
		msg, err := ddt.sub.Next(ddt.ctx)
		if err != nil {
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
