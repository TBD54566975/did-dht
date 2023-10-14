package did

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateDIDDHT(t *testing.T) {
	t.Run("test generate did:dht with no options", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)
	})
}

func TestToDNSPacket(t *testing.T) {
	t.Run("test to dns packet", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		did := DHT(doc.ID)

		packet, err := did.ToDNSPacket(*doc)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		println(packet.String())
	})
}
