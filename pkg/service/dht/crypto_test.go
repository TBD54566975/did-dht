package dht

import (
	"testing"

	sdkcrypto "github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/did/jwk"
	"github.com/stretchr/testify/assert"

	"did-dht/pkg/service/gossip"
)

func TestSignVerifyRecordJWS(t *testing.T) {
	privKey, jwk, err := jwk.GenerateDIDJWK(sdkcrypto.Ed25519)
	assert.NoError(t, err)

	didDoc, err := jwk.Expand()
	assert.NoError(t, err)

	id := didDoc.ID
	kid := didDoc.VerificationMethod[0].ID

	signer, err := jwx.NewJWXSigner(id, kid, privKey)
	assert.NoError(t, err)

	record := gossip.Record{
		DID:      id,
		Endpoint: "http://tbd.dev",
	}
	signedRecord, err := SignRecordJWS(*signer, record)
	assert.NoError(t, err)
	assert.NotEmpty(t, signedRecord.JWS)

	pubKeyJWK := didDoc.VerificationMethod[0].PublicKeyJWK
	verifier, err := jwx.NewJWXVerifierFromJWK(id, *pubKeyJWK)
	assert.NoError(t, err)

	err = VerifyRecordJWS(*verifier, *signedRecord)
	assert.NoError(t, err)
}
