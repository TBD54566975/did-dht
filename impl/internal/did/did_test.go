package did

import (
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/did/ion"
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
		pubKey, _, err := crypto.GenerateSECP256k1Key()
		require.NoError(t, err)
		pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK("key1", pubKey)
		require.NoError(t, err)

		opts := CreateDIDDHTOpts{
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           "did:dht:123456789abcdefghi#key1",
						Type:         "JsonWebKey2020",
						Controller:   "did:dht:123456789abcdefghi",
						PublicKeyJWK: pubKeyJWK,
					},
					Purposes: []ion.PublicKeyPurpose{ion.AssertionMethod, ion.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "did:dht:123456789abcdefghi#vcs",
					Type:            "VerifiableCredentialService",
					ServiceEndpoint: "https://example.com/vc/",
				},
				{
					ID:              "did:dht:123456789abcdefghi#hub",
					Type:            "MessagingService",
					ServiceEndpoint: "https://example.com/hub/",
				},
			},
		}

		privKey, doc, err := GenerateDIDDHT(opts)
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		assert.NotEmpty(t, doc.ID)
		assert.Empty(t, doc.Context)
		assert.NotEmpty(t, doc.VerificationMethod)
		assert.NotEmpty(t, doc.Authentication)
		assert.NotEmpty(t, doc.AssertionMethod)
		assert.Empty(t, doc.KeyAgreement)
		assert.NotEmpty(t, doc.CapabilityDelegation)
		assert.NotEmpty(t, doc.CapabilityInvocation)
		assert.NotEmpty(t, doc.Services)

		assert.Len(t, doc.VerificationMethod, 2)
		assert.Len(t, doc.Authentication, 1)
		assert.Len(t, doc.AssertionMethod, 2)
		assert.Len(t, doc.KeyAgreement, 0)
		assert.Len(t, doc.CapabilityDelegation, 1)
		assert.Len(t, doc.CapabilityInvocation, 2)
		assert.Len(t, doc.Services, 2)

		assert.NotEmpty(t, doc.VerificationMethod[0].ID)
		assert.EqualValues(t, doc.ID+"#0", doc.VerificationMethod[0].ID)
		assert.NotEmpty(t, doc.VerificationMethod[0].Controller)
		assert.Equal(t, doc.ID, doc.VerificationMethod[0].Controller)
		assert.NotEmpty(t, doc.VerificationMethod[0].Type)
		assert.NotEmpty(t, doc.VerificationMethod[0].PublicKeyJWK)

		assert.Equal(t, doc.Services[0].ID, "did:dht:123456789abcdefghi#vcs")
		assert.Equal(t, doc.Services[0].Type, "VerifiableCredentialService")
		assert.Equal(t, doc.Services[0].ServiceEndpoint, "https://example.com/vc/")

		assert.Equal(t, doc.Services[1].ID, "did:dht:123456789abcdefghi#hub")
		assert.Equal(t, doc.Services[1].Type, "MessagingService")
		assert.Equal(t, doc.Services[1].ServiceEndpoint, "https://example.com/hub/")
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

	t.Run("doc with multiple keys and services - test to dns packet round trip", func(t *testing.T) {
		pubKey, _, err := crypto.GenerateSECP256k1Key()
		require.NoError(t, err)
		pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK("key1", pubKey)
		require.NoError(t, err)

		opts := CreateDIDDHTOpts{
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           "key1",
						Type:         "JsonWebKey2020",
						Controller:   "did:dht:123456789abcdefghi",
						PublicKeyJWK: pubKeyJWK,
					},
					Purposes: []ion.PublicKeyPurpose{ion.AssertionMethod, ion.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "vcs",
					Type:            "VerifiableCredentialService",
					ServiceEndpoint: "https://example.com/vc/",
				},
				{
					ID:              "hub",
					Type:            "MessagingService",
					ServiceEndpoint: "https://example.com/hub/",
				},
			},
		}
		privKey, doc, err := GenerateDIDDHT(opts)
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
