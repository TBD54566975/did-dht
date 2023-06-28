package internal

import (
	"strings"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/pkg/errors"
)

// GetMethodForDID gets a DID method from a did, the second part of the did (e.g. did:test:abcd, the method is 'test')
func GetMethodForDID(id string) (did.Method, error) {
	split := strings.Split(id, ":")
	if len(split) < 3 {
		return "", errors.New("malformed did: did has fewer than three parts")
	}
	if split[0] != "did" {
		return "", errors.New("malformed did: did must start with `did`")
	}
	return did.Method(split[1]), nil
}
