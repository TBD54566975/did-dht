package did

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

func TestClient(t *testing.T) {
	client, err := NewGatewayClient("https://diddht.tbddev.org")
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

	since := time.Since(start)
	t.Logf("time to put and get: %s", since)
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
}
