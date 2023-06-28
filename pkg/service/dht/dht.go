package dht

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	sdkcrypto "github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/did/key"
	"github.com/ipfs/boxo/ipns"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
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
	advertisePeriod = time.Minute * 30
	peerLimit       = 10
)

type Service struct {
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

func NewService(cfg *config.Config) (*Service, error) {
	var ddt Service
	ddt.cfg = cfg
	storage, err := db.NewStorage(cfg.ServerConfig.DBFile)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate storage")
		return nil, err
	}
	ddt.storage = storage

	multiaddrString := fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ServerConfig.APIHost, cfg.ServerConfig.ListenPort)

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, err := multiaddr.NewMultiaddr(multiaddrString)
	if err != nil {
		logrus.WithError(err).Error("failed to parse multiaddr")
		return nil, err
	}

	var extMultiAddr multiaddr.Multiaddr
	if cfg.ServerConfig.BroadcastIP == "" {
		logrus.Warn("external IP not defined, Peers might not be able to resolve this node if behind NAT")
	} else {
		// here we're creating the multiaddr that others should use to connect to me
		extMultiAddr, err = multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ServerConfig.BroadcastIP, cfg.ServerConfig.ListenPort))
		if err != nil {
			logrus.WithError(err).Error("error creating multiaddress")
			return nil, err
		}
	}
	addressFactory := func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		if extMultiAddr != nil {
			// append the external facing multiaddr we created above to the addressFactory so it will be broadcast
			// out when connecting to a bootstrap node.
			addrs = append(addrs, extMultiAddr)
		}
		return addrs
	}

	// get or create a new service identity
	privKey, err := ddt.setupServiceIdentity()
	if err != nil {
		logrus.WithError(err).Error("failed to setup service identity")
		return nil, err
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

	// set variable for our external address
	if extMultiAddr != nil {
		ddt.externalAddress = fmt.Sprintf("%s/p2p/%s", extMultiAddr, ddt.host.ID())
	} else {
		ddt.externalAddress = fmt.Sprintf("%s/p2p/%s", multiaddrString, ddt.host.ID())
	}

	ctx := context.Background()

	// set up autonat
	if _, err = autonat.New(h); err != nil {
		logrus.WithError(err).Error("failed to set up autonat")
		return nil, err
	}

	// connect to bootstrap peers
	if len(cfg.DHTConfig.BootstrapPeers) > 0 {
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
	if cfg.DHTConfig.LocalDiscovery {
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

// gets or creates a service identity
func (s *Service) setupServiceIdentity() (*crypto.Ed25519PrivateKey, error) {
	did, gotPrivKey, err := s.storage.ReadIdentity()
	if err != nil {
		logrus.WithError(err).Error("failed to read identity")
		return nil, err
	}
	if did != "" && gotPrivKey != nil {
		logrus.Infof("found existing identity: %s", did)
		return gotPrivKey, nil
	}

	logrus.Info("generating new identity")
	privKey, pubKey, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		logrus.WithError(err).Error("failed to generate key")
		return nil, err
	}
	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		logrus.WithError(err).Error("failed to get raw public key bytes")
		return nil, err
	}
	didKey, err := key.CreateDIDKey(sdkcrypto.Ed25519, pubKeyBytes)
	if err != nil {
		logrus.WithError(err).Error("failed to create did key")
		return nil, err
	}
	logrus.Infof("generated new identity: %s", didKey.String())
	if err = s.storage.WriteIdentity(didKey.String(), *privKey.(*crypto.Ed25519PrivateKey)); err != nil {
		logrus.WithError(err).Error("failed to write identity")
		return nil, err
	}
	return privKey.(*crypto.Ed25519PrivateKey), nil
}

func (s *Service) setupGossipSub(ctx context.Context) error {
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

func (s *Service) setupDHT(ctx context.Context) error {
	validator := record.NamespacedValidator{
		"pk":                      record.PublicKeyValidator{},
		"ipns":                    ipns.Validator{KeyBook: s.host.Peerstore()},
		s.cfg.DHTConfig.Namespace: NewValidator(s.cfg.DHTConfig.Namespace),
	}
	d, err := dht.New(ctx, s.host, dht.Mode(dht.ModeServer), dht.Validator(validator))
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate d service")
		return err
	}
	if err = d.Bootstrap(ctx); err != nil {
		logrus.WithError(err).Error("failed to bootstrap d service")
		return err
	}
	s.host = routedhost.Wrap(s.host, d)
	s.dht = d
	return nil
}

func (s *Service) bootstrapPeers(ctx context.Context) error {
	// connect to bootstrap bootstrapPeers
	logrus.Info("connecting to bootstrap bootstrapPeers")
	var wg sync.WaitGroup

	bootstrapPeers := s.cfg.DHTConfig.BootstrapPeers
	numBootstrapPeers := len(bootstrapPeers)
	for _, peerAddr := range bootstrapPeers {
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
		return errors.New("no bootstrap bootstrapPeers could be connected to")
	}
	return nil
}

func (s *Service) setupPeerDiscovery(ctx context.Context) error {
	// refresh the dht route table
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.dht.RefreshRoutingTable():
	}
	d := routing.NewRoutingDiscovery(s.dht)
	s.discovery = d

	// advertise ourselves
	discutil.Advertise(ctx, d, s.cfg.DHTConfig.Namespace, discovery.TTL(advertisePeriod))

	// connect to peers
	logrus.Info("finding peers")
	peerChan, err := d.FindPeers(ctx, s.cfg.DHTConfig.Namespace, discovery.Limit(peerLimit))
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

func (s *Service) setupLocalDiscovery(ctx context.Context) error {
	ldn := new(localDiscoveryNotifee)
	ldn.PeerChan = make(chan peer.AddrInfo)
	svc := mdns.NewMdnsService(s.host, s.cfg.DHTConfig.Namespace, ldn)
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
