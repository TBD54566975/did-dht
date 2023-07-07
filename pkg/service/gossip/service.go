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
	cfg     *config.Config
	storage db.GossipStorage
	ps      *pubsub.PubSub
}

func NewGossipService(ctx context.Context, host host.Host, cfg *config.Config, storage db.GossipStorage, peerID peer.ID) (*Service, error) {
	opts := []pubsub.Option{
		pubsub.WithMessageAuthor(host.ID()),
		pubsub.WithPeerExchange(true),
	}
	ps, err := pubsub.NewGossipSub(ctx, host, opts...)
	if err != nil {
		return nil, util.LoggingErrorMsgf(err, "failed to instantiate pubsub service")
	}
	return &Service{
		cfg:     cfg,
		storage: storage,
		ps:      ps,
	}, nil
}
