package did

import (
	"crypto/ed25519"
	"fmt"
	"strings"
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/cryptosuite"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVectors from the spec https://did-dht.com/#test-vectors
func TestVectors(t *testing.T) {
	type testVectorDNSRecord struct {
		Name       string   `json:"name"`
		RecordType string   `json:"type"`
		TTL        string   `json:"ttl"`
		Record     []string `json:"rdata"`
	}

	// https://did-dht.com/#vector-1
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
		packet, err := didID.ToDNSPacket(*doc, nil, nil, nil)
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
						strings.Contains(s, strings.Join(expectedRecord.Record, "")) {
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
		didDHTDoc, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.Empty(t, didDHTDoc.Types)
		require.Empty(t, didDHTDoc.Gateways)

		decodedDocJSON, err := json.Marshal(didDHTDoc.Doc)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedDIDDocJSON), string(decodedDocJSON))
	})

	// https://did-dht.com/#vector-2
	t.Run("test vector 2", func(t *testing.T) {
		var pubKeyJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector1PublicKeyJWK1, &pubKeyJWK)

		pubKey, err := pubKeyJWK.ToPublicKey()
		require.NoError(t, err)

		var secpJWK jwx.PublicKeyJWK
		retrieveTestVectorAs(t, vector2PublicKeyJWK2, &secpJWK)

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
		packet, err := didID.ToDNSPacket(*doc, []TypeIndex{1, 2, 3},
			[]AuthoritativeGateway{"gateway1.example-did-dht-gateway.com.", "gateway2.example-did-dht-gateway.com."}, nil)
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
						strings.Contains(s, strings.Join(expectedRecord.Record, "")) {
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
		didDHTDoc, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.NotEmpty(t, didDHTDoc.Types)
		require.Equal(t, didDHTDoc.Types, []TypeIndex{1, 2, 3})
		require.NotEmpty(t, didDHTDoc.Gateways)
		require.Equal(t, didDHTDoc.Gateways, []AuthoritativeGateway{"gateway1.example-did-dht-gateway.com.", "gateway2.example-did-dht-gateway.com."})
		require.Empty(t, didDHTDoc.PreviousDID)

		decodedDocJSON, err := json.Marshal(didDHTDoc.Doc)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedDIDDocJSON), string(decodedDocJSON))
	})

	// https://did-dht.com/#vector-3
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
			Services: []did.Service{
				{
					ID:              "service-1",
					Type:            "TestLongService",
					ServiceEndpoint: []string{"https://test-lllllllllllllllllllllllllllllllllllooooooooooooooooooooonnnnnnnnnnnnnnnnnnngggggggggggggggggggggggggggggggggggggsssssssssssssssssssssssssseeeeeeeeeeeeeeeeeeerrrrrrrrrrrrrrrvvvvvvvvvvvvvvvvvvvviiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiccccccccccccccccccccccccccccccceeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee.com/1"},
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
		previousDID := PreviousDID{
			PreviousDID: "did:dht:x3heus3ke8fhgb5pbecday9wtbfynd6m19q4pm6gcf5j356qhjzo",
			Signature:   "Tt9DRT6J32v7O2lzbfasW63_FfagiMHTHxtaEOD7p85zHE0r_EfiNleyL6BZGyB1P-oQ5p6_7KONaHAjr2K6Bw",
		}
		packet, err := didID.ToDNSPacket(*doc, nil, []AuthoritativeGateway{"gateway1.example-did-dht-gateway.com."}, &previousDID)
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
						strings.Contains(s, expectedRecord.TTL) {
						// make sure all parts of the record are contained within s
						for _, r := range expectedRecord.Record {
							if !strings.Contains(s, r) {
								break
							}
						}
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
		didDHTDoc, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.Empty(t, didDHTDoc.Types)
		require.NotEmpty(t, didDHTDoc.Gateways)
		require.Equal(t, didDHTDoc.Gateways, []AuthoritativeGateway{"gateway1.example-did-dht-gateway.com."})
		require.NotEmpty(t, didDHTDoc.PreviousDID)
		require.Equal(t, didDHTDoc.PreviousDID.PreviousDID.String(), "did:dht:x3heus3ke8fhgb5pbecday9wtbfynd6m19q4pm6gcf5j356qhjzo")

		decodedDocJSON, err := json.Marshal(didDHTDoc.Doc)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedDIDDocJSON), string(decodedDocJSON))
	})
}
