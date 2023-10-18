package dht

import (
	"context"
	"testing"
	"time"

	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/torrent/bencode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/internal/util"
)

func TestGetPutDHT(t *testing.T) {
	d, err := NewDHT(config.GetDefaultBootstrapPeers())
	require.NoError(t, err)

	pubKey, privKey, err := util.GenerateKeypair()
	require.NoError(t, err)

	put := &bep44.Put{
		V:   []byte("hello dht"),
		K:   (*[32]byte)(pubKey),
		Seq: time.Now().UnixMilli() / 1000,
	}
	put.Sign(privKey)

	id, err := d.Put(context.Background(), *put)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	got, err := d.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	var payload string
	err = bencode.Unmarshal(got.V, &payload)
	require.NoError(t, err)

	assert.Equal(t, string(put.V.([]byte)), payload)
}
