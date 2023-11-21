package did

import (
	"crypto/ed25519"
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/did"
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
		pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK(nil, pubKey)
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
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
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

		assert.Equal(t, doc.Services[0].ID, doc.ID+"#vcs")
		assert.Equal(t, doc.Services[0].Type, "VerifiableCredentialService")
		assert.Equal(t, doc.Services[0].ServiceEndpoint, "https://example.com/vc/")

		assert.Equal(t, doc.Services[1].ID, doc.ID+"#hub")
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

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		decodedDoc, types, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, decodedDoc)
		require.Empty(t, types)

		assert.EqualValues(t, *doc, *decodedDoc)
	})

	t.Run("doc with types - test to dns packet round trip", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, []TypeIndex{1, 2, 3})
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		decodedDoc, types, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, decodedDoc)
		require.NotEmpty(t, types)
		require.Equal(t, types, []TypeIndex{1, 2, 3})

		assert.EqualValues(t, *doc, *decodedDoc)
	})

	t.Run("doc with multiple keys and services - test to dns packet round trip", func(t *testing.T) {
		pubKey, _, err := crypto.GenerateSECP256k1Key()
		require.NoError(t, err)
		pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK(nil, pubKey)
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
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
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

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		decodedDoc, types, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, decodedDoc)
		require.Empty(t, types)

		assert.EqualValues(t, *doc, *decodedDoc)
	})
}

func TestVectors(t *testing.T) {

	type testVectorDNSRecord struct {
		RecordType string `json:"type"`
		TTL        string `json:"ttl"`
		Record     string `json:"rdata"`
	}

	t.Run("test vector 1", func(t *testing.T) {
		var pubKeyJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector1PublicKeyJWK1, &pubKeyJWK)

		pubKey, err := pubKeyJWK.ToPublicKey()
		require.NoError(t, err)

		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		var expectedDIDDocument did.Document
		retrieveTestVectorAs(t, vector1DIDDocument, &expectedDIDDocument)
		assert.EqualValues(t, expectedDIDDocument, *doc)

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		var expectedDNSRecords map[string]testVectorDNSRecord
		retrieveTestVectorAs(t, vector1DNSRecords, &expectedDNSRecords)

		for _, record := range packet.Answer {
			expectedRecord, ok := expectedDNSRecords[record.Header().Name]
			require.True(t, ok)

			s := record.String()
			assert.Contains(t, s, expectedRecord.RecordType)
			assert.Contains(t, s, expectedRecord.TTL)
			assert.Contains(t, s, expectedRecord.Record)
		}
	})

	t.Run("test vector 2", func(t *testing.T) {
		var pubKeyJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector1PublicKeyJWK1, &pubKeyJWK)

		pubKey, err := pubKeyJWK.ToPublicKey()
		require.NoError(t, err)

		var secpJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector2PublicKeyJWK2, &secpJWK)

		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         "JsonWebKey2020",
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: "https://test-service.com",
				},
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		var expectedDIDDocument did.Document
		retrieveTestVectorAs(t, vector2DIDDocument, &expectedDIDDocument)
		assert.EqualValues(t, expectedDIDDocument, *doc)

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, []TypeIndex{1, 2, 3})
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		var expectedDNSRecords map[string]testVectorDNSRecord
		retrieveTestVectorAs(t, vector2DNSRecords, &expectedDNSRecords)

		println(packet.String())
		for _, record := range packet.Answer {
			expectedRecord, ok := expectedDNSRecords[record.Header().Name]
			require.True(t, ok)

			s := record.String()
			assert.Contains(t, s, expectedRecord.RecordType)
			assert.Contains(t, s, expectedRecord.TTL)
			assert.Contains(t, s, expectedRecord.Record)
		}
	})
}
