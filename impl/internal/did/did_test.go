package did

import (
	"crypto/ed25519"
	"fmt"
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/cryptosuite"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/goccy/go-json"
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
		packet, err := didID.ToDNSPacket(*doc, nil, nil, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		pb, _ := packet.Pack()
		println("1 - DNS Length: ", len(pb))

		didDHTDoc, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.Empty(t, didDHTDoc.Types)
		require.Empty(t, didDHTDoc.Gateways)
		require.Empty(t, didDHTDoc.PreviousDID)

		jsonDoc, err := json.Marshal(doc)
		require.NoError(t, err)

		jsonDecodedDoc, err := json.Marshal(didDHTDoc.Doc)
		require.NoError(t, err)

		assert.JSONEq(t, string(jsonDoc), string(jsonDecodedDoc))
	})

	t.Run("simple doc - test to dns packet round trip - cbor", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		didID := DHT(doc.ID)
		packet, err := didID.ToCBOR(*doc, nil, nil, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		println("1 - CBOR Length: ", len(packet))

		didDHTDoc, err := didID.FromCBOR(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.Empty(t, didDHTDoc.Types)
		require.Empty(t, didDHTDoc.Gateways)
		require.Empty(t, didDHTDoc.PreviousDID)

		jsonDoc, err := json.Marshal(doc)
		require.NoError(t, err)

		jsonDecodedDoc, err := json.Marshal(didDHTDoc.Doc)
		require.NoError(t, err)

		assert.JSONEq(t, string(jsonDoc), string(jsonDecodedDoc))
	})

	t.Run("doc with types and a gateway - test to dns packet round trip", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		didID := DHT(doc.ID)
		packet, err := didID.ToDNSPacket(*doc, []TypeIndex{1, 2, 3}, []AuthoritativeGateway{"gateway1.example-did-dht-gateway.com."}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		pb, _ := packet.Pack()
		println("2 - DNS Length: ", len(pb))

		didDHTDoc, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.NotEmpty(t, didDHTDoc.Types)
		require.Empty(t, didDHTDoc.PreviousDID)
		require.Equal(t, didDHTDoc.Types, []TypeIndex{1, 2, 3})
		require.NotEmpty(t, didDHTDoc.Gateways)
		require.Equal(t, didDHTDoc.Gateways, []AuthoritativeGateway{"gateway1.example-did-dht-gateway.com."})

		assert.EqualValues(t, *doc, didDHTDoc.Doc)
	})

	t.Run("doc with types and a gateway - test to dns packet round trip - cbor", func(t *testing.T) {
		privKey, doc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, privKey)
		require.NotEmpty(t, doc)

		didID := DHT(doc.ID)
		packet, err := didID.ToCBOR(*doc, []TypeIndex{1, 2, 3}, []AuthoritativeGateway{"gateway1.example-did-dht-gateway.com."}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		println("2 - CBOR Length: ", len(packet))

		didDHTDoc, err := didID.FromCBOR(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.NotEmpty(t, didDHTDoc.Types)
		require.Empty(t, didDHTDoc.PreviousDID)
		require.Equal(t, didDHTDoc.Types, []TypeIndex{1, 2, 3})
		require.NotEmpty(t, didDHTDoc.Gateways)
		require.Equal(t, didDHTDoc.Gateways, []AuthoritativeGateway{"gateway1.example-did-dht-gateway.com."})

		assert.EqualValues(t, *doc, didDHTDoc.Doc)
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
					ServiceEndpoint: []string{"https://example.com/vc/"},
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
		packet, err := didID.ToDNSPacket(*doc, nil, nil, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		pb, _ := packet.Pack()
		println("3 - DNS Length: ", len(pb))

		didDHTDoc, err := didID.FromDNSPacket(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.Empty(t, didDHTDoc.Types)
		require.Empty(t, didDHTDoc.Gateways)
		require.Empty(t, didDHTDoc.PreviousDID)

		decodedJSON, err := json.Marshal(didDHTDoc.Doc)
		require.NoError(t, err)

		docJSON, err := json.Marshal(doc)
		require.NoError(t, err)

		assert.JSONEq(t, string(docJSON), string(decodedJSON))
	})

	t.Run("doc with multiple keys and services - test to dns packet round trip - cbor", func(t *testing.T) {
		pubKey, _, err := crypto.GenerateSECP256k1Key()
		require.NoError(t, err)
		pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK(nil, pubKey)
		require.NoError(t, err)

		opts := CreateDIDDHTOpts{
			VerificationMethods: []VerificationMethod{
				{
					VerificationMethod: did.VerificationMethod{
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
					ServiceEndpoint: []string{"https://example.com/vc/"},
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
		packet, err := didID.ToCBOR(*doc, nil, nil, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		println("3 - CBOR Length: ", len(packet))

		didDHTDoc, err := didID.FromCBOR(packet)
		require.NoError(t, err)
		require.NotEmpty(t, didDHTDoc)
		require.NotEmpty(t, didDHTDoc.Doc)
		require.Empty(t, didDHTDoc.Types)
		require.Empty(t, didDHTDoc.Gateways)
		require.Empty(t, didDHTDoc.PreviousDID)

		decodedJSON, err := json.Marshal(didDHTDoc.Doc)
		require.NoError(t, err)

		docJSON, err := json.Marshal(doc)
		require.NoError(t, err)

		assert.JSONEq(t, string(docJSON), string(decodedJSON))
	})
}

func TestDIDDHTFeatures(t *testing.T) {
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

	t.Run("prev", func(t *testing.T) {
		previousPrivKey, previousDoc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, previousPrivKey)
		require.NotEmpty(t, previousDoc)

		previousDIDDHT := DHT(previousDoc.ID)
		previousDID, err := CreatePreviousDIDRecord(previousPrivKey, previousDIDDHT, "did:dht:sr6jgmcc84xig18ix66qbiwnzeiumocaaybh13f5w97bfzus4pcy")
		assert.NoError(t, err)
		assert.NotEmpty(t, previousDID)

		println(previousDID.PreviousDID.String())
		println(previousDID.Signature)
	})

	t.Run("Test Previous DID", func(t *testing.T) {
		// generate previous DID
		previousPrivKey, previousDoc, err := GenerateDIDDHT(CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, previousPrivKey)
		require.NotEmpty(t, previousDoc)

		// generate new DID
		doc, err := CreateDIDDHTDID(pubKey.(ed25519.PublicKey), CreateDIDDHTOpts{})
		assert.NoError(t, err)
		assert.NotEmpty(t, doc)

		// set previous DID signature
		previousDIDDHT := DHT(previousDoc.ID)
		currentDIDDHT := DHT(doc.ID)
		previousDID, err := CreatePreviousDIDRecord(previousPrivKey, previousDIDDHT, currentDIDDHT)
		assert.NoError(t, err)
		assert.NotEmpty(t, previousDID)
		assert.NotEmpty(t, previousDID.PreviousDID)
		assert.Equal(t, previousDID.PreviousDID.String(), previousDoc.ID)
		assert.NotEmpty(t, previousDID.Signature)

		// validate previous DID signature
		err = ValidatePreviousDIDSignatureValid(currentDIDDHT, *previousDID)
		assert.NoError(t, err)

		// construct the DNS packet with the previous DID entry
		dnsMsg, err := currentDIDDHT.ToDNSPacket(*doc, nil, nil, previousDID)
		assert.NoError(t, err)
		assert.NotEmpty(t, dnsMsg)

		// parse the DNS packet
		didDHTDoc, err := currentDIDDHT.FromDNSPacket(dnsMsg)
		assert.NoError(t, err)
		assert.NotEmpty(t, didDHTDoc)
		assert.NotEmpty(t, didDHTDoc.Doc)
		assert.NotEmpty(t, didDHTDoc.PreviousDID)
		assert.Equal(t, didDHTDoc.PreviousDID.PreviousDID.String(), previousDoc.ID)
		assert.NotEmpty(t, didDHTDoc.PreviousDID.Signature)

		// validate previous DID signature with wrong DID
		err = ValidatePreviousDIDSignatureValid("did:dht:123456789abcdefghi", *previousDID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get identity key from the current DID")

		// validate previous DID signature with wrong signature
		previousDID.Signature = "wrong signature"
		err = ValidatePreviousDIDSignatureValid(currentDIDDHT, *previousDID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode the previous DID's signature")
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
