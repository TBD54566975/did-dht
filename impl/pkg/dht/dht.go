package dht

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	errutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent/types/infohash"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	dhtint "github.com/TBD54566975/did-dht/internal/dht"
	"github.com/TBD54566975/did-dht/internal/util"
	"github.com/TBD54566975/did-dht/pkg/telemetry"
)

// DHT is a wrapper around anacrolix/dht that implements the BEP-44 DHT protocol.
type DHT struct {
	*dht.Server
}

// NewDHT returns a new instance of DHT with the given bootstrap peers.
func NewDHT(bootstrapPeers []string) (*DHT, error) {
	logrus.WithField("bootstrap_peers", len(bootstrapPeers)).Info("initializing DHT")

	c := dht.NewDefaultServerConfig()
	c.Exp = time.Hour * 24
	c.NoSecurity = false
	conn, err := net.ListenPacket("udp", "0.0.0.0:6881")
	if err != nil {
		return nil, errutil.LoggingErrorMsg(err, "failed to listen on udp port 6881")
	}
	c.Conn = conn
	c.Logger = log.NewLogger().WithFilterLevel(log.Debug)
	c.Logger.SetHandlers(logrusHandler{})
	c.StartingNodes = func() ([]dht.Addr, error) { return dht.ResolveHostPorts(bootstrapPeers) }
	// set up rate limiter - 100 requests per second, 500 requests burst
	c.SendLimiter = rate.NewLimiter(100, 500)
	s, err := dht.NewServer(c)
	if err != nil {
		return nil, errutil.LoggingErrorMsg(err, "failed to create dht server")
	}
	if tried, err := s.Bootstrap(); err != nil {
		return nil, errutil.LoggingErrorMsg(err, "error bootstrapping")
	} else {
		logrus.WithField("bootstrap_peers", tried.NumResponses).Info("bootstrapped DHT successfully")
	}
	return &DHT{Server: s}, nil
}

// NewTestDHT returns a new instance of DHT that does not make external connections
func NewTestDHT(t *testing.T, bootstrapPeers ...dht.Addr) *DHT {
	c := dht.NewDefaultServerConfig()
	c.WaitToReply = true

	conn, err := net.ListenPacket("udp", "localhost:0")
	require.NoError(t, err)
	c.Conn = conn

	if len(bootstrapPeers) == 0 {
		bootstrapPeers = []dht.Addr{dht.NewAddr(c.Conn.LocalAddr())}
	}
	c.StartingNodes = func() ([]dht.Addr, error) { return bootstrapPeers, nil }

	s, err := dht.NewServer(c)
	require.NoError(t, err)
	require.NotNil(t, s)

	if _, err = s.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap: %v", err)
	}

	return &DHT{Server: s}
}

// Put puts the given BEP-44 value into the DHT and returns its z32-encoded key.
func (d *DHT) Put(ctx context.Context, request bep44.Put) (string, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "DHT.Put")
	defer span.End()

	// Check if there are any nodes in the DHT
	if len(d.Server.Nodes()) == 0 {
		logrus.WithContext(ctx).Warn("no nodes available in the DHT for publishing")
	}

	key := util.Z32Encode(request.K[:])
	t, err := dhtint.Put(ctx, request.Target(), d.Server, nil, func(int64) bep44.Put {
		return request
	})
	if err != nil {
		if t == nil {
			return "", fmt.Errorf("failed to put key[%s] into dht: %v", key, err)
		}
		return "", fmt.Errorf("failed to put key[%s] into dht, tried %d nodes, got %d responses", key, t.NumAddrsTried, t.NumResponses)
	} else {
		logrus.WithContext(ctx).WithField("key", key).Debug("successfully put key into dht")
	}
	return util.Z32Encode(request.K[:]), nil
}

// GetFull returns the full BEP-44 result for the given key from the DHT, using our modified
// implementation of getput.Get. It should ONLY be used when it's needed to get the signature
// data for a record.
func (d *DHT) GetFull(ctx context.Context, key string) (*dhtint.FullGetResult, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "DHT.GetFull")
	defer span.End()

	z32Decoded, err := util.Z32Decode(key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode key [%s]", key)
	}
	res, t, err := dhtint.Get(ctx, infohash.HashBytes(z32Decoded), d.Server, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get key[%s] from dht; tried %d nodes, got %d responses", key, t.NumAddrsTried, t.NumResponses)
	}
	return &res, nil
}
