package did

import (
	"crypto/ed25519"
	"encoding/base64"
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
	prefix, err := d.Suffix()
	if err != nil {
		return false
	}
	pk, err := zbase32.DecodeString(string(d)[len(prefix)+1:])
	return err == nil && len(pk) == ed25519.PublicKeySize
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
	VerificationMethods []VerificationMethod
	// Services is a list of services to include in the DID Document
	Services []did.Service
}

type VerificationMethod struct {
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
	var keyAgreement []did.VerificationMethodSet
	authentication := []did.VerificationMethodSet{"#0"}
	assertionMethod := []did.VerificationMethodSet{"#0"}
	capabilityInvocation := []did.VerificationMethodSet{"#0"}
	capabilityDelegation := []did.VerificationMethodSet{"#0"}
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
			if vm.VerificationMethod.ID == "" || strings.Contains(vm.VerificationMethod.ID, "#") {
				return nil, fmt.Errorf("verification method id %s is invalid", vm.VerificationMethod.ID)
			}
			vm.VerificationMethod.ID = id + "#" + vm.VerificationMethod.ID
			if vm.VerificationMethod.Controller != "" {
				vm.VerificationMethod.Controller = id
			}
			vms = append(vms, vm.VerificationMethod)

			// add purposes
			vmID := vm.VerificationMethod.ID[strings.LastIndex(vm.VerificationMethod.ID, "#"):]
			for _, purpose := range vm.Purposes {
				switch purpose {
				case ion.Authentication:
					authentication = append(authentication, vmID)
				case ion.AssertionMethod:
					assertionMethod = append(assertionMethod, vmID)
				case ion.KeyAgreement:
					keyAgreement = append(keyAgreement, vmID)
				case ion.CapabilityInvocation:
					capabilityInvocation = append(capabilityInvocation, vmID)
				case ion.CapabilityDelegation:
					capabilityDelegation = append(capabilityDelegation, vmID)
				default:
					return nil, fmt.Errorf("unknown key purpose: %s:%s", vmID, purpose)
				}
			}
		}
	}
	if len(opts.Services) > 0 {
		seenIDs := make(map[string]bool)
		for i, s := range opts.Services {
			if seenIDs[s.ID] {
				return nil, fmt.Errorf("service id %s is not unique", s.ID)
			}

			// mark as seen
			seenIDs[s.ID] = true

			// update ID in place
			opts.Services[i].ID = id + "#" + s.ID
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

// ToDNSPacket converts a DID DHT Document to a DNS packet
func (d DHT) ToDNSPacket(doc did.Document) (*dns.Msg, error) {
	var records []dns.RR
	var rootRecord []string
	keyLookup := make(map[string]string)

	// build all key records
	var vmIDs []string
	for i, vm := range doc.VerificationMethod {
		recordIdentifier := fmt.Sprintf("k%d", i)
		vmID := vm.ID
		if strings.Contains(vmID, "#") {
			vmID = vmID[strings.LastIndex(vm.ID, "#")+1:]
		}
		keyLookup[vm.ID] = recordIdentifier

		var keyType int
		switch vm.PublicKeyJWK.ALG {
		case "EdDSA":
			keyType = 0
		case "ES256K":
			keyType = 1
		default:
			return nil, fmt.Errorf("unsupported key type: %s", vm.PublicKeyJWK.ALG)
		}

		// convert the public key to a base64url encoded string
		pubKey, err := vm.PublicKeyJWK.ToPublicKey()
		if err != nil {
			return nil, err
		}
		pubKeyBytes, err := crypto.PubKeyToBytes(pubKey)
		if err != nil {
			return nil, err
		}
		keyBase64Url := base64.RawURLEncoding.EncodeToString(pubKeyBytes)

		keyRecord := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("_%s._did.", recordIdentifier),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: []string{fmt.Sprintf("id=%s,t=%d,k=%s", vmID, keyType, keyBase64Url)},
		}

		records = append(records, &keyRecord)
		vmIDs = append(vmIDs, recordIdentifier)
	}
	// add verification methods to the root record
	rootRecord = append(rootRecord, fmt.Sprintf("vm=%s", strings.Join(vmIDs, ",")))

	var svcIDs []string
	for i, service := range doc.Services {
		recordIdentifier := fmt.Sprintf("s%d", i)
		sID := service.ID
		if strings.Contains(sID, "#") {
			sID = sID[strings.LastIndex(service.ID, "#")+1:]
		}

		serviceRecord := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("_%s._did.", recordIdentifier),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: []string{fmt.Sprintf("id=%s,t=%s,uri=%s", sID, service.Type, service.ServiceEndpoint)},
		}

		records = append(records, &serviceRecord)
		svcIDs = append(svcIDs, recordIdentifier)
	}
	// add services to the root record
	if len(svcIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("svc=%s", strings.Join(svcIDs, ",")))
	}

	// add verification relationships to the root record
	var authIDs []string
	for _, auth := range doc.Authentication {
		authIDs = append(authIDs, keyLookup[doc.ID+auth.(string)])
	}
	if len(authIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("auth=%s", strings.Join(authIDs, ",")))
	}

	var assertionIDs []string
	for _, assertion := range doc.AssertionMethod {
		assertionIDs = append(assertionIDs, keyLookup[doc.ID+assertion.(string)])
	}
	if len(assertionIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("asm=%s", strings.Join(assertionIDs, ",")))
	}

	var keyAgreementIDs []string
	for _, keyAgreement := range doc.KeyAgreement {
		keyAgreementIDs = append(keyAgreementIDs, keyLookup[doc.ID+keyAgreement.(string)])
	}
	if len(keyAgreementIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("agm=%s", strings.Join(keyAgreementIDs, ",")))
	}

	var capabilityInvocationIDs []string
	for _, capabilityInvocation := range doc.CapabilityInvocation {
		capabilityInvocationIDs = append(capabilityInvocationIDs, keyLookup[doc.ID+capabilityInvocation.(string)])
	}
	if len(capabilityInvocationIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("inv=%s", strings.Join(capabilityInvocationIDs, ",")))
	}

	var capabilityDelegationIDs []string
	for _, capabilityDelegation := range doc.CapabilityDelegation {
		capabilityDelegationIDs = append(capabilityDelegationIDs, keyLookup[doc.ID+capabilityDelegation.(string)])
	}
	if len(capabilityDelegationIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("del=%s", strings.Join(capabilityDelegationIDs, ",")))
	}

	// add the root record
	rootAnswer := dns.TXT{
		Hdr: dns.RR_Header{
			Name:   "_did.",
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    7200,
		},
		Txt: []string{strings.Join(rootRecord, ";")},
	}
	records = append(records, &rootAnswer)

	// build the dns packet
	return &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:            0,
			Response:      true,
			Authoritative: true,
		},
		Answer: records,
	}, nil
}

// FromDNSPacket converts a DNS packet to a DID DHT Document
func (d DHT) FromDNSPacket(msg *dns.Msg) (*did.Document, error) {
	doc := did.Document{
		ID: d.String(),
	}

	keyLookup := make(map[string]string)
	for _, rr := range msg.Answer {
		switch record := rr.(type) {
		case *dns.TXT:
			if strings.HasPrefix(record.Hdr.Name, "_k") {
				data := parseTxtData(strings.Join(record.Txt, ","))
				vmID := data["id"]
				keyType := keyTypeLookUp(data["t"])
				keyBase64URL := data["k"]

				// Convert keyBase64URL back to PublicKeyJWK
				pubKeyBytes, err := base64.RawURLEncoding.DecodeString(keyBase64URL)
				if err != nil {
					return nil, err
				}
				pubKey, err := crypto.BytesToPubKey(pubKeyBytes, keyType)
				if err != nil {
					return nil, err
				}
				pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK(vmID, pubKey)
				if err != nil {
					return nil, err
				}

				vm := did.VerificationMethod{
					ID:           d.String() + "#" + vmID,
					Type:         cryptosuite.JSONWebKey2020Type,
					Controller:   d.String(),
					PublicKeyJWK: pubKeyJWK,
				}
				doc.VerificationMethod = append(doc.VerificationMethod, vm)

				// add to key lookup (e.g.  "k1" -> "#key1")
				keyLookup[strings.Split(record.Hdr.Name, ".")[0][1:]] = "#" + vmID
			} else if strings.HasPrefix(record.Hdr.Name, "_s") {
				data := parseTxtData(strings.Join(record.Txt, ","))
				sID := data["id"]
				serviceType := data["t"]
				serviceEndpoint := data["uri"]

				service := did.Service{
					ID:              d.String() + "#" + sID,
					Type:            serviceType,
					ServiceEndpoint: serviceEndpoint,
				}
				doc.Services = append(doc.Services, service)

			} else if record.Hdr.Name == "_did" {
				rootData := strings.Join(record.Txt, ";")
				rootItems := strings.Split(rootData, ";")

				for _, item := range rootItems {
					kv := strings.Split(item, "=")
					if len(kv) != 2 {
						continue
					}

					key, values := kv[0], kv[1]
					valueItems := strings.Split(values, ",")

					switch key {
					case "vm":
						// These are already processed in the "_k" prefix case
						continue
					case "srv":
						// These are already processed in the "_s" prefix case
						continue
					case "auth":
						for _, valueItem := range valueItems {
							doc.Authentication = append(doc.Authentication, keyLookup[valueItem])
						}
					case "asm":
						for _, valueItem := range valueItems {
							doc.AssertionMethod = append(doc.AssertionMethod, keyLookup[valueItem])
						}
					case "agm":
						for _, valueItem := range valueItems {
							doc.KeyAgreement = append(doc.KeyAgreement, keyLookup[valueItem])
						}
					case "inv":
						for _, valueItem := range valueItems {
							doc.CapabilityInvocation = append(doc.CapabilityInvocation, keyLookup[valueItem])
						}
					case "del":
						for _, valueItem := range valueItems {
							doc.CapabilityDelegation = append(doc.CapabilityDelegation, keyLookup[valueItem])
						}
					}
				}
			}
		}
	}

	return &doc, nil
}

func parseTxtData(data string) map[string]string {
	pairs := strings.Split(data, ",")
	result := make(map[string]string)
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func keyTypeLookUp(keyType string) crypto.KeyType {
	switch keyType {
	case "0":
		return crypto.Ed25519
	case "1":
		return crypto.SECP256k1
	default:
		return ""
	}
}
