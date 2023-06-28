package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"log"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
)

func TestSignJWS(t *testing.T) {
	privkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("failed to generate private key: %s", err)
		return
	}

	headers := jws.NewHeaders()
	headers.Set(jws.KeyIDKey, "mykeyid")
	buf, err := jws.Sign([]byte("Lorem ipsum"), jws.WithKey(jwa.RS256, privkey, jws.WithProtectedHeaders(headers)))
	if err != nil {
		log.Printf("failed to created JWS message: %s", err)
		return
	}

	// When you receive a JWS message, you can verify the signature
	// and grab the payload sent in the message in one go:
	verified, err := jws.Verify(buf, jws.WithKey(jwa.RS256, &privkey.PublicKey))
	if err != nil {
		log.Printf("failed to verify message: %s", err)
		return
	}
	println(string(buf))
	println(string(verified))
}
