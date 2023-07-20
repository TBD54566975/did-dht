package dht

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/internal/record"
	"did-dht/internal/resolution"
)

// VerifyRecord verifies a record by resolving the DID and checking the JWS, returning an error if the record is invalid.
func VerifyRecord(ctx context.Context, resolver *resolution.ServiceResolver, r Record) error {
	resolved, err := resolver.Resolve(ctx, r.DID)
	if err != nil {
		logrus.WithError(err).Warnf("failed to resolve DID: %s, attemtping to get iss", r.DID)
		parsedJWT, err := jwt.Parse([]byte(r.JWS), jwt.WithVerify(false))
		if err != nil {
			return errors.Wrapf(err, "parsing JWS")
		}
		if parsedJWT.Issuer() != "" && parsedJWT.Issuer() != r.DID {
			logrus.Infof("iss is not the same as DID, attempting to resolve: %s", parsedJWT.Issuer())
			resolved, err = resolver.Resolve(ctx, parsedJWT.Issuer())
			if err != nil {
				return errors.Wrapf(err, "resolving issuer DID: %s", parsedJWT.Issuer())
			}
			logrus.Infof("resolved iss DID: %s", resolved.Document.ID)
		}
	}

	if resolved.Document.IsEmpty() {
		return errors.Errorf("resolved DID is empty: %s", r.DID)
	}

	// decode JWS and get the KID
	headers, err := jwx.GetJWSHeaders([]byte(r.JWS))
	if err != nil {
		return errors.Wrapf(err, "getting JWS headers")
	}
	kid, ok := headers.Get(jws.KeyIDKey)
	if !ok {
		return errors.New("JWS missing kid")
	}
	kidStr, ok := kid.(string)
	if !ok {
		return errors.New("kid is not a string")
	}

	pubKey, err := did.GetKeyFromVerificationMethod(resolved.Document, kidStr)
	if err != nil {
		return errors.Wrapf(err, "getting verification information from DID Document: %s, for KID: %s", r.DID, kidStr)
	}

	// verify the JWS
	verifier, err := jwx.NewJWXVerifier(r.DID, kidStr, pubKey)
	if err != nil {
		return errors.Wrapf(err, "creating JWS verifier")
	}
	jsonMap, err := util.ToJSONMap(r)
	if err != nil {
		return errors.Wrapf(err, "converting payload to JSON map")
	}
	return record.VerifyRecordJWSFromJSON(*verifier, jsonMap)
}
