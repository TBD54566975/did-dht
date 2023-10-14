package dht

import (
	"context"
	"crypto/ed25519"

	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/dht/v2/exts/getput"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/types/infohash"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/internal/util"
)

type DHT struct {
	*dht.Server
}

func NewDHT() (*DHT, error) {
	c := dht.NewDefaultServerConfig()
	c.StartingNodes = func() ([]dht.Addr, error) { return dht.ResolveHostPorts(getDefaultBootstrapPeers()) }
	s, err := dht.NewServer(c)
	if err != nil {
		logrus.WithError(err).Error("failed to create dht server")
		return nil, err
	}
	return &DHT{Server: s}, nil
}

// Get returns the BEP-44 value for the given key from the DHT.
// The key is a z32-encoded string, such as "yj47pezutnpw9pyudeeai8cx8z8d6wg35genrkoqf9k3rmfzy58o".
func (d *DHT) Get(ctx context.Context, key string) ([]byte, error) {
	z32Decoded, err := util.Z32Decode(key)
	if err != nil {
		logrus.WithError(err).Error("failed to decode key")
		return nil, err
	}
	res, t, err := getput.Get(ctx, infohash.HashBytes(z32Decoded), d.Server, nil, nil)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get key<%s> from dht; tried %d nodes, got %d responses", key, t.NumAddrsTried, t.NumResponses)
		return nil, err
	}
	var payload string
	if err = bencode.Unmarshal(res.V, &payload); err != nil {
		logrus.WithError(err).Error("failed to unmarshal payload value")
		return nil, err
	}
	return []byte(payload), nil
}

// Put puts the given BEP-44 value into the DHT and returns its z32-encoded key.
func (d *DHT) Put(ctx context.Context, key ed25519.PublicKey, request bep44.Put) (string, error) {
	t, err := getput.Put(ctx, request.Target(), d.Server, nil, func(int64) bep44.Put {
		return request
	})
	if err != nil {
		logrus.WithError(err).Errorf("failed to put key into dht, tried %d nodes, got %d responses", t.NumAddrsTried, t.NumResponses)
		return "", err
	}
	return util.Z32Encode(key), nil
}

func getDefaultBootstrapPeers() []string {
	return []string{
		"router.magnets.im:6881",
		"router.bittorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"router.utorrent.com:6881",
		"router.nuh.dev:6881",
	}
}
