package did

import (
	"fmt"
	"strings"

	"github.com/TBD54566975/ssi-sdk/did"
)

type (
	DHT string
)

const (
	// Prefix did:dht prefix
	Prefix               = "did:dht"
	DHTMethod did.Method = "dht"
)

func (d DHT) IsValid() bool {
	return true
}

func (d DHT) String() string {
	return string(d)
}

// Suffix returns the value without the `did:dht` prefix
func (d DHT) Suffix() (string, error) {
	if suffix, ok := strings.CutPrefix(string(d), Prefix+":"); ok {
		return suffix, nil
	}
	return "", fmt.Errorf("invalid did:dht: %s", d)
}

func (DHT) Method() did.Method {
	return DHTMethod
}

func GenerateDIDDHT() {

}

func CreateDIDDHTDID(pubKey []byte) (string, error) {
	return "", nil
}
