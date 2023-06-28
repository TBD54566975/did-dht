package internal

import (
	"fmt"

	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/pkg/errors"

	"did-dht/pkg/service"
)

// SignRecordJWS signs a record by creating a JWS, returning the signed record.
func SignRecordJWS(signer jwx.Signer, record service.Record) (*service.Record, error) {
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
func VerifyRecordJWS(verifier jwx.Verifier, record service.Record) error {
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
