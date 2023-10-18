package dht

import (
	"context"

	errutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/dht/v2/exts/getput"
	"github.com/anacrolix/torrent/types/infohash"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/internal/util"
)

type DHT struct {
	*dht.Server
}

// NewDHT returns a new instance of DHT with the given bootstrap peers.
func NewDHT(bootstrapPeers []string) (*DHT, error) {
	c := dht.NewDefaultServerConfig()
	c.StartingNodes = func() ([]dht.Addr, error) { return dht.ResolveHostPorts(bootstrapPeers) }
	s, err := dht.NewServer(c)
	if err != nil {
		logrus.WithError(err).Error("failed to create dht server")
		return nil, err
	}
	return &DHT{Server: s}, nil
}

// Put puts the given BEP-44 value into the DHT and returns its z32-encoded key.
func (d *DHT) Put(ctx context.Context, request bep44.Put) (string, error) {
	t, err := getput.Put(ctx, request.Target(), d.Server, nil, func(int64) bep44.Put {
		return request
	})
	if err != nil {
		return "", errutil.LoggingNewErrorf("failed to put key into dht, tried %d nodes, got %d responses", t.NumAddrsTried, t.NumResponses)
	}
	return util.Z32Encode(request.K[:]), nil
}

// Get returns the BEP-44 result for the given key from the DHT.
// The key is a z32-encoded string, such as "yj47pezutnpw9pyudeeai8cx8z8d6wg35genrkoqf9k3rmfzy58o".
func (d *DHT) Get(ctx context.Context, key string) (*getput.GetResult, error) {
	z32Decoded, err := util.Z32Decode(key)
	if err != nil {
		logrus.WithError(err).Error("failed to decode key")
		return nil, err
	}
	res, t, err := getput.Get(ctx, infohash.HashBytes(z32Decoded), d.Server, nil, nil)
	if err != nil {
		return nil, errutil.LoggingNewErrorf("failed to get key<%s> from dht; tried %d nodes, got %d responses", key, t.NumAddrsTried, t.NumResponses)
	}
	return &res, nil
}
