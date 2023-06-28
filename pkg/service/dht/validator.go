package dht

import (
	"encoding/json"
	"fmt"

	record "github.com/libp2p/go-libp2p-record"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/internal"
)

var _ record.Validator = (*Validator)(nil)

type Validator struct {
	ns string
}

func NewValidator(ns string) Validator {
	return Validator{ns: ns}
}

func (v Validator) Validate(key string, value []byte) error {
	ns, key, err := record.SplitKey(key)
	if err != nil {
		return err
	}
	if ns != v.ns {
		return fmt.Errorf("namespace not %s", v.ns)
	}

	// validate the key
	if !internal.IsValidDID(key) {
		return fmt.Errorf("key did not contain a valid DID: %s", key)
	}

	// validate the value
	var r Record
	if err = json.Unmarshal(value, &r); err != nil {
		return errors.WithMessage(err, "failed to unmarshal record")
	}

	if r.DID != key {
		return fmt.Errorf("DID<%s> does not match key<%s>", r.DID, key)
	}
	if r.Endpoint == "" {
		return fmt.Errorf("endpoint is empty")
	}

	// TODO(gabe): enable signature validation
	if r.JWS == "" {
		logrus.Warn("JWS is empty")
	}
	return nil
}

// Select conforms to the Validator interface, it always returns 0 as all records are equivalently valid.
// TODO(gabe): for now just choose the first one, in the future discern the most recent/valid
func (v Validator) Select(_ string, _ [][]byte) (int, error) {
	return 0, nil
}
