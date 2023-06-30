package dht

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	sdkcrypto "github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/did/key"
	"github.com/TBD54566975/ssi-sdk/util"
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
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/internal/resolution"
	"did-dht/pkg/db"
)

const (
	advertisePeriod     = time.Minute * 30
	peerDiscoveryPeriod = time.Minute * 5
	peerLimit           = 10
	protocolPrefix      = "/diddht"
)

type Service struct {
	cfg             *config.Config
	externalAddress string
	signer          jwx.Signer
	storage         *db.Storage
	resolver        *resolution.ServiceResolver

	// p2p services

	host      host.Host
	gossipSub *pubsub.PubSub
	dht       *dht.IpfsDHT
	discovery *routing.RoutingDiscovery
	gossiper  *Gossiper
}

func NewService(cfg *config.Config) (*Service, error) {
	var ddt Service
	ddt.cfg = cfg
	storage, err := db.NewStorage(cfg.ServerConfig.DBFile)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate storage")
	}
	ddt.storage = storage

	// create a new resolver
	localResolutionMethods := []string{did.KeyMethod.String(), did.PKHMethod.String(), did.WebMethod.String(), did.JWKMethod.String()}
	ddt.resolver, err = resolution.NewServiceResolver(nil, localResolutionMethods, cfg.DHTConfig.ResolverEndpoint)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate resolver")
	}

	multiaddrString := fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ServerConfig.APIHost, cfg.ServerConfig.ListenPort)

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, err := multiaddr.NewMultiaddr(multiaddrString)
	if err != nil {
		return nil, util.LoggingErrorMsgf(err, "failed to parse multiaddr: %s", multiaddrString)
	}

	var extMultiAddr multiaddr.Multiaddr
	if cfg.ServerConfig.BroadcastIP == "" {
		logrus.Warn("external IP not defined, Peers might not be able to resolve this node if behind NAT")
	} else {
		// here we're creating the multiaddr that others should use to connect to me
		extMultiAddr, err = multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ServerConfig.BroadcastIP, cfg.ServerConfig.ListenPort))
		if err != nil {
			return nil, util.LoggingErrorMsg(err, "failed to create multiaddress")
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
		return nil, util.LoggingErrorMsg(err, "failed to setup service identity")
	}

	// create a new libp2p host that listens on a random TCP port
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(sourceMultiAddr.String()),
		libp2p.AddrsFactory(addressFactory),
		libp2p.Identity(privKey),
		libp2p.EnableNATService(),
		libp2p.ForceReachabilityPublic(),
		libp2p.EnableHolePunching(),
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate libp2p host")
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
		return nil, util.LoggingErrorMsg(err, "failed to set up autonat")
	}

	// set up relay
	if _, err = relay.New(h); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to set up relay")
	}

	// connect to bootstrap peers
	if len(cfg.DHTConfig.BootstrapPeers) > 0 {
		if err = ddt.bootstrapPeers(ctx); err != nil {
			return nil, util.LoggingErrorMsg(err, "failed to bootstrap peers")
		}
	}

	// init dht and associate it with the host
	if err = ddt.setupDHT(ctx); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to set up dht")
	}

	// create a new PubSub service using the GossipSub router
	if err = ddt.setupGossipSub(ctx); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to set up gossipsub")
	}

	// if local is set, set up local discovery
	if cfg.DHTConfig.LocalDiscovery {
		if err = ddt.setupLocalDiscovery(ctx); err != nil {
			return nil, util.LoggingErrorMsg(err, "failed to set up local discovery")
		}
	}

	// set up peer discovery after refreshing the route table, try connecting to peers
	if err = ddt.setupPeerDiscovery(ctx); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to set up peer discovery")
	}

	return &ddt, nil
}

// gets or creates a service identity
func (s *Service) setupServiceIdentity() (*crypto.Ed25519PrivateKey, error) {
	did, gotPrivKey, err := s.storage.ReadIdentity()
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to read identity")
	}

	var privKey *crypto.Ed25519PrivateKey
	var didKey *key.DIDKey
	if did != "" && gotPrivKey != nil {
		logrus.Infof("found existing identity: %s", did)
		privKey = gotPrivKey
		k := key.DIDKey(did)
		didKey = &k
	} else {
		logrus.Info("generating new identity")
		privKey, didKey, err = s.generateNewIdentity()
		if err != nil {
			return nil, util.LoggingErrorMsg(err, "failed to generate new identity")
		}
	}

	// create and store a signer for the key
	expanded, err := didKey.Expand()
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to expand did key")
	}
	privKeyBytes, err := privKey.Raw()
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to get raw private key bytes")
	}
	signer, err := jwx.NewJWXSigner(didKey.String(), expanded.VerificationMethod[0].ID, ed25519.PrivateKey(privKeyBytes))
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to create jwx signer")
	}
	s.signer = *signer

	// return the priv key
	return privKey, nil
}

func (s *Service) generateNewIdentity() (*crypto.Ed25519PrivateKey, *key.DIDKey, error) {
	privKey, pubKey, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, nil, util.LoggingErrorMsg(err, "failed to generate key")
	}
	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		return nil, nil, util.LoggingErrorMsg(err, "failed to get raw public key bytes")
	}
	didKey, err := key.CreateDIDKey(sdkcrypto.Ed25519, pubKeyBytes)
	if err != nil {
		return nil, nil, util.LoggingErrorMsg(err, "failed to create did key")
	}
	logrus.Infof("generated new identity: %s", didKey.String())
	if err = s.storage.WriteIdentity(didKey.String(), *privKey.(*crypto.Ed25519PrivateKey)); err != nil {
		return nil, nil, util.LoggingErrorMsg(err, "failed to write identity")
	}
	return privKey.(*crypto.Ed25519PrivateKey), didKey, nil
}

func (s *Service) setupGossipSub(ctx context.Context) error {
	opts := []pubsub.Option{
		pubsub.WithMessageAuthor(s.host.ID()),
		pubsub.WithPeerExchange(true),
	}
	ps, err := pubsub.NewGossipSub(ctx, s.host, opts...)
	if err != nil {
		return util.LoggingErrorMsgf(err, "failed to instantiate pubsub service")
	}
	s.gossipSub = ps
	return nil
}

func (s *Service) setupDHT(ctx context.Context) error {
	validator := record.NamespacedValidator{
		"pk":                      record.PublicKeyValidator{},
		"ipns":                    ipns.Validator{KeyBook: s.host.Peerstore()},
		s.cfg.DHTConfig.Namespace: NewValidator(s.cfg.DHTConfig.Namespace, s.resolver),
	}
	d, err := dht.New(
		ctx,
		s.host,
		dht.Mode(dht.ModeAutoServer),
		dht.Validator(validator),
		dht.ProtocolPrefix(protocolPrefix),
	)
	if err != nil {
		return util.LoggingErrorMsg(err, "failed to instantiate dht service")
	}
	if err = d.Bootstrap(ctx); err != nil {
		return util.LoggingErrorMsg(err, "failed to bootstrap dht service")
	}

	s.host = routedhost.Wrap(s.host, d)
	s.dht = d
	return nil
}

func (s *Service) bootstrapPeers(ctx context.Context) error {
	// connect to bootstrap bootstrapPeers
	logrus.Info("connecting to bootstrap peers")
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

	// discover and connect to peers periodically
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				pctx, cancel := context.WithTimeout(ctx, peerDiscoveryPeriod)
				s.discover(pctx)
				cancel()
				time.Sleep(peerDiscoveryPeriod)
			}
		}
	}()
	return nil
}

func (s *Service) discover(ctx context.Context) {
	logrus.Info("finding peers")
	peerChan, err := s.discovery.FindPeers(ctx, s.cfg.DHTConfig.Namespace, discovery.Limit(peerLimit))
	if err != nil {
		logrus.WithError(err).Error("failed to find peers")
		return
	}
	for p := range peerChan {
		p := p
		select {
		case <-ctx.Done():
			return
		default:
			if p.ID == s.host.ID() {
				continue
			}
			if err = s.host.Connect(ctx, p); err != nil {
				logrus.WithError(err).Errorf("failed to connect to peer %s", p.ID)
			} else {
				logrus.Infof("connected to peer %s", p.ID)
			}
			s.host.ConnManager().Protect(p.ID, "discoveredPeer")
		}
	}
}

func (s *Service) setupLocalDiscovery(ctx context.Context) error {
	ldn := new(localDiscoveryNotifee)
	ldn.PeerChan = make(chan peer.AddrInfo)
	svc := mdns.NewMdnsService(s.host, s.cfg.DHTConfig.Namespace, ldn)
	if err := svc.Start(); err != nil {
		return util.LoggingErrorMsg(err, "failed to start mdns service")
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
					logrus.Infof("connected to local peer %s", pi.ID)
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
