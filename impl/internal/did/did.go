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
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/tv42/zbase32"
)

type (
	DHT                  string
	TypeIndex            int
	AuthoritativeGateway string
)

type PreviousDID struct {
	PreviousDID DHT    `json:"did"`
	Signature   string `json:"signature"`
}

const (
	// Prefix did:dht prefix
	Prefix               = "did:dht"
	DHTMethod did.Method = "dht"

	// Version corresponds to the version fo the specification https://did-dht.com/#dids-as-dns-records
	Version int = 0

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

// IdentityKey returns the ed25519 public key for the DHT identifier https://did-dht.com/#identity-key
func (d DHT) IdentityKey() (ed25519.PublicKey, error) {
	suffix, err := d.Suffix()
	if err != nil {
		return nil, err
	}
	pk, err := zbase32.DecodeString(suffix)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func (DHT) Method() did.Method {
	return DHTMethod
}

// CreateDIDDHTOpts is a set of options for creating a did:dht identifier
// Note: this does not include additional properties only present in the DNS representation (e.g. gateways, types)
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
			if vm.VerificationMethod.Type != cryptosuite.JSONWebKeyType {
				return nil, fmt.Errorf("verification method type %s is not supported", vm.VerificationMethod.Type)
			}
			if vm.VerificationMethod.PublicKeyJWK == nil {
				return nil, fmt.Errorf("verification method public key jwk is required")
			}

			// mark as seen
			seenIDs[vm.VerificationMethod.ID] = true

			// if verification method ID is set, make sure it's fully qualified
			if vm.VerificationMethod.ID != "" {
				if strings.HasPrefix(vm.VerificationMethod.ID, "#") {
					vm.VerificationMethod.ID = id + vm.VerificationMethod.ID
				} else if !strings.Contains(vm.VerificationMethod.ID, "#") {
					vm.VerificationMethod.ID = id + "#" + vm.VerificationMethod.ID
				}
			} else {
				// if no verification method ID is set, set it to the JWK thumbprint
				thumbprint, err := vm.VerificationMethod.PublicKeyJWK.Thumbprint()
				if err != nil {
					return nil, fmt.Errorf("failed to calculate JWK thumbprint: %v", err)
				}
				vm.VerificationMethod.ID = id + "#" + thumbprint
				vm.VerificationMethod.PublicKeyJWK.KID = thumbprint
			}

			// make sure the JWK KID matches the unqualified VM ID
			vm.VerificationMethod.PublicKeyJWK.KID = strings.TrimPrefix(vm.VerificationMethod.ID, id+"#")

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
	key0JWK.ALG = string(crypto.EdDSA)
	if err != nil {
		return nil, err
	}
	vm0 := did.VerificationMethod{
		ID:           id + "#0",
		Type:         cryptosuite.JSONWebKeyType,
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
func (d DHT) ToDNSPacket(doc did.Document, types []TypeIndex, gateways []AuthoritativeGateway, previousDID *PreviousDID) (*dns.Msg, error) {
	var records []dns.RR
	var rootRecord []string
	keyLookup := make(map[string]string)

	suffix, err := d.Suffix()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get suffix while decoding DNS packet")
	}

	// handle the previous DID if it's present
	if previousDID != nil {
		// make sure it's valid
		if err = ValidatePreviousDIDSignatureValid(d, *previousDID); err != nil {
			return nil, err
		}
		// add it to the record set
		records = append(records, &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   "_prv._did.",
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: chunkTextRecord(fmt.Sprintf("id=%s;s=%s", previousDID.PreviousDID, previousDID.Signature)),
		})
	}

	// first append the version to the root record
	rootRecord = append(rootRecord, fmt.Sprintf("v=%d", Version))

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
			Txt: chunkTextRecord(controllerTxt),
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
			Txt: chunkTextRecord(akaTxt),
		}
		records = append(records, &akaAnswer)
	}

	// add all gateways
	for _, gateway := range gateways {
		gatewayAnswer := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("_did.%s.", suffix),
				Rrtype: dns.TypeNS,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: chunkTextRecord(string(gateway)),
		}
		records = append(records, &gatewayAnswer)
	}

	// build all key records
	var vmIDs []string
	for i, vm := range doc.VerificationMethod {
		recordIdentifier := fmt.Sprintf("k%d", i)
		keyLookup[vm.ID] = recordIdentifier

		keyType := keyTypeForJWK(*vm.PublicKeyJWK)
		if keyType < 0 {
			return nil, fmt.Errorf("unsupported key type given alg: %s", vm.PublicKeyJWK.ALG)
		}

		// convert the public key to a base64url encoded string
		pubKey, err := vm.PublicKeyJWK.ToPublicKey()
		if err != nil {
			return nil, err
		}

		// as per the spec's guidance DNS representations use compressed keys, so we must marshal them as such
		pubKeyBytes, err := crypto.PubKeyToBytes(pubKey, crypto.ECDSAMarshalCompressed)
		if err != nil {
			return nil, err
		}

		txtRecord := ""

		// calculate the JWK thumbprint
		thumbprint, err := vm.PublicKeyJWK.Thumbprint()
		if err != nil {
			return nil, fmt.Errorf("failed to calculate JWK thumbprint: %v", err)
		}

		// only include the id if it's not the JWK thumbprint
		unqualifiedVMID := strings.TrimPrefix(vm.ID, doc.ID+"#")
		if unqualifiedVMID != thumbprint {
			txtRecord += fmt.Sprintf("id=%s;", unqualifiedVMID)
		}
		txtRecord += fmt.Sprintf("t=%d;k=%s", keyType, base64.RawURLEncoding.EncodeToString(pubKeyBytes))

		// only include the alg if it's not the default alg for the key type
		forKeyType := algIsDefaultForJWK(*vm.PublicKeyJWK)
		if !forKeyType {
			txtRecord += fmt.Sprintf(";a=%s", vm.PublicKeyJWK.ALG)
		}

		// note the controller if it differs from the DID
		if vm.Controller != doc.ID {
			// handle the case where the controller of the identity key is not the DID itself
			if vm.ID == doc.ID+"#0" && (vm.Controller != "" || vm.Controller != doc.ID) {
				return nil, fmt.Errorf("controller of identity key must be the DID itself, instead it is: %s", vm.Controller)
			}
			txtRecord += fmt.Sprintf(";c=%s", vm.Controller)
		}

		if len(txtRecord) > 255 {
			return nil, fmt.Errorf("key value exceeds 255 characters")
		}
		keyRecord := dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("_%s._did.", recordIdentifier),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    7200,
			},
			Txt: chunkTextRecord(txtRecord),
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
			Txt: chunkTextRecord(svcTxt),
		}

		records = append(records, &serviceRecord)
		svcIDs = append(svcIDs, recordIdentifier)
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

	// add services to the root record
	if len(svcIDs) != 0 {
		rootRecord = append(rootRecord, fmt.Sprintf("svc=%s", strings.Join(svcIDs, ",")))
	}

	// add the root record
	rootAnswer := dns.TXT{
		Hdr: dns.RR_Header{
			Name:   fmt.Sprintf("_did.%s.", suffix),
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
			Txt: chunkTextRecord("id=" + strings.Join(typesStr, ",")),
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

// DIDDHTDocument is a DID DHT Document along with additional metadata the DID supports
type DIDDHTDocument struct {
	Doc         did.Document           `json:"did,omitempty"`
	Types       []TypeIndex            `json:"types,omitempty"`
	Gateways    []AuthoritativeGateway `json:"gateways,omitempty"`
	PreviousDID *PreviousDID           `json:"previousDid,omitempty"`
}

// FromDNSPacket converts a DNS packet to a DID DHT Document
// Returns the DID Document, a list of types, a list of authoritative gateways, and an error
func (d DHT) FromDNSPacket(msg *dns.Msg) (*DIDDHTDocument, error) {
	didID := d.String()
	doc := did.Document{
		ID: didID,
	}

	identityKey, err := DHT(didID).IdentityKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get identity key while decoding DNS packet")
	}

	suffix, err := d.Suffix()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get suffix while decoding DNS packet")
	}

	// track the authoritative gateways
	var gateways []AuthoritativeGateway
	// track the types
	var types []TypeIndex
	// track the previous DID
	var previousDID *PreviousDID
	keyLookup := make(map[string]string)
	for _, rr := range msg.Answer {
		switch record := rr.(type) {
		case *dns.TXT:
			if strings.HasPrefix(record.Hdr.Name, "_cnt") {
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				controllers := strings.Split(unchunkedTextRecord, ",")
				if len(controllers) == 1 {
					doc.Controller = controllers[0]
				} else {
					doc.Controller = controllers
				}
			}
			if strings.HasPrefix(record.Hdr.Name, "_aka") {
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				doc.AlsoKnownAs = strings.Split(unchunkedTextRecord, ",")
			}
			if strings.HasPrefix(record.Hdr.Name, "_k") {
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				data := parseTxtData(unchunkedTextRecord)
				vmID := data["id"]
				keyType := keyTypeLookUp(data["t"])
				keyBase64URL := data["k"]
				controller := data["c"]
				alg := data["a"]

				// set the controller to the DID if it's not provided
				if controller == "" {
					controller = didID
				}

				// Convert keyBase64URL back to PublicKeyJWK
				pubKeyBytes, err := base64.RawURLEncoding.DecodeString(keyBase64URL)
				if err != nil {
					return nil, err
				}

				// as per the spec's guidance DNS representations use compressed keys, so we must unmarshall them as such
				pubKey, err := crypto.BytesToPubKey(pubKeyBytes, keyType, crypto.ECDSAUnmarshalCompressed)
				if err != nil {
					return nil, err
				}

				pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK(nil, pubKey)
				if err != nil {
					return nil, err
				}

				// set the algorithm if it's not the default for the key type
				if alg == "" {
					defaultAlg := defaultAlgForJWK(*pubKeyJWK)
					if defaultAlg == "" {
						return nil, fmt.Errorf("unable to provide default alg for unsupported key type: %s", keyType)
					}
					pubKeyJWK.ALG = defaultAlg
				} else {
					pubKeyJWK.ALG = alg
				}

				// compare pubkey to identity key to see if they're equal, and if they are set the vmID and kid to 0
				if identityKey.Equal(pubKey) {
					vmID = "0"
					pubKeyJWK.KID = "0"
				}

				// if the verification method ID is not set, set it to the thumbprint
				if vmID == "" {
					thumbprint, err := pubKeyJWK.Thumbprint()
					if err != nil {
						return nil, fmt.Errorf("failed to calculate JWK thumbprint: %v", err)
					}
					vmID = thumbprint
					pubKeyJWK.KID = thumbprint
				} else {
					pubKeyJWK.KID = vmID
				}

				vm := did.VerificationMethod{
					ID:           didID + "#" + vmID,
					Type:         cryptosuite.JSONWebKeyType,
					Controller:   controller,
					PublicKeyJWK: pubKeyJWK,
				}
				doc.VerificationMethod = append(doc.VerificationMethod, vm)

				// add to key lookup (e.g.  "k1" -> "key1")
				keyLookup[strings.Split(record.Hdr.Name, ".")[0][1:]] = vmID
			} else if strings.HasPrefix(record.Hdr.Name, "_s") {
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				data := parseTxtData(unchunkedTextRecord)
				sID := data["id"]
				serviceType := data["t"]
				serviceEndpoint := data["se"]
				service := did.Service{
					ID:              didID + "#" + sID,
					Type:            serviceType,
					ServiceEndpoint: strings.Split(serviceEndpoint, ","),
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
				if record.Txt[0] == "" {
					return nil, fmt.Errorf("types record is empty")
				}
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				typesStr := strings.Split(strings.TrimPrefix(unchunkedTextRecord, "id="), ",")
				for _, t := range typesStr {
					tInt, err := strconv.Atoi(t)
					if err != nil {
						return nil, err
					}
					types = append(types, TypeIndex(tInt))
				}
			} else if record.Hdr.Name == fmt.Sprintf("_did.%s.", suffix) && record.Hdr.Rrtype == dns.TypeNS {
				if record.Txt[0] == "" {
					return nil, fmt.Errorf("gateway record is empty")
				}
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				gateways = append(gateways, AuthoritativeGateway(unchunkedTextRecord))
			} else if record.Hdr.Name == "_prv._did." && record.Hdr.Rrtype == dns.TypeTXT {
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				data := parseTxtData(unchunkedTextRecord)
				previousDID = &PreviousDID{
					PreviousDID: DHT(data["id"]),
					Signature:   data["s"],
				}
				// validate previous DID signature
				if err = ValidatePreviousDIDSignatureValid(d, *previousDID); err != nil {
					return nil, err
				}
			} else if record.Hdr.Name == fmt.Sprintf("_did.%s.", suffix) && record.Hdr.Rrtype == dns.TypeTXT {
				unchunkedTextRecord := unchunkTextRecord(record.Txt)
				rootItems := strings.Split(unchunkedTextRecord, ";")

				seenVersion := false
				for _, item := range rootItems {
					kv := strings.Split(item, "=")
					if len(kv) != 2 {
						continue
					}

					key, values := kv[0], kv[1]
					valueItems := strings.Split(values, ",")

					switch key {
					case "v":
						if len(valueItems) != 1 || valueItems[0] != strconv.Itoa(Version) {
							return nil, fmt.Errorf("invalid version: %s", values)
						}
						seenVersion = true
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
				if !seenVersion {
					return nil, fmt.Errorf("root record missing version identifier")
				}
			}
		}
	}

	return &DIDDHTDocument{
		Doc:         doc,
		Types:       types,
		Gateways:    gateways,
		PreviousDID: previousDID,
	}, nil
}

// CreatePreviousDIDRecord creates a PreviousDID record for the given previous DID and current DID
func CreatePreviousDIDRecord(previousDIDPrivateKey ed25519.PrivateKey, previousDID, currentDID DHT) (*PreviousDID, error) {
	currentDIDIdentityKey, err := currentDID.IdentityKey()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get identity key from currentDID: %s", currentDID)
	}
	previousDIDSignature := ed25519.Sign(previousDIDPrivateKey, currentDIDIdentityKey)
	return &PreviousDID{
		PreviousDID: previousDID,
		Signature:   base64.RawURLEncoding.EncodeToString(previousDIDSignature),
	}, nil
}

// ValidatePreviousDIDSignatureValid validates the signature of the previous DID over the current DID
func ValidatePreviousDIDSignatureValid(currentDID DHT, previousDID PreviousDID) error {
	identityKey, err := currentDID.IdentityKey()
	if err != nil {
		return errors.Wrapf(err, "failed to get identity key from the current DID: %s", currentDID)
	}
	previousDIDKey, err := previousDID.PreviousDID.IdentityKey()
	if err != nil {
		return errors.Wrapf(err, "failed to get identity key from the previous DID: %s", previousDID.PreviousDID)
	}
	decodedSignature, err := base64.RawURLEncoding.DecodeString(previousDID.Signature)
	if err != nil {
		return errors.Wrap(err, "failed to decode the previous DID's signature")
	}
	if ok := ed25519.Verify(previousDIDKey, identityKey, decodedSignature); !ok {
		return errors.New("the previous DID signature is invalid")
	}
	return nil
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

// algIsDefaultForJWK returns true if the given JWK ALG is the default for the given key type
// according to the key type index https://did-dht.com/registry/#key-type-index
func algIsDefaultForJWK(jwk jwx.PublicKeyJWK) bool {
	// Ed25519 : EdDSA
	if jwk.CRV == crypto.Ed25519.String() && jwk.KTY == jwa.OKP.String() {
		return jwk.ALG == string(crypto.EdDSA)
	}
	// secp256k1 : ES256K
	if jwk.CRV == crypto.SECP256k1.String() && jwk.KTY == jwa.EC.String() {
		return jwk.ALG == string(crypto.ES256K)
	}
	// P-256 : ES256
	if jwk.CRV == crypto.P256.String() && jwk.KTY == jwa.EC.String() {
		return jwk.ALG == string(crypto.ES256)
	}
	// X25519 : ECDH-ES+A256KW
	if jwk.CRV == crypto.X25519.String() && jwk.KTY == jwa.OKP.String() {
		return jwk.ALG == string(crypto.ECDHESA256KW)
	}
	return false
}

// defaultAlgForJWK returns the default signature algorithm for the given JWK based on the key type index
// https://did-dht.com/registry/#key-type-index
func defaultAlgForJWK(jwk jwx.PublicKeyJWK) string {
	// Ed25519 : EdDSA
	if jwk.CRV == crypto.Ed25519.String() && jwk.KTY == jwa.OKP.String() {
		return string(crypto.EdDSA)
	}
	// secp256k1 : ES256K
	if jwk.CRV == crypto.SECP256k1.String() && jwk.KTY == jwa.EC.String() {
		return string(crypto.ES256K)
	}
	// P-256 : ES256
	if jwk.CRV == crypto.P256.String() && jwk.KTY == jwa.EC.String() {
		return string(crypto.ES256)
	}
	// X25519 : ECDH-ES+A256KW
	if jwk.CRV == crypto.X25519.String() && jwk.KTY == jwa.OKP.String() {
		return string(crypto.ECDHESA256KW)
	}
	return ""
}

// keyTypeLookUp returns the key type for the given key type index
// https://did-dht.com/registry/#key-type-index
func keyTypeLookUp(keyType string) crypto.KeyType {
	switch keyType {
	case "0":
		return crypto.Ed25519
	case "1":
		return crypto.SECP256k1
	case "2":
		return crypto.P256
	case "3":
		return crypto.X25519
	default:
		return ""
	}
}

// keyTypeForJWK returns the key type index for the given JWK according to the key type index
// https://did-dht.com/registry/#key-type-index
func keyTypeForJWK(jwk jwx.PublicKeyJWK) int {
	// Ed25519 : EdDSA : 0
	if jwk.CRV == crypto.Ed25519.String() && jwk.KTY == jwa.OKP.String() {
		return 0
	}
	// secp256k1 : ES256K : 1
	if jwk.CRV == crypto.SECP256k1.String() && jwk.KTY == jwa.EC.String() {
		return 1
	}
	// P-256 : ES256 : 2
	if jwk.CRV == crypto.P256.String() && jwk.KTY == jwa.EC.String() {
		return 2
	}
	// X25519 : ECDH-ES+A256KW : 3
	if jwk.CRV == crypto.X25519.String() && jwk.KTY == jwa.OKP.String() {
		return 3
	}
	return -1
}

// chunkTextRecord splits a text record into chunks of 255 characters, taking into account multi-byte characters
func chunkTextRecord(record string) []string {
	var chunks []string
	runeArray := []rune(record) // Convert to rune slice to properly handle multi-byte characters
	for len(runeArray) > 0 {
		if len(runeArray) <= 255 {
			chunks = append(chunks, string(runeArray))
			break
		}

		chunks = append(chunks, string(runeArray[:255]))
		runeArray = runeArray[255:]
	}
	return chunks
}

// unchunkTextRecord joins chunks of a text record
func unchunkTextRecord(chunks []string) string {
	return strings.Join(chunks, "")
}
