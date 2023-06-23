package service

import (
	"context"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/pkg/db"
)

type DIDDHTService struct {
	storage *db.Storage
	// p2p host
	host host.Host
	// p2p gossip sub router
	gossipSub *pubsub.PubSub
	// p2p dht
	dht *dht.IpfsDHT
	// p2p discovery
	// discovery *routing.RoutingDiscovery
	gossiper *Gossiper
}

func NewDIDDHTService(cfg config.Config) (*DIDDHTService, error) {
	storage, err := db.NewStorage(cfg.DBFile)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate storage")
		return nil, err
	}

	// create a new libp2p host that listens on a random TCP port
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate libp2p host")
		return nil, err
	}
	logrus.Infof("Host created with id: %s", h.ID())
	logrus.Info(h.Addrs())

	// create a new PubSub service using the GossipSub router
	opts := []pubsub.Option{
		pubsub.WithMessageAuthor(h.ID()),
		pubsub.WithPeerExchange(true),
	}
	ctx := context.Background()
	ps, err := pubsub.NewGossipSub(ctx, h, opts...)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate pubsub service")
		return nil, err
	}

	// init dht and associate it with the host
	dht, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate dht service")
		return nil, err
	}
	if err = dht.Bootstrap(ctx); err != nil {
		logrus.WithError(err).Error("failed to bootstrap dht service")
		return nil, err
	}
	h = routedhost.Wrap(h, dht)

	return &DIDDHTService{
		storage:   storage,
		host:      h,
		gossipSub: ps,
		dht:       dht,
	}, nil
}

func (s *DIDDHTService) Start(ctx context.Context, topic string) error {
	gossiper, err := StartGossiper(ctx, s.storage, s.gossipSub, s.host.ID(), "did-dht-og", topic)
	if err != nil {
		logrus.WithError(err).Error("failed to start gossiper")
		return err
	}
	s.gossiper = gossiper
	return nil
}

func (s *DIDDHTService) Info() (string, []peer.ID) {
	return s.host.ID().String(), s.gossiper.ListPeers()
}

func (s *DIDDHTService) Gossip(ctx context.Context, msg string) error {
	if s.gossiper == nil {
		return errors.New("gossiper not started")
	}
	return s.gossiper.Publish(ctx, msg)
}
