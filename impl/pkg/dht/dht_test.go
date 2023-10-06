package dht

import (
	"context"
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/util"
)

func TestGetPutDHT(t *testing.T) {
	d, err := NewDHT()
	require.NoError(t, err)

	records := [][]any{
		{"foo", "bar"},
	}
	pubKey, privKey, err := util.GenerateKeypair()
	require.NoError(t, err)

	putReq, err := CreatePutRequest(pubKey, privKey, records)
	require.NoError(t, err)

	id, err := d.Put(context.Background(), pubKey, *putReq)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	got, err := d.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, got)
}

func TestDID(t *testing.T) {
	pubKey, privKey, err := util.GenerateKeypair()
	require.NoError(t, err)

	pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK("#0", pubKey)
	require.NoError(t, err)

	id := util.Z32Encode(pubKey)

	records := [][]any{
		{"_did", map[string]any{
			"id": "did:pk" + id,
			"verificationMethod": []map[string]any{
				{
					"id":         "#0",
					"controller": "did:pk" + id,
					"type":       "JsonWebKey2020",
					"publicKeyJwk": map[string]any{
						"kty": "OKP",
						"crv": "Ed25519",
						"x":   pubKeyJWK.X,
					},
				},
			},
		}},
	}

	putReq, err := CreatePutRequest(pubKey, privKey, records)
	require.NoError(t, err)

	d, err := NewDHT()
	require.NoError(t, err)

	_, err = d.Put(context.Background(), pubKey, *putReq)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	got, err := d.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, got)
}
