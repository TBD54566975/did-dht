package did

import (
	"crypto/ed25519"
	"fmt"
	"strings"
	"testing"

	"github.com/TBD54566975/ssi-sdk/cryptosuite"
	"github.com/goccy/go-json"

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
						Type:         cryptosuite.JSONWebKeyType,
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

		jsonDoc, err := json.Marshal(doc)
		require.NoError(t, err)

		jsonDecodedDoc, err := json.Marshal(decodedDoc)
		require.NoError(t, err)

		assert.JSONEq(t, string(jsonDoc), string(jsonDecodedDoc))
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
						Type:         cryptosuite.JSONWebKeyType,
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
					Sig:             []string{"1", "2"},
					Enc:             "3",
				},
				{
					ID:              "hub",
					Type:            "MessagingService",
					ServiceEndpoint: []string{"https://example.com/hub/", "https://example.com/hub2/"},
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

		decodedJSON, err := json.Marshal(decodedDoc)
		require.NoError(t, err)

		docJSON, err := json.Marshal(doc)
		require.NoError(t, err)

		assert.JSONEq(t, string(docJSON), string(decodedJSON))
	})
}

func TestVectors(t *testing.T) {
	type testVectorDNSRecord struct {
		Name       string `json:"name"`
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

		docJSON, err := json.Marshal(doc)
		require.NoError(t, err)

		expectedDIDDocJSON, err := json.Marshal(expectedDIDDocument)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedDIDDocJSON), string(docJSON))

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		var expectedDNSRecords []testVectorDNSRecord
		retrieveTestVectorAs(t, vector1DNSRecords, &expectedDNSRecords)

		// Initialize a map to track matched records
		matchedRecords := make(map[int]bool)
		for i := range expectedDNSRecords {
			matchedRecords[i] = false // Initialize all expected records as unmatched
		}

		for _, record := range packet.Answer {
			for i, expectedRecord := range expectedDNSRecords {
				if record.Header().Name == expectedRecord.Name {
					s := record.String()
					if strings.Contains(s, expectedRecord.RecordType) &&
						strings.Contains(s, expectedRecord.TTL) &&
						strings.Contains(s, expectedRecord.Record) {
						matchedRecords[i] = true // Mark as matched
						break
					}
				}
			}
		}

		// Check if all expected records have been matched
		for i, matched := range matchedRecords {
			require.True(t, matched, fmt.Sprintf("Expected DNS record %d: %+v not matched", i, expectedDNSRecords[i]))
		}

		// Make sure going back to DID Document is consistent
		decodedDoc, types, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, decodedDoc)
		require.Empty(t, types)

		decodedDocJSON, err := json.Marshal(decodedDoc)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedDIDDocJSON), string(decodedDocJSON))
	})

	t.Run("test vector 2", func(t *testing.T) {
		var pubKeyJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector1PublicKeyJWK1, &pubKeyJWK)

		pubKey, err := pubKeyJWK.ToPublicKey()
		require.NoError(t, err)

		var secpJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector2PublicKeyJWK2, &secpJWK)

		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			AuthoritativeGateways: []string{
				"gateway1.example-did-dht-gateway.com.",
				"gateway2.example-did-dht-gateway.com.",
			},
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh", "did:example:ijkl"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		var expectedDIDDocument did.Document
		retrieveTestVectorAs(t, vector2DIDDocument, &expectedDIDDocument)

		docJSON, err := json.Marshal(doc)
		require.NoError(t, err)

		expectedDIDDocJSON, err := json.Marshal(expectedDIDDocument)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedDIDDocJSON), string(docJSON))

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, []TypeIndex{1, 2, 3})
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		var expectedDNSRecords []testVectorDNSRecord
		retrieveTestVectorAs(t, vector2DNSRecords, &expectedDNSRecords)

		// Initialize a map to track matched records
		matchedRecords := make(map[int]bool)
		for i := range expectedDNSRecords {
			matchedRecords[i] = false // Initialize all expected records as unmatched
		}

		for _, record := range packet.Answer {
			for i, expectedRecord := range expectedDNSRecords {
				if record.Header().Name == expectedRecord.Name {
					s := record.String()
					if strings.Contains(s, expectedRecord.RecordType) &&
						strings.Contains(s, expectedRecord.TTL) &&
						strings.Contains(s, expectedRecord.Record) {
						matchedRecords[i] = true // Mark as matched
						break
					}
				}
			}
		}

		// Check if all expected records have been matched
		for i, matched := range matchedRecords {
			require.True(t, matched, fmt.Sprintf("Expected DNS record %d: %+v not matched", i, expectedDNSRecords[i]))
		}

		// Make sure going back to DID Document is consistent
		decodedDoc, types, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, decodedDoc)
		require.NotEmpty(t, types)
		require.Equal(t, types, []TypeIndex{1, 2, 3})

		decodedDocJSON, err := json.Marshal(decodedDoc)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedDIDDocJSON), string(decodedDocJSON))
	})

	t.Run("test vector 3", func(t *testing.T) {
		var pubKeyJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector3PublicKeyJWK1, &pubKeyJWK)

		pubKey, err := pubKeyJWK.ToPublicKey()
		require.NoError(t, err)

		var x25519JWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector3PublicKeyJWK2, &x25519JWK)

		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           x25519JWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &x25519JWK,
					},
					Purposes: []did.PublicKeyPurpose{did.KeyAgreement},
				},
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		var expectedDIDDocument did.Document
		retrieveTestVectorAs(t, vector3DIDDocument, &expectedDIDDocument)

		docJSON, err := json.Marshal(doc)
		require.NoError(t, err)

		expectedDIDDocJSON, err := json.Marshal(expectedDIDDocument)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedDIDDocJSON), string(docJSON))

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		var expectedDNSRecords []testVectorDNSRecord
		retrieveTestVectorAs(t, vector3DNSRecords, &expectedDNSRecords)

		// Initialize a map to track matched records
		matchedRecords := make(map[int]bool)
		for i := range expectedDNSRecords {
			matchedRecords[i] = false // Initialize all expected records as unmatched
		}

		for _, record := range packet.Answer {
			for i, expectedRecord := range expectedDNSRecords {
				if record.Header().Name == expectedRecord.Name {
					s := record.String()
					if strings.Contains(s, expectedRecord.RecordType) &&
						strings.Contains(s, expectedRecord.TTL) &&
						strings.Contains(s, expectedRecord.Record) {
						matchedRecords[i] = true // Mark as matched
						break
					}
				}
			}
		}

		// Check if all expected records have been matched
		for i, matched := range matchedRecords {
			require.True(t, matched, fmt.Sprintf("Expected DNS record %d: %+v not matched", i, expectedDNSRecords[i]))
		}

		// Make sure going back to DID Document is consistent
		decodedDoc, types, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, decodedDoc)
		require.Empty(t, types)

		decodedDocJSON, err := json.Marshal(decodedDoc)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedDIDDocJSON), string(decodedDocJSON))
	})
}

func TestMisc(t *testing.T) {
	t.Run("DHT.Method()", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		didID := DHT(doc.ID)
		require.Equal(t, didID.Method(), did.Method("dht"))
	})

	var pubKeyJWK jwx.PublicKeyJWK
	retrieveTestVectorAs(t, vector1PublicKeyJWK1, &pubKeyJWK)

	pubKey, err := pubKeyJWK.ToPublicKey()
	require.NoError(t, err)

	var secpJWK jwx.PublicKeyJWK
	retrieveTestVectorAs(t, vector2PublicKeyJWK2, &secpJWK)

	t.Run("single aka field", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, doc)
	})

	t.Run("verification method without ID", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, doc)
	})

	t.Run("multiple controllers", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd", "did:example:ijkl"},
			AlsoKnownAs: []string{"did:example:efgh"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, doc)
	})

	t.Run("verification method ID # prefix", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           "#key-1",
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, doc)
	})

	t.Run("verification method purposes", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.Authentication, did.KeyAgreement, did.CapabilityDelegation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, doc)
	})
}

func TestCreationFailures(t *testing.T) {
	var pubKeyJWK jwx.PublicKeyJWK
	retrieveTestVectorAs(t, vector1PublicKeyJWK1, &pubKeyJWK)

	pubKey, err := pubKeyJWK.ToPublicKey()
	require.NoError(t, err)

	var secpJWK jwx.PublicKeyJWK
	retrieveTestVectorAs(t, vector2PublicKeyJWK2, &secpJWK)

	t.Run("verification method id 0", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh", "did:example:ijkl"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           "#0",
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.EqualError(t, err, "verification method id 0 is reserved for the identity key")
		assert.Nil(t, doc)
	})
	t.Run("duplicate verification method id", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh", "did:example:ijkl"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.EqualError(t, err, fmt.Sprintf("verification method id %s is not unique", secpJWK.KID))
		assert.Nil(t, doc)
	})

	t.Run("unsupported verification method", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh", "did:example:ijkl"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         "fake",
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.EqualError(t, err, "verification method type fake is not supported")
		assert.Nil(t, doc)
	})

	t.Run("nil verification public key", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh", "did:example:ijkl"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: nil,
					},
					Purposes: []did.PublicKeyPurpose{did.AssertionMethod, did.CapabilityInvocation},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.EqualError(t, err, "verification method public key jwk is required")
		assert.Nil(t, doc)
	})

	t.Run("unknown key purpose", func(t *testing.T) {
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{
			Controller:  []string{"did:example:abcd"},
			AlsoKnownAs: []string{"did:example:efgh"},
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
						ID:           secpJWK.KID,
						Type:         cryptosuite.JSONWebKeyType,
						PublicKeyJWK: &secpJWK,
					},
					Purposes: []did.PublicKeyPurpose{"fake purpose"},
				},
			},
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestService",
					ServiceEndpoint: []string{"https://test-service.com/1", "https://test-service.com/2"},
				},
			},
		})
		assert.Error(t, err)
		assert.Nil(t, doc)
	})

}
