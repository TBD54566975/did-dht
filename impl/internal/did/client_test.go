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

	// wait for the record to be published
	time.Sleep(10 * time.Second)

	gotDID, _, err := client.GetDIDDocument(doc.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, doc, gotDID)
}
