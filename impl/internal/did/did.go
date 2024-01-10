package did

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/cryptosuite"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/miekg/dns"
	"github.com/tv42/zbase32"
)

type (
	DHT       string
	TypeIndex int
)

const (
	// Prefix did:dht prefix
	Prefix                               = "did:dht"
	DHTMethod      did.Method            = "dht"
	JSONWebKeyType cryptosuite.LDKeyType = "JsonWebKey"

	Discoverable           TypeIndex = 0
	Organization           TypeIndex = 1
	GovernmentOrganization TypeIndex = 2
	Corporation            TypeIndex = 3
	LocalBusiness          TypeIndex = 4
	SoftwarePackage        TypeIndex = 5
	WebApplication         TypeIndex = 6
	FinancialInstitution   TypeIndex = 7
)

func (d DHT) IsValid() bool {
	suffix, err := d.Suffix()
	if err != nil {
		return false
	}
	pk, err := zbase32.DecodeString(suffix)
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
	return "", fmt.Errorf("invalid did:dht prefix: %s", d)
}

func (DHT) Method() did.Method {
	return DHTMethod
}

type CreateDIDDHTOpts struct {
	// Controller is the DID Controller, can be a list of DIDs
	Controller []string
	// AlsoKnownAs is a list of alternative identifiers for the DID Document
	AlsoKnownAs []string
	// VerificationMethods is a list of verification methods to include in the DID Document
	// Cannot contain id #0 which is reserved for the identity key
	VerificationMethods []VerificationMethod
	// Services is a list of services to include in the DID Document
	Services []did.Service
}

type VerificationMethod struct {
	VerificationMethod did.VerificationMethod `json:"verificationMethod"`
	Purposes           []did.PublicKeyPurpose `json:"purposes"`
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

	// validate opts and build controller, aka, verification methods, key purposes, and services
	identityKeyID := id + "#0"

	var controller any
	if len(opts.Controller) != 0 {
		if len(opts.Controller) == 1 {
			// if there's only one controller, set it to the first controller
			controller = opts.Controller[0]
		} else {
			// if there's more than one controller, set it to the list of controllers
			controller = opts.Controller
		}
	}
	var aka any
	if len(opts.AlsoKnownAs) != 0 {
		if len(opts.AlsoKnownAs) == 1 {
			// if there's only one aka, set it to the first aka
			aka = opts.AlsoKnownAs[0]
		} else {
			// if there's more than one aka, set it to the list of akas
			aka = opts.AlsoKnownAs
		}
	}

	var vms []did.VerificationMethod
	var keyAgreement []did.VerificationMethodSet
	authentication := []did.VerificationMethodSet{identityKeyID}
	assertionMethod := []did.VerificationMethodSet{identityKeyID}
	capabilityInvocation := []did.VerificationMethodSet{identityKeyID}
	capabilityDelegation := []did.VerificationMethodSet{identityKeyID}
	if len(opts.VerificationMethods) > 0 {
		seenIDs := make(map[string]bool)
		for _, vm := range opts.VerificationMethods {
			if vm.VerificationMethod.ID == identityKeyID || vm.VerificationMethod.ID == "#0" || vm.VerificationMethod.ID == "0" {
				return nil, fmt.Errorf("verification method id 0 is reserved for the identity key")
			}
			if seenIDs[vm.VerificationMethod.ID] {
				return nil, fmt.Errorf("verification method id %s is not unique", vm.VerificationMethod.ID)
			}
			if vm.VerificationMethod.Type != JSONWebKeyType {
				return nil, fmt.Errorf("verification method type %s is not supported", vm.VerificationMethod.Type)
			}
			if vm.VerificationMethod.PublicKeyJWK == nil {
				return nil, fmt.Errorf("verification method public key jwk is required")
			}

			// mark as seen
			seenIDs[vm.VerificationMethod.ID] = true

			// update ID and controller in place, setting to thumbprint if none is provided

			// e.g. nothing -> did:dht:123456789abcdefghi#<jwk key id>
			if vm.VerificationMethod.ID == "" {
				vm.VerificationMethod.ID = id + "#" + vm.VerificationMethod.PublicKeyJWK.KID
			} else {
				// make sure the verification method ID and KID match
				vm.VerificationMethod.PublicKeyJWK.KID = vm.VerificationMethod.ID
			}

			// e.g. #key-1 -> did:dht:123456789abcdefghi#key-1
			if strings.HasPrefix(vm.VerificationMethod.ID, "#") {
				vm.VerificationMethod.ID = id + vm.VerificationMethod.ID
			}

			// e.g. key-1 -> did:dht:123456789abcdefghi#key-1
			if !strings.Contains(vm.VerificationMethod.ID, "#") {
				vm.VerificationMethod.ID = id + "#" + vm.VerificationMethod.ID
			}

			// if there's no controller, set it to the DID itself
			if vm.VerificationMethod.Controller == "" {
				vm.VerificationMethod.Controller = id
			}
			vms = append(vms, vm.VerificationMethod)

			// add purposes
			vmID := vm.VerificationMethod.ID
			for _, purpose := range vm.Purposes {
				switch purpose {
				case did.Authentication:
					authentication = append(authentication, vmID)
				case did.AssertionMethod:
					assertionMethod = append(assertionMethod, vmID)
				case did.KeyAgreement:
					keyAgreement = append(keyAgreement, vmID)
				case did.CapabilityInvocation:
					capabilityInvocation = append(capabilityInvocation, vmID)
				case did.CapabilityDelegation:
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
	kid := "0"
	key0JWK, err := jwx.PublicKeyToPublicKeyJWK(&kid, pubKey)
	if err != nil {
		return nil, err
	}
	vm0 := did.VerificationMethod{
		ID:           id + "#0",
		Type:         JSONWebKeyType,
		Controller:   id,
		PublicKeyJWK: key0JWK,
	}
	return &did.Document{
		ID:                   id,
		Controller:           controller,
		AlsoKnownAs:          aka,
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

// ToDNSPacket converts a DID DHT Document to a DNS packet with an optional list of types to include
func (d DHT) ToDNSPacket(doc did.Document, types []TypeIndex) (*dns.Msg, error) {
	var records []dns.RR
	var rootRecord []string
	keyLookup := make(map[string]string)

	// build controller and aka records
	if doc.Controller != nil {
		var controllerTxt string
		switch c := doc.Controller.(type) {
		case string:
			controllerTxt = c
		case []string:
			controllerTxt = strings.Join(c, ",")
		}
		controllerAnswer := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   "_cnt._did.",
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: []string{controllerTxt},
		}
		records = append(records, &controllerAnswer)
	}
	if doc.AlsoKnownAs != nil {
		var akaTxt string
		switch a := doc.AlsoKnownAs.(type) {
		case string:
			akaTxt = a
		case []string:
			akaTxt = strings.Join(a, ",")
		}
		akaAnswer := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   "_aka._did.",
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: []string{akaTxt},
		}
		records = append(records, &akaAnswer)
	}

	// build all key records
	var vmIDs []string
	for i, vm := range doc.VerificationMethod {
		recordIdentifier := fmt.Sprintf("k%d", i)
		keyLookup[vm.ID] = recordIdentifier

		keyType := keyTypeByAlg(crypto.SignatureAlgorithm(vm.PublicKeyJWK.ALG))
		if keyType < 0 {
			return nil, fmt.Errorf("unsupported key type given alg: %s", vm.PublicKeyJWK.ALG)
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

		vmKeyFragment := vm.ID[strings.LastIndex(vm.ID, "#")+1:]
		keyRecord := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("_%s._did.", recordIdentifier),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: []string{fmt.Sprintf("id=%s;t=%d;k=%s", vmKeyFragment, keyType, keyBase64Url)},
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

		svcTxt := fmt.Sprintf("id=%s;t=%s;se=%s", sID, service.Type, parseServiceData(service.ServiceEndpoint))
		if service.Sig != nil {
			svcTxt += fmt.Sprintf(";sig=%s", parseServiceData(service.Sig))
		}
		if service.Enc != nil {
			svcTxt += fmt.Sprintf(";enc=%s", parseServiceData(service.Enc))
		}
		serviceRecord := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("_%s._did.", recordIdentifier),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: []string{svcTxt},
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
		authIDs = append(authIDs, keyLookup[auth.(string)])
	}
	if len(authIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("auth=%s", strings.Join(authIDs, ",")))
	}

	var assertionIDs []string
	for _, assertion := range doc.AssertionMethod {
		assertionIDs = append(assertionIDs, keyLookup[assertion.(string)])
	}
	if len(assertionIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("asm=%s", strings.Join(assertionIDs, ",")))
	}

	var keyAgreementIDs []string
	for _, keyAgreement := range doc.KeyAgreement {
		keyAgreementIDs = append(keyAgreementIDs, keyLookup[keyAgreement.(string)])
	}
	if len(keyAgreementIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("agm=%s", strings.Join(keyAgreementIDs, ",")))
	}

	var capabilityInvocationIDs []string
	for _, capabilityInvocation := range doc.CapabilityInvocation {
		capabilityInvocationIDs = append(capabilityInvocationIDs, keyLookup[capabilityInvocation.(string)])
	}
	if len(capabilityInvocationIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("inv=%s", strings.Join(capabilityInvocationIDs, ",")))
	}

	var capabilityDelegationIDs []string
	for _, capabilityDelegation := range doc.CapabilityDelegation {
		capabilityDelegationIDs = append(capabilityDelegationIDs, keyLookup[capabilityDelegation.(string)])
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

	// add types record
	if len(types) != 0 {
		var typesStr []string
		for _, t := range types {
			typesStr = append(typesStr, strconv.Itoa(int(t)))
		}
		typesAnswer := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   "_typ._did.",
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: []string{"id=" + strings.Join(typesStr, ",")},
		}
		records = append(records, &typesAnswer)
	}

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

// make a best-effort to parse a service endpoints and other service data which we expect as either a single string
// value or an array of strings
func parseServiceData(serviceEndpoint any) string {
	switch se := serviceEndpoint.(type) {
	case string:
		return se
	case []string:
		if len(se) == 1 {
			return se[0]
		}
		return strings.Join(se, ",")
	case []any:
		if len(se) == 1 {
			return fmt.Sprintf("%v", se[0])
		}
		var result []string
		for _, v := range se {
			result = append(result, fmt.Sprintf("%v", v))
		}
		return strings.Join(result, ",")
	}
	return ""
}

// FromDNSPacket converts a DNS packet to a DID DHT Document
func (d DHT) FromDNSPacket(msg *dns.Msg) (*did.Document, []TypeIndex, error) {
	doc := did.Document{
		ID: d.String(),
	}

	var types []TypeIndex
	keyLookup := make(map[string]string)
	for _, rr := range msg.Answer {
		switch record := rr.(type) {
		case *dns.TXT:
			if strings.HasPrefix(record.Hdr.Name, "_cnt") {
				doc.Controller = strings.Split(record.Txt[0], ",")
			}
			if strings.HasPrefix(record.Hdr.Name, "_aka") {
				doc.AlsoKnownAs = strings.Split(record.Txt[0], ",")
			}
			if strings.HasPrefix(record.Hdr.Name, "_k") {
				data := parseTxtData(strings.Join(record.Txt, ","))
				vmID := data["id"]
				keyType := keyTypeLookUp(data["t"])
				keyBase64URL := data["k"]

				// Convert keyBase64URL back to PublicKeyJWK
				pubKeyBytes, err := base64.RawURLEncoding.DecodeString(keyBase64URL)
				if err != nil {
					return nil, nil, err
				}
				pubKey, err := crypto.BytesToPubKey(pubKeyBytes, keyType)
				if err != nil {
					return nil, nil, err
				}
				pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK(&vmID, pubKey)
				if err != nil {
					return nil, nil, err
				}

				vm := did.VerificationMethod{
					ID:           d.String() + "#" + vmID,
					Type:         JSONWebKeyType,
					Controller:   d.String(),
					PublicKeyJWK: pubKeyJWK,
				}
				doc.VerificationMethod = append(doc.VerificationMethod, vm)

				// add to key lookup (e.g.  "k1" -> "key1")
				keyLookup[strings.Split(record.Hdr.Name, ".")[0][1:]] = vmID
			} else if strings.HasPrefix(record.Hdr.Name, "_s") {
				data := parseTxtData(strings.Join(record.Txt, ","))
				sID := data["id"]
				serviceType := data["t"]
				serviceEndpoint := data["se"]
				var serviceEndpointValue any
				if strings.Contains(serviceEndpoint, ",") {
					serviceEndpointValue = strings.Split(serviceEndpoint, ",")
				} else {
					serviceEndpointValue = serviceEndpoint
				}
				service := did.Service{
					ID:              d.String() + "#" + sID,
					Type:            serviceType,
					ServiceEndpoint: serviceEndpointValue,
				}
				if data["sig"] != "" {
					if strings.Contains(data["sig"], ",") {
						service.Sig = strings.Split(data["sig"], ",")
					} else {
						service.Sig = data["sig"]
					}
				}
				if data["enc"] != "" {
					if strings.Contains(data["enc"], ",") {
						service.Enc = strings.Split(data["enc"], ",")
					} else {
						service.Enc = data["enc"]
					}
				}
				doc.Services = append(doc.Services, service)

			} else if record.Hdr.Name == "_typ._did." {
				if record.Txt[0] == "" || len(record.Txt) != 1 {
					return nil, nil, fmt.Errorf("invalid types record")
				}
				typesStr := strings.Split(strings.TrimPrefix(record.Txt[0], "id="), ",")
				for _, t := range typesStr {
					tInt, err := strconv.Atoi(t)
					if err != nil {
						return nil, nil, err
					}
					types = append(types, TypeIndex(tInt))
				}
			} else if record.Hdr.Name == "_did." {
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
					case "auth":
						for _, valueItem := range valueItems {
							doc.Authentication = append(doc.Authentication, doc.ID+"#"+keyLookup[valueItem])
						}
					case "asm":
						for _, valueItem := range valueItems {
							doc.AssertionMethod = append(doc.AssertionMethod, doc.ID+"#"+keyLookup[valueItem])
						}
					case "agm":
						for _, valueItem := range valueItems {
							doc.KeyAgreement = append(doc.KeyAgreement, doc.ID+"#"+keyLookup[valueItem])
						}
					case "inv":
						for _, valueItem := range valueItems {
							doc.CapabilityInvocation = append(doc.CapabilityInvocation, doc.ID+"#"+keyLookup[valueItem])
						}
					case "del":
						for _, valueItem := range valueItems {
							doc.CapabilityDelegation = append(doc.CapabilityDelegation, doc.ID+"#"+keyLookup[valueItem])
						}
					}
				}
			}
		}
	}

	return &doc, types, nil
}

func parseTxtData(data string) map[string]string {
	pairs := strings.Split(data, ";")
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
	case "2":
		return crypto.P256
	default:
		return ""
	}
}

func keyTypeByAlg(alg crypto.SignatureAlgorithm) int {
	switch alg {
	case crypto.EdDSA:
		return 0
	case crypto.ES256K:
		return 1
	case crypto.ES256:
		return 2
	default:
		return -1
	}
}
