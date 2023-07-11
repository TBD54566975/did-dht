package record

import (
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

// SignRecordJWS signs a record by creating a JWS, returning the signed record.
func SignRecordJWS(signer jwx.Signer, payload map[string]any) (*SignedRecord, error) {
	if _, ok := payload["jws"]; ok {
		return nil, errors.New("payload already has a JWS")
	}
	recordJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal record")
	}
	var record map[string]any
	if err = json.Unmarshal(recordJSON, &record); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal record")
	}
	jwsBytes, err := signer.SignWithDefaults(record)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign record")
	}
	return &SignedRecord{
		Payload: payload,
		JWS:     string(jwsBytes),
	}, nil
}

// VerifyRecordJWSFromJSON verifies a record by checking the JWS from a JSON map, returning an error
// if the record is invalid.
func VerifyRecordJWSFromJSON(verifier jwx.Verifier, record map[string]any) error {
	if _, ok := record["jws"]; !ok {
		return errors.New("record does not have a JWS")
	}
	jws, ok := record["jws"].(string)
	if !ok {
		return errors.New("jws is not a string")
	}
	delete(record, "jws")
	return VerifyRecordJWS(verifier, SignedRecord{
		Payload: record,
		JWS:     jws,
	})
}

// VerifyRecordJWS verifies a record by checking the JWS, returning an error if the record is invalid.
func VerifyRecordJWS(verifier jwx.Verifier, record SignedRecord) error {
	if record.JWS == "" {
		return errors.New("record does not have a JWS")
	}

	_, token, err := verifier.VerifyAndParse(record.JWS)
	if err != nil {
		return errors.Wrap(err, "failed to verify record")
	}

	// make sure there are no extra properties in the token
	if len(record.Payload) != len(token.PrivateClaims()) {
		return errors.New("record and token have a mismatched number of properties")
	}

	// verify all properties in the token have been signed over and they match
	for k, v := range record.Payload {
		gotV, ok := token.Get(k)
		if !ok {
			return errors.Errorf("record property not signed over: %s", k)
		}
		if gotV != v {
			return errors.Errorf("record property value does not match for key: %s", k)
		}
	}
	return nil
}
