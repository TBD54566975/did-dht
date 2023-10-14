package util

import (
	"crypto/ed25519"

	"github.com/tv42/zbase32"
)

// Z32Encode returns the zbase32 representation of the input data.
func Z32Encode(data []byte) string {
	return zbase32.EncodeToString(data)
}

// Z32Decode returns the decoded zbase32 representation of the input data.
func Z32Decode(data string) ([]byte, error) {
	return zbase32.DecodeString(data)
}

// GenerateKeypair generates a public/private keypair using ed25519.
func GenerateKeypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}
