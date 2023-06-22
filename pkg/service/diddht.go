package service

import (
	"context"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/pkg/db"
)

type DIDDHTService struct {
	storage *db.Storage
	h       *host.Host
	ps      *pubsub.PubSub
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

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(context.Background(), h)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate pubsub service")
		return nil, err
	}
	return &DIDDHTService{
		storage: storage,
		h:       &h,
		ps:      ps,
	}, nil
}
