package dht

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"time"

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

// Get returns the value for the given key from the DHT.
// The key is a z32-encoded string, such as "yj47pezutnpw9pyudeeai8cx8z8d6wg35genrkoqf9k3rmfzy58o".
func (d *DHT) Get(ctx context.Context, key string) (string, error) {
	z32Decoded, err := util.Z32Decode(key)
	if err != nil {
		logrus.WithError(err).Error("failed to decode key")
		return "", err
	}
	res, t, err := getput.Get(ctx, infohash.HashBytes(z32Decoded), d.Server, nil, nil)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get key<%s> from dht; tried %d nodes, got %d responses", key, t.NumAddrsTried, t.NumResponses)
		return "", err
	}
	var payload string
	if err = bencode.Unmarshal(res.V, &payload); err != nil {
		logrus.WithError(err).Error("failed to unmarshal payload value")
		return "", err
	}
	decoded, err := util.Decode([]byte(payload))
	if err != nil {
		logrus.WithError(err).Error("failed to decode value from dht")
		return "", err
	}
	return string(decoded), nil
}

// Put puts the given value into the DHT. It's recommended to use CreatePutRequest to create the request.
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

// CreatePutRequest creates a put request for the given records. Requires a public/private keypair and the records to put.
// The records are expected to be a slice of slices of values, such as:
//
//	[][]any{
//		{"foo", "bar"},
//	}
func CreatePutRequest(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, records [][]any) (*bep44.Put, error) {
	recordsBytes, err := json.Marshal(records)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal records")
		return nil, err
	}
	encodedV, err := util.Encode(recordsBytes)
	if err != nil {
		logrus.WithError(err).Error("failed to encode records")
		return nil, err
	}
	put := &bep44.Put{
		V:   encodedV,
		K:   (*[32]byte)(publicKey),
		Seq: time.Now().UnixMilli() / 1000,
	}
	put.Sign(privateKey)
	return put, nil
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
