package service

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	discutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/pkg/db"
)

const (
	advertisePeriod = time.Minute * 30
	peerLimit       = 10
)

type DIDDHTService struct {
	cfg     config.Config
	storage *db.Storage

	// p2p host
	host host.Host
	// p2p gossip sub router
	gossipSub *pubsub.PubSub
	// p2p dht
	dht *dht.IpfsDHT
	// p2p discovery
	discovery *routing.RoutingDiscovery

	gossiper *Gossiper
}

func NewDIDDHTService(cfg config.Config) (*DIDDHTService, error) {
	var ddt DIDDHTService
	ddt.cfg = cfg
	storage, err := db.NewStorage(cfg.DBFile)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate storage")
		return nil, err
	}
	ddt.storage = storage

	// create a new libp2p host that listens on a random TCP port
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate libp2p host")
		return nil, err
	}
	ddt.host = h
	logrus.Infof("Host created with id: %s", h.ID())
	logrus.Info(h.Addrs())

	ctx := context.Background()
	// create a new PubSub service using the GossipSub router
	if err = ddt.setupGossipSub(ctx); err != nil {
		logrus.WithError(err).Error("failed to set up gossipsub")
		return nil, err
	}

	// init dht and associate it with the host
	if err = ddt.setupDHT(ctx); err != nil {
		logrus.WithError(err).Error("failed to set up dht")
		return nil, err
	}

	// set up peer discovery after refreshing the route table, try connecting to peers
	if err = ddt.setupPeerDiscovery(ctx); err != nil {
		logrus.WithError(err).Error("failed to set up peer discovery")
		return nil, err
	}
	return &ddt, nil
}

func (s *DIDDHTService) setupGossipSub(ctx context.Context) error {
	opts := []pubsub.Option{
		pubsub.WithMessageAuthor(s.host.ID()),
		pubsub.WithPeerExchange(true),
	}
	ps, err := pubsub.NewGossipSub(ctx, s.host, opts...)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate pubsub service")
		return err
	}
	s.gossipSub = ps
	return nil
}

func (s *DIDDHTService) setupDHT(ctx context.Context) error {
	dht, err := dht.New(ctx, s.host, dht.Mode(dht.ModeServer))
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate dht service")
		return err
	}
	if err = dht.Bootstrap(ctx); err != nil {
		logrus.WithError(err).Error("failed to bootstrap dht service")
		return err
	}
	s.host = routedhost.Wrap(s.host, dht)
	s.dht = dht
	return nil
}

func (s *DIDDHTService) setupPeerDiscovery(ctx context.Context) error {
	// refresh the dht route table
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.dht.RefreshRoutingTable():
	}
	d := routing.NewRoutingDiscovery(s.dht)
	s.discovery = d

	// advertise ourselves
	discutil.Advertise(ctx, d, s.cfg.Namespace, discovery.TTL(advertisePeriod))

	// connect to peers
	peerChan, err := d.FindPeers(ctx, s.cfg.Namespace, discovery.Limit(peerLimit))
	if err != nil {
		logrus.WithError(err).Error("failed to find peers")
		return err
	}
	for p := range peerChan {
		p := p
		go func() {
			if err = s.host.Connect(ctx, p); err != nil {
				logrus.WithError(err).Errorf("failed to connect to peer %s", p.ID)
			} else {
				logrus.Infof("connected to peer %s", p.ID)
			}
		}()
	}
	return nil
}

func (s *DIDDHTService) Start(ctx context.Context) error {
	gossiper, err := StartGossiper(ctx, s.storage, s.gossipSub, s.host.ID(), s.cfg.Name, s.cfg.Topic)
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
