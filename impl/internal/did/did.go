package did

import (
	"crypto/ed25519"
	"fmt"
	"strings"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/cryptosuite"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/did/ion"
	"github.com/miekg/dns"
	"github.com/tv42/zbase32"
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

type CreateDIDDHTOpts struct {
	// VerificationMethods is a list of verification methods to include in the DID Document
	// Cannot contain id #0 which is reserved for the identity key
	VerificationMethods []VerificationMethods
	// Services is a list of services to include in the DID Document
	Services []did.Service
}

type VerificationMethods struct {
	VerificationMethod did.VerificationMethod `json:"verificationMethod"`
	Purposes           []ion.PublicKeyPurpose `json:"purposes"`
}

// GenerateDIDDHT generates a did:dht identifier given a set of options
func GenerateDIDDHT(opts CreateDIDDHTOpts) (ed25519.PrivateKey, *did.Document, error) {
	// generate the identity key
	pubKey, privKey, err := crypto.GenerateEd25519Key()
	if err != nil {
		return nil, nil, err
	}

	doc, err := CreateDIDDHTDID(pubKey, opts)
	return privKey, doc, err
}

// CreateDIDDHTDID creates a did:dht identifier for the given ed25519 public key
func CreateDIDDHTDID(pubKey ed25519.PublicKey, opts CreateDIDDHTOpts) (*did.Document, error) {
	// generate the did:dht identifier
	id := GetDIDDHTIdentifier(pubKey)

	// validate opts and build verification methods, key purposes, and services
	var vms []did.VerificationMethod
	var authentication []did.VerificationMethodSet
	var assertionMethod []did.VerificationMethodSet
	var keyAgreement []did.VerificationMethodSet
	var capabilityInvocation []did.VerificationMethodSet
	var capabilityDelegation []did.VerificationMethodSet
	if len(opts.VerificationMethods) > 0 {
		seenIDs := make(map[string]bool)
		for _, vm := range opts.VerificationMethods {
			if vm.VerificationMethod.ID == "#0" || vm.VerificationMethod.ID == "0" {
				return nil, fmt.Errorf("verification method id #0 is reserved for the identity key")
			}
			if seenIDs[vm.VerificationMethod.ID] {
				return nil, fmt.Errorf("verification method id %s is not unique", vm.VerificationMethod.ID)
			}
			if vm.VerificationMethod.Type != cryptosuite.JSONWebKey2020Type {
				return nil, fmt.Errorf("verification method type %s is not supported", vm.VerificationMethod.Type)
			}
			if vm.VerificationMethod.PublicKeyJWK == nil {
				return nil, fmt.Errorf("verification method public key jwk is required")
			}

			// mark as seen
			seenIDs[vm.VerificationMethod.ID] = true

			// update ID and controller in place
			vm.VerificationMethod.ID = id + "#" + vm.VerificationMethod.ID
			vm.VerificationMethod.Controller = id
			vms = append(vms, vm.VerificationMethod)

			// add purposes
			for _, purpose := range vm.Purposes {
				switch purpose {
				case ion.Authentication:
					authentication = append(authentication, vm.VerificationMethod.ID)
				case ion.AssertionMethod:
					assertionMethod = append(assertionMethod, vm.VerificationMethod.ID)
				case ion.KeyAgreement:
					keyAgreement = append(keyAgreement, vm.VerificationMethod.ID)
				case ion.CapabilityInvocation:
					capabilityInvocation = append(capabilityInvocation, vm.VerificationMethod.ID)
				case ion.CapabilityDelegation:
					capabilityDelegation = append(capabilityDelegation, vm.VerificationMethod.ID)
				default:
					return nil, fmt.Errorf("unknown key purpose: %s:%s", vm.VerificationMethod.ID, purpose)
				}
			}
		}
	}
	if len(opts.Services) > 0 {
		seenIDs := make(map[string]bool)
		for _, s := range opts.Services {
			if seenIDs[s.ID] {
				return nil, fmt.Errorf("service id %s is not unique", s.ID)
			}

			// mark as seen
			seenIDs[s.ID] = true

			// update ID in place
			s.ID = id + "#" + s.ID
		}
	}

	// create the did document
	key0JWK, err := jwx.PublicKeyToPublicKeyJWK("0", pubKey)
	if err != nil {
		return nil, err
	}
	vm0 := did.VerificationMethod{
		ID:           id + "#0",
		Type:         cryptosuite.JSONWebKey2020Type,
		Controller:   id,
		PublicKeyJWK: key0JWK,
	}
	return &did.Document{
		Context:              []string{did.KnownDIDContext},
		ID:                   id,
		VerificationMethod:   append([]did.VerificationMethod{vm0}, vms...),
		Services:             opts.Services,
		Authentication:       authentication,
		AssertionMethod:      assertionMethod,
		KeyAgreement:         keyAgreement,
		CapabilityInvocation: capabilityInvocation,
		CapabilityDelegation: capabilityDelegation,
	}, nil
}

// GetDIDDHTIdentifier returns the did:dht identifier for the given public key
func GetDIDDHTIdentifier(pubKey []byte) string {
	return strings.Join([]string{Prefix, zbase32.EncodeToString(pubKey)}, ":")
}

func (d DHT) ToDNSPacket(did did.Document) ([]dns.RR, error) {
	if d.String() != did.ID {
		return nil, fmt.Errorf("did and dht id mismatch")
	}
	return nil, nil
}

func (d DHT) FromDNSPacket() (*did.Document, error) {
	return nil, nil
}
