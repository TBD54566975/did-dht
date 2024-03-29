package did

import (
	"testing"
	"time"

	"github.com/anacrolix/dht/v2/bep44"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

func TestClient(t *testing.T) {
	// client, err := NewGatewayClient("https://diddht.tbddev.org")
	// client, err := NewGatewayClient("https://diddht-staging.tbd.engineering")
	client, err := NewGatewayClient("http://0.0.0.0:8305")

	require.NoError(t, err)
	require.NotNil(t, client)

	start := time.Now()

	sk, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)

	packet, err := DHT(doc.ID).ToDNSPacket(*doc, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, packet)

	bep44Put, err := dht.CreatePKARRPublishRequest(sk, *packet)
	assert.NoError(t, err)
	assert.NotEmpty(t, bep44Put)

	err = client.PutDocument(doc.ID, *bep44Put)
	assert.NoError(t, err)

	gotDID, _, err := client.GetDIDDocument(doc.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, doc, gotDID)

	println(doc.ID)
	since := time.Since(start)
	t.Logf("time to put and get: %s", since)
}

func TestGet(t *testing.T) {
	// client, err := NewGatewayClient("https://diddht.tbddev.org")
	// client, err := NewGatewayClient("https://diddht-staging.tbd.engineering")
	client, err := NewGatewayClient("http://0.0.0.0:8305")

	require.NoError(t, err)
	require.NotNil(t, client)
	// 3h1ukmqyrfbzhr94ykum1j33gdoiibugabcbm8jfcyy1ft77efsy <-- in staging
	// aptterwukoxrqgx75jwei67hw53c8uxdukgx4yhta8m4gedbseco <-- in dev
	did, _, err := client.GetDIDDocument("did:dht:d4j51t8hm6by36pscqun97f874br81fgyz1cwdbzqfyo38bq7o7o")
	assert.NoError(t, err)
	assert.NotNil(t, did)

}

func TestClientInvalidGateway(t *testing.T) {
	g, err := NewGatewayClient("\n")
	assert.Error(t, err)
	assert.Nil(t, g)
}

func TestInvalidDIDDocument(t *testing.T) {
	client, err := NewGatewayClient("https://diddht.tbddev.test")
	require.NoError(t, err)
	require.NotNil(t, client)

	did, ty, err := client.GetDIDDocument("this is not a valid did")
	assert.Error(t, err)
	assert.Nil(t, ty)
	assert.Nil(t, did)

	did, ty, err = client.GetDIDDocument("did:dht:example")
	assert.EqualError(t, err, "invalid did")
	assert.Nil(t, ty)
	assert.Nil(t, did)

	did, ty, err = client.GetDIDDocument("did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y")
	assert.Error(t, err) // this should error because the gateway URL is invalid
	assert.Nil(t, ty)
	assert.Nil(t, did)

	client, err = NewGatewayClient("https://tbd.website")
	require.NoError(t, err)
	require.NotNil(t, client)

	did, ty, err = client.GetDIDDocument("did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y")
	assert.Error(t, err) // this should error because the gateway URL will return a non-200
	assert.Nil(t, ty)
	assert.Nil(t, did)

	err = client.PutDocument("did:dht:example", bep44.Put{})
	assert.Error(t, err)

	err = client.PutDocument("did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y", bep44.Put{
		K: &[32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		V: []byte{0, 0, 0},
	})
	assert.Error(t, err)
}
