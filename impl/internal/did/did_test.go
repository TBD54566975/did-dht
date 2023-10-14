package did

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDIDDHT(t *testing.T) {
	t.Run("test generate did:dht with no options", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		assert.NotEmpty(t, doc.ID)
		assert.Empty(t, doc.Context)
		assert.NotEmpty(t, doc.VerificationMethod)
		assert.NotEmpty(t, doc.Authentication)
		assert.NotEmpty(t, doc.AssertionMethod)
		assert.NotEmpty(t, doc.CapabilityDelegation)
		assert.NotEmpty(t, doc.CapabilityInvocation)
		assert.Empty(t, doc.Services)

		assert.Len(t, doc.VerificationMethod, 1)
		assert.Len(t, doc.Authentication, 1)
		assert.Len(t, doc.AssertionMethod, 1)
		assert.Len(t, doc.CapabilityDelegation, 1)
		assert.Len(t, doc.CapabilityInvocation, 1)

		assert.NotEmpty(t, doc.VerificationMethod[0].ID)
		assert.EqualValues(t, doc.ID+"#0", doc.VerificationMethod[0].ID)
		assert.NotEmpty(t, doc.VerificationMethod[0].Controller)
		assert.Equal(t, doc.ID, doc.VerificationMethod[0].Controller)
		assert.NotEmpty(t, doc.VerificationMethod[0].Type)
		assert.NotEmpty(t, doc.VerificationMethod[0].PublicKeyJWK)
	})

	t.Run("test generate did:dht with opts", func(t *testing.T) {

	})
}

func TestToDNSPacket(t *testing.T) {
	t.Run("simple doc - test to dns packet round trip", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		did := DHT(doc.ID)

		packet, err := did.ToDNSPacket(*doc)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		decodedDoc, err := did.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, decodedDoc)

		assert.EqualValues(t, *doc, *decodedDoc)
	})
}
