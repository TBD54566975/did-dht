package dht

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/internal/resolution"
	"did-dht/pkg/service/gossip"
)

// SignRecordJWS signs a record by creating a JWS, returning the signed record.
func SignRecordJWS(signer jwx.Signer, record gossip.Record) (*gossip.Record, error) {
	if record.JWS != "" {
		return nil, errors.New("record already has a JWS")
	}
	if record.DID == "" {
		return nil, errors.New("record must have a DID")
	}
	if record.Endpoint == "" {
		return nil, errors.New("record must have an endpoint")
	}
	recordJSON := map[string]any{
		"did":      record.DID,
		"endpoint": record.Endpoint,
	}
	jwsBytes, err := signer.SignWithDefaults(recordJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign record")
	}
	record.JWS = string(jwsBytes)
	return &record, nil
}

// VerifyRecordJWS verifies a record by checking the JWS, returning an error if the record is invalid.
func VerifyRecordJWS(verifier jwx.Verifier, record gossip.Record) error {
	if record.JWS == "" {
		return errors.New("record does not have a JWS")
	}
	if record.DID == "" {
		return errors.New("record must have a DID")
	}
	if record.Endpoint == "" {
		return errors.New("record must have an endpoint")
	}

	// TODO(gabe): consider checking whether the JWS was signed by a KID in the document
	_, token, err := verifier.VerifyAndParse(record.JWS)
	if err != nil {
		return errors.Wrap(err, "failed to verify record")
	}

	// check the DID property exists
	gotDID, ok := token.Get("did")
	if !ok {
		return errors.New("record does not have a DID")
	}
	didStr, ok := gotDID.(string)
	if !ok {
		return errors.New("record DID is not a string")
	}
	if didStr != record.DID {
		return fmt.Errorf("record DID<%s> does not match DID in JWS<%s>", record.DID, didStr)
	}

	// check the endpoint property exists
	gotEndpoint, ok := token.Get("endpoint")
	if !ok {
		return errors.New("record does not have an endpoint")
	}
	endpointStr, ok := gotEndpoint.(string)
	if !ok {
		return errors.New("record endpoint is not a string")
	}
	if endpointStr != record.Endpoint {
		return fmt.Errorf("record endpoint<%s> does not match endpoint in JWS<%s>", record.Endpoint, endpointStr)
	}
	return nil
}

// VerifyRecord verifies a record by resolving the DID and checking the JWS, returning an error if the record is invalid.
func VerifyRecord(ctx context.Context, resolver *resolution.ServiceResolver, r gossip.Record) error {
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
	return VerifyRecordJWS(*verifier, r)
}
