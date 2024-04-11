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
	client, err := NewGatewayClient("https://diddht.tbddev.org")

	require.NoError(t, err)
	require.NotNil(t, client)

	start := time.Now()

	sk, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)

	packet, err := DHT(doc.ID).ToDNSPacket(*doc, nil, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, packet)

	bep44Put, err := dht.CreatePkarrPublishRequest(sk, *packet)
	assert.NoError(t, err)
	assert.NotEmpty(t, bep44Put)

	err = client.PutDocument(doc.ID, *bep44Put)
	assert.NoError(t, err)

	gotDID, _, _, err := client.GetDIDDocument(doc.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, doc, gotDID)

	since := time.Since(start)
	t.Logf("time to put and get: %s", since)
}

func TestGet(t *testing.T) {
	client, err := NewGatewayClient("http://localhost:8305")

	require.NoError(t, err)
	require.NotNil(t, client)

	doc, _, _, err := client.GetDIDDocument("did:dht:uqaj3fcr9db6jg6o9pjs53iuftyj45r46aubogfaceqjbo6pp9sy")
	require.NoError(t, err)
	require.NotNil(t, doc)
}

func TestClientInvalidGateway(t *testing.T) {
	g, err := NewGatewayClient("\n")
	assert.Error(t, err)
	assert.Nil(t, g)
}

func TestInvalidDIDDocument(t *testing.T) {
	client, err := NewGatewayClient("https://diddht.tbddev.test")
	require.NoError(t, err)
	require.NotEmpty(t, client)

	did, types, gateways, err := client.GetDIDDocument("this is not a valid did")
	assert.Error(t, err)
	assert.Empty(t, did)
	assert.Empty(t, types)
	assert.Empty(t, gateways)

	did, types, gateways, err = client.GetDIDDocument("did:dht:example")
	assert.EqualError(t, err, "invalid did")
	assert.Empty(t, did)
	assert.Empty(t, types)
	assert.Empty(t, gateways)

	did, types, gateways, err = client.GetDIDDocument("did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y")
	assert.Error(t, err) // this should error because the gateway URL is invalid
	assert.Empty(t, did)
	assert.Empty(t, types)
	assert.Empty(t, gateways)

	client, err = NewGatewayClient("https://tbd.website")
	require.NoError(t, err)
	require.NotEmpty(t, client)

	did, types, gateways, err = client.GetDIDDocument("did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y")
	assert.Error(t, err) // this should error because the gateway URL will return a non-200
	assert.Empty(t, did)
	assert.Empty(t, types)
	assert.Empty(t, gateways)

	err = client.PutDocument("did:dht:example", bep44.Put{})
	assert.Error(t, err)

	err = client.PutDocument("did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y", bep44.Put{
		K: &[32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		V: []byte{0, 0, 0},
	})
	assert.Error(t, err)
}
