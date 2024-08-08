package did

import (
	"fmt"
	"strings"

	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/cryptosuite"
	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/fxamacker/cbor/v2"
)

const (
	// CBOR map keys
	keyID                   byte = 1
	keyVerificationMethod   byte = 2
	keyAuthentication       byte = 3
	keyAssertionMethod      byte = 4
	keyKeyAgreement         byte = 5
	keyCapabilityInvocation byte = 6
	keyCapabilityDelegation byte = 7
	keyService              byte = 8
	keyController           byte = 9
	keyAlsoKnownAs          byte = 10
	keyTypes                byte = 11
	keyGateways             byte = 12
	keyPreviousDID          byte = 13

	// VerificationMethod keys
	keyVMID         byte = 1
	keyVMType       byte = 2
	keyVMController byte = 3
	keyVMPublicKey  byte = 4

	// PublicKeyJWK keys
	keyJWKKid byte = 1
	keyJWKAlg byte = 2
	keyJWKCrv byte = 3
	keyJWKKty byte = 4
	keyJWKX   byte = 5
	keyJWKY   byte = 6

	// Service keys
	keyServiceID       byte = 1
	keyServiceType     byte = 2
	keyServiceEndpoint byte = 3
)

// ToCBOR converts a DID DHT Document to a CBOR byte array
// ToCBOR converts a DID DHT Document to a CBOR byte array
func (d DHT) ToCBOR(doc did.Document, types []TypeIndex, gateways []AuthoritativeGateway, previousDID *PreviousDID) ([]byte, error) {
	em, err := cbor.EncOptions{Sort: cbor.SortCanonical}.EncMode()
	if err != nil {
		return nil, err
	}

	cborMap := make(map[byte]any)

	// Extract the DID suffix
	didSuffix := strings.TrimPrefix(doc.ID, "did:dht:")

	// Only include non-empty fields
	if didSuffix != "" {
		cborMap[keyID] = didSuffix
	}

	if len(doc.VerificationMethod) > 0 {
		vms := make([]any, len(doc.VerificationMethod))
		for i, vm := range doc.VerificationMethod {
			vmMap := make(map[byte]any)
			vmMap[keyVMID] = strings.TrimPrefix(vm.ID, doc.ID+"#")
			vmMap[keyVMType] = vm.Type
			if vm.Controller != doc.ID {
				vmMap[keyVMController] = vm.Controller
			}
			if vm.PublicKeyJWK != nil {
				jwkMap := make(map[byte]any)
				jwkMap[keyJWKKid] = vm.PublicKeyJWK.KID
				jwkMap[keyJWKAlg] = vm.PublicKeyJWK.ALG
				jwkMap[keyJWKCrv] = vm.PublicKeyJWK.CRV
				jwkMap[keyJWKKty] = vm.PublicKeyJWK.KTY
				jwkMap[keyJWKX] = vm.PublicKeyJWK.X
				if vm.PublicKeyJWK.Y != "" {
					jwkMap[keyJWKY] = vm.PublicKeyJWK.Y
				}
				vmMap[keyVMPublicKey] = jwkMap
			}
			vms[i] = vmMap
		}
		cborMap[keyVerificationMethod] = vms
	}

	addVerificationRelationship := func(key byte, relationships []did.VerificationMethodSet) {
		if len(relationships) > 0 {
			refs := make([]string, len(relationships))
			for i, r := range relationships {
				refs[i] = strings.TrimPrefix(r.(string), doc.ID+"#")
			}
			cborMap[key] = refs
		}
	}

	addVerificationRelationship(keyAuthentication, doc.Authentication)
	addVerificationRelationship(keyAssertionMethod, doc.AssertionMethod)
	addVerificationRelationship(keyKeyAgreement, doc.KeyAgreement)
	addVerificationRelationship(keyCapabilityInvocation, doc.CapabilityInvocation)
	addVerificationRelationship(keyCapabilityDelegation, doc.CapabilityDelegation)

	if len(doc.Services) > 0 {
		services := make([]any, len(doc.Services))
		for i, svc := range doc.Services {
			svcMap := make(map[byte]any)
			svcMap[keyServiceID] = strings.TrimPrefix(svc.ID, doc.ID+"#")
			svcMap[keyServiceType] = svc.Type
			svcMap[keyServiceEndpoint] = svc.ServiceEndpoint
			services[i] = svcMap
		}
		cborMap[keyService] = services
	}

	if doc.Controller != nil {
		cborMap[keyController] = doc.Controller
	}

	if doc.AlsoKnownAs != nil {
		cborMap[keyAlsoKnownAs] = doc.AlsoKnownAs
	}

	if len(types) > 0 {
		cborMap[keyTypes] = types
	}

	if len(gateways) > 0 {
		cborMap[keyGateways] = gateways
	}

	if previousDID != nil {
		cborMap[keyPreviousDID] = map[string]any{
			"did":       previousDID.PreviousDID,
			"signature": previousDID.Signature,
		}
	}

	return em.Marshal(cborMap)
}

// FromCBOR converts a CBOR byte array to a DID DHT Document
func (d DHT) FromCBOR(cborData []byte) (*DIDDHTDocument, error) {
	var cborMap map[byte]any
	err := cbor.Unmarshal(cborData, &cborMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal CBOR data: %w", err)
	}

	var doc did.Document
	var types []TypeIndex
	var gateways []AuthoritativeGateway
	var previousDID *PreviousDID

	if id, ok := cborMap[keyID].(string); ok {
		doc.ID = "did:dht:" + id
	}

	if vms, ok := cborMap[keyVerificationMethod].([]any); ok {
		for _, vmInterface := range vms {
			if vm, ok := vmInterface.(map[any]any); ok {
				verificationMethod := did.VerificationMethod{
					ID:         doc.ID + "#" + getMapValue(vm, keyVMID).(string),
					Type:       cryptosuite.LDKeyType(getMapValue(vm, keyVMType).(string)),
					Controller: doc.ID, // Default to the DID itself
				}
				if controller, ok := getMapValue(vm, keyVMController).(string); ok {
					verificationMethod.Controller = controller
				}
				if jwkInterface := getMapValue(vm, keyVMPublicKey); jwkInterface != nil {
					if jwk, ok := jwkInterface.(map[any]any); ok {
						verificationMethod.PublicKeyJWK = &jwx.PublicKeyJWK{
							KID: getMapValue(jwk, keyJWKKid).(string),
							ALG: getMapValue(jwk, keyJWKAlg).(string),
							CRV: getMapValue(jwk, keyJWKCrv).(string),
							KTY: getMapValue(jwk, keyJWKKty).(string),
							X:   getMapValue(jwk, keyJWKX).(string),
						}
						if y, ok := getMapValue(jwk, keyJWKY).(string); ok {
							verificationMethod.PublicKeyJWK.Y = y
						}
					}
				}
				doc.VerificationMethod = append(doc.VerificationMethod, verificationMethod)
			}
		}
	}

	getVerificationRelationship := func(key byte) []did.VerificationMethodSet {
		if relationships, ok := cborMap[key].([]any); ok {
			var result []did.VerificationMethodSet
			for _, r := range relationships {
				if ref, ok := r.(string); ok {
					result = append(result, doc.ID+"#"+ref)
				}
			}
			return result
		}
		return nil
	}

	doc.Authentication = getVerificationRelationship(keyAuthentication)
	doc.AssertionMethod = getVerificationRelationship(keyAssertionMethod)
	doc.KeyAgreement = getVerificationRelationship(keyKeyAgreement)
	doc.CapabilityInvocation = getVerificationRelationship(keyCapabilityInvocation)
	doc.CapabilityDelegation = getVerificationRelationship(keyCapabilityDelegation)

	if services, ok := cborMap[keyService].([]any); ok {
		for _, svcInterface := range services {
			if svc, ok := svcInterface.(map[any]any); ok {
				service := did.Service{
					ID:   doc.ID + "#" + getMapValue(svc, keyServiceID).(string),
					Type: getMapValue(svc, keyServiceType).(string),
				}
				if endpoints, ok := getMapValue(svc, keyServiceEndpoint).([]any); ok {
					var serviceEndpoint []string
					for _, e := range endpoints {
						serviceEndpoint = append(serviceEndpoint, e.(string))
					}
					service.ServiceEndpoint = serviceEndpoint
				} else if endpoint, ok := getMapValue(svc, keyServiceEndpoint).(string); ok {
					service.ServiceEndpoint = endpoint
				}
				doc.Services = append(doc.Services, service)
			}
		}
	}

	if controller, ok := cborMap[keyController]; ok {
		doc.Controller = controller
	}

	if alsoKnownAs, ok := cborMap[keyAlsoKnownAs].([]any); ok {
		doc.AlsoKnownAs = make([]string, len(alsoKnownAs))
		var akas []string
		for _, aka := range alsoKnownAs {
			akas = append(akas, aka.(string))
		}
		doc.AlsoKnownAs = akas
	}

	if typesInterface, ok := cborMap[keyTypes].([]any); ok {
		for _, t := range typesInterface {
			if typeInt, ok := t.(int64); ok {
				types = append(types, TypeIndex(typeInt))
			}
		}
	}

	if gatewaysInterface, ok := cborMap[keyGateways].([]any); ok {
		for _, g := range gatewaysInterface {
			if gatewayString, ok := g.(string); ok {
				gateways = append(gateways, AuthoritativeGateway(gatewayString))
			}
		}
	}

	if prev, ok := cborMap[keyPreviousDID].(map[string]any); ok {
		previousDID = &PreviousDID{
			PreviousDID: DHT(prev["did"].(string)),
			Signature:   prev["signature"].(string),
		}
	}

	return &DIDDHTDocument{
		Doc:         doc,
		Types:       types,
		Gateways:    gateways,
		PreviousDID: previousDID,
	}, nil
}

func getMapValue(m map[any]any, key byte) any {
	for k, v := range m {
		switch typedKey := k.(type) {
		case uint64:
			if byte(typedKey) == key {
				return v
			}
		case byte:
			if typedKey == key {
				return v
			}
		}
	}
	return nil
}
