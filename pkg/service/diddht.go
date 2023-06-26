package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	discutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/host/autonat"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/pkg/db"
)

const (
	advertisePeriod = time.Second * 5
	peerLimit       = 10
)

type DIDDHTService struct {
	cfg     *config.Config
	storage *db.Storage

	externalAddress string

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

func NewDIDDHTService(cfg *config.Config) (*DIDDHTService, error) {
	var ddt DIDDHTService
	ddt.cfg = cfg
	storage, err := db.NewStorage(cfg.DBFile)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate storage")
		return nil, err
	}
	ddt.storage = storage

	// TODO(gabe) use a persistent identity
	// generate an identity for the node
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		logrus.WithError(err).Error("failed to generate key")
		return nil, err
	}

	multiaddrString := fmt.Sprintf("/ip4/%s/tcp/%d", cfg.APIHost, cfg.ListenPort)

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, err := multiaddr.NewMultiaddr(multiaddrString)
	if err != nil {
		logrus.WithError(err).Error("failed to parse multiaddr")
		return nil, err
	}

	var extMultiAddr multiaddr.Multiaddr
	if cfg.BroadcastIP == "" {
		logrus.Warn("external IP not defined, Peers might not be able to resolve this node if behind NAT")
		ddt.externalAddress = fmt.Sprintf("%s/p2p/%s", multiaddrString, ddt.host.ID())
	} else {
		// here we're creating the multiaddr that others should use to connect to me
		extMultiAddr, err = multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.BroadcastIP, cfg.ListenPort))
		if err != nil {
			logrus.WithError(err).Error("error creating multiaddress")
			return nil, err
		}
		ddt.externalAddress = fmt.Sprintf("%s/p2p/%s", extMultiAddr, ddt.host.ID())
	}
	addressFactory := func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		if extMultiAddr != nil {
			// append the external facing multiaddr we created above to the addressFactory so it will be broadcast
			// out when connecting to a bootstrap node.
			addrs = append(addrs, extMultiAddr)
		}
		return addrs
	}

	// create a new libp2p host that listens on a random TCP port
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(sourceMultiAddr.String()),
		libp2p.AddrsFactory(addressFactory),
		libp2p.Identity(privKey),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.ForceReachabilityPublic(),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate libp2p host")
		return nil, err
	}
	ddt.host = h
	logrus.Infof("Host created with id: %s, %q", h.ID(), h.Addrs())
	logrus.Info(h.Addrs())

	ctx := context.Background()

	// set up autonat
	if _, err = autonat.New(h); err != nil {
		logrus.WithError(err).Error("failed to set up autonat")
		return nil, err
	}

	// connect to bootstrap peers
	if len(cfg.BootstrapPeers) > 0 {
		if err = ddt.bootstrapPeers(ctx); err != nil {
			logrus.WithError(err).Error("failed to bootstrap peers")
			return nil, err
		}
	}

	// init dht and associate it with the host
	if err = ddt.setupDHT(ctx); err != nil {
		logrus.WithError(err).Error("failed to set up dht")
		return nil, err
	}

	// create a new PubSub service using the GossipSub router
	if err = ddt.setupGossipSub(ctx); err != nil {
		logrus.WithError(err).Error("failed to set up gossipsub")
		return nil, err
	}

	// if local is set, set up local discovery
	if cfg.LocalDiscovery {
		if err = ddt.setupLocalDiscovery(ctx); err != nil {
			logrus.WithError(err).Error("failed to set up local discovery")
			return nil, err
		}
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

func (s *DIDDHTService) bootstrapPeers(ctx context.Context) error {
	// connect to bootstrap peers
	logrus.Info("connecting to bootstrap peers")
	var wg sync.WaitGroup

	numBootstrapPeers := len(s.cfg.BootstrapPeers)
	for _, peerAddr := range s.cfg.BootstrapPeers {
		peerInfo, err := peer.AddrInfoFromString(peerAddr)
		if err != nil {
			logrus.WithError(err).Errorf("failed to parse bootstrap peer: %s", peerAddr)
			numBootstrapPeers--
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err = s.host.Connect(ctx, *peerInfo); err != nil {
				logrus.WithError(err).Warnf("could not connect to bootstrap peer: %s", peerInfo.String())
				numBootstrapPeers--
			} else {
				logrus.Infof("connection established with bootstrap node: %s", peerInfo.String())
			}
		}()
	}
	wg.Wait()

	if numBootstrapPeers == 0 {
		return errors.New("no bootstrap peers could be connected to")
	}
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
	logrus.Info("finding peers")
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

func (s *DIDDHTService) setupLocalDiscovery(ctx context.Context) error {
	ldn := new(localDiscoveryNotifee)
	ldn.PeerChan = make(chan peer.AddrInfo)
	svc := mdns.NewMdnsService(s.host, s.cfg.Namespace, ldn)
	if err := svc.Start(); err != nil {
		logrus.WithError(err).Error("failed to start mdns service")
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case pi := <-ldn.PeerChan:
				logrus.Infof("found local peer %s", pi.ID)
				if err := s.host.Connect(ctx, pi); err != nil {
					logrus.WithError(err).Errorf("failed to connect to peer %s", pi.ID)
				} else {
					logrus.Infof("connected to peer %s", pi.ID)
				}
			}
		}
	}()
	return nil
}

type localDiscoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (ldn *localDiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	ldn.PeerChan <- pi
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

func (s *DIDDHTService) Info() (string, string, []peer.ID) {
	return s.host.ID().String(), s.externalAddress, s.gossiper.ListPeers()
}

func (s *DIDDHTService) Gossip(ctx context.Context, msg string) error {
	if s.gossiper == nil {
		return errors.New("gossiper not started")
	}
	return s.gossiper.Publish(ctx, msg)
}
