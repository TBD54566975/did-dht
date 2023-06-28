package db

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

const (
	identityNamespace = "identity"
	identityKey       = "service-identity"
)

type Identity struct {
	DID              string `json:"did"`
	PrivateKeyBase58 string `json:"privateKeyBase58"`
}

type IdentityStorage interface {
	WriteIdentity(did string, privateKey crypto.Ed25519PrivateKey) error
	ReadIdentity() (string, crypto.Ed25519PrivateKey, error)
}

func (s *Storage) WriteIdentity(did string, privateKey crypto.Ed25519PrivateKey) error {
	privKeyBytes, err := privateKey.Raw()
	if err != nil {
		return errors.WithMessage(err, "failed to get raw private key")
	}
	privKeyBase58 := base58.Encode(privKeyBytes)
	identity := &Identity{
		DID:              did,
		PrivateKeyBase58: privKeyBase58,
	}
	identityBytes, err := json.Marshal(identity)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal identity")
	}
	return s.Write(identityNamespace, identityKey, identityBytes)
}

func (s *Storage) ReadIdentity() (string, *crypto.Ed25519PrivateKey, error) {
	identity, err := s.Read(identityNamespace, identityKey)
	if err != nil {
		return "", nil, errors.WithMessage(err, "failed to read identity")
	}
	if len(identity) == 0 {
		return "", nil, nil
	}
	var gotIdentity Identity
	if err = json.Unmarshal(identity, &gotIdentity); err != nil {
		return "", nil, errors.WithMessage(err, "failed to unmarshal identity")
	}
	privKeyBytes, err := base58.Decode(gotIdentity.PrivateKeyBase58)
	if err != nil {
		return "", nil, errors.WithMessage(err, "failed to decode private key")
	}
	privKey, err := crypto.UnmarshalEd25519PrivateKey(privKeyBytes)
	if err != nil {
		return "", nil, errors.WithMessage(err, "failed to unmarshal private key")
	}
	return gotIdentity.DID, privKey.(*crypto.Ed25519PrivateKey), nil
}
