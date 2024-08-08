package did

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/TBD54566975/ssi-sdk/crypto"
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

	// PublicKey keys
	keyPKType byte = 1
	keyPKData byte = 2
	keyPKAlg  byte = 3
	keyPKCrv  byte = 4
	keyPKKty  byte = 5

	// Service keys
	keyServiceID       byte = 1
	keyServiceType     byte = 2
	keyServiceEndpoint byte = 3
	keyServiceEnc      byte = 4
	keyServiceSig      byte = 5
)

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

			// Convert the public key to bytes
			pubKey, err := vm.PublicKeyJWK.ToPublicKey()
			if err != nil {
				return nil, err
			}
			pubKeyBytes, err := crypto.PubKeyToBytes(pubKey, crypto.ECDSAMarshalCompressed)
			if err != nil {
				return nil, err
			}

			keyType := keyTypeForJWK(*vm.PublicKeyJWK)
			pkData := fmt.Sprintf("t=%d;k=%s", keyType, base64.RawURLEncoding.EncodeToString(pubKeyBytes))

			// Only include the alg if it's not the default for the key type
			if !algIsDefaultForJWK(*vm.PublicKeyJWK) {
				pkData += fmt.Sprintf(";a=%s", vm.PublicKeyJWK.ALG)
			}

			// Only include the controller if it's different from the DID
			if vm.Controller != doc.ID {
				pkData += fmt.Sprintf(";c=%s", vm.Controller)
			}

			vmMap[keyVMPublicKey] = pkData
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
			svcMap[keyServiceID] = svc.ID
			svcMap[keyServiceType] = svc.Type
			svcMap[keyServiceEndpoint] = svc.ServiceEndpoint
			if svc.Enc != nil {
				svcMap[keyServiceEnc] = svc.Enc
			}
			if svc.Sig != nil {
				svcMap[keyServiceSig] = svc.Sig
			}
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

func (d DHT) FromCBOR(cborData []byte) (*DIDDHTDocument, error) {
	var cborMap map[any]any
	if err := cbor.Unmarshal(cborData, &cborMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CBOR data: %w", err)
	}

	var doc did.Document
	var types []TypeIndex
	var gateways []AuthoritativeGateway
	var previousDID *PreviousDID

	if id, ok := getMapValue(cborMap, keyID).(string); ok {
		doc.ID = "did:dht:" + id
	}

	// Get the identity key from the DID
	identityKey, err := DHT(doc.ID).IdentityKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity key: %w", err)
	}

	if vms, ok := getMapValue(cborMap, keyVerificationMethod).([]any); ok {
		for _, vmInterface := range vms {
			if vm, ok := vmInterface.(map[any]any); ok {
				vmID := doc.ID + "#" + getMapValue(vm, keyVMID).(string)
				pkData := getMapValue(vm, keyVMPublicKey).(string)

				pkParts := strings.Split(pkData, ";")
				var keyType int
				var keyData, alg, controller string

				for _, part := range pkParts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) != 2 {
						continue
					}
					switch kv[0] {
					case "t":
						fmt.Sscanf(kv[1], "%d", &keyType)
					case "k":
						keyData = kv[1]
					case "a":
						alg = kv[1]
					case "c":
						controller = kv[1]
					}
				}

				if controller == "" {
					controller = doc.ID
				}

				keyBytes, err := base64.RawURLEncoding.DecodeString(keyData)
				if err != nil {
					return nil, err
				}

				pubKey, err := crypto.BytesToPubKey(keyBytes, keyTypeLookUp(fmt.Sprintf("%d", keyType)), crypto.ECDSAUnmarshalCompressed)
				if err != nil {
					return nil, err
				}

				jwk, err := jwx.PublicKeyToPublicKeyJWK(nil, pubKey)
				if err != nil {
					return nil, err
				}

				if alg == "" {
					alg = defaultAlgForJWK(*jwk)
				}
				jwk.ALG = alg

				// Check if this is the identity key
				if bytes.Equal(keyBytes, identityKey) {
					jwk.KID = "0"
					vmID = doc.ID + "#0"
				} else {
					jwk.KID = strings.TrimPrefix(vmID, doc.ID+"#")
				}

				verificationMethod := did.VerificationMethod{
					ID:           vmID,
					Type:         cryptosuite.JSONWebKeyType,
					Controller:   controller,
					PublicKeyJWK: jwk,
				}

				doc.VerificationMethod = append(doc.VerificationMethod, verificationMethod)
			}
		}
	}

	getVerificationRelationship := func(key byte) []did.VerificationMethodSet {
		if relationships, ok := getMapValue(cborMap, key).([]any); ok {
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

	if services, ok := getMapValue(cborMap, keyService).([]any); ok {
		for _, svcInterface := range services {
			if svc, ok := svcInterface.(map[any]any); ok {
				service := did.Service{
					ID:              getMapValue(svc, keyServiceID).(string),
					Type:            getMapValue(svc, keyServiceType).(string),
					ServiceEndpoint: getMapValue(svc, keyServiceEndpoint),
				}
				if enc := getMapValue(svc, keyServiceEnc); enc != nil {
					service.Enc = enc
				}
				if sig := getMapValue(svc, keyServiceSig); sig != nil {
					service.Sig = sig
				}
				doc.Services = append(doc.Services, service)
			}
		}
	}

	if controller := getMapValue(cborMap, keyController); controller != nil {
		doc.Controller = controller
	}

	if alsoKnownAs, ok := getMapValue(cborMap, keyAlsoKnownAs).([]any); ok {
		doc.AlsoKnownAs = make([]string, len(alsoKnownAs))
		var akas []string
		for _, aka := range alsoKnownAs {
			akas = append(akas, aka.(string))
		}
		doc.AlsoKnownAs = akas
	}

	if typesInterface := getMapValue(cborMap, keyTypes); typesInterface != nil {
		switch typedTypes := typesInterface.(type) {
		case []any:
			for _, t := range typedTypes {
				switch typedT := t.(type) {
				case uint64:
					types = append(types, TypeIndex(typedT))
				case int64:
					types = append(types, TypeIndex(typedT))
				case float64:
					types = append(types, TypeIndex(typedT))
				}
			}
		case []uint64:
			for _, t := range typedTypes {
				types = append(types, TypeIndex(t))
			}
		case []int64:
			for _, t := range typedTypes {
				types = append(types, TypeIndex(t))
			}
		}
	}

	if gatewaysInterface := getMapValue(cborMap, keyGateways); gatewaysInterface != nil {
		if gws, ok := gatewaysInterface.([]any); ok {
			for _, g := range gws {
				if gatewayString, ok := g.(string); ok {
					gateways = append(gateways, AuthoritativeGateway(gatewayString))
				}
			}
		}
	}

	if prev := getMapValue(cborMap, keyPreviousDID); prev != nil {
		if prevMap, ok := prev.(map[any]any); ok {
			previousDID = &PreviousDID{
				PreviousDID: DHT(getMapValue(prevMap, "did").(string)),
				Signature:   getMapValue(prevMap, "signature").(string),
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

func getMapValue(m map[any]any, key any) any {
	if v, ok := m[key]; ok {
		return v
	}

	// If the key is a byte, try to find it as a uint64
	if byteKey, ok := key.(byte); ok {
		if v, ok := m[uint64(byteKey)]; ok {
			return v
		}
	}

	// If the key is a uint64, try to find it as a byte
	if uint64Key, ok := key.(uint64); ok {
		if v, ok := m[byte(uint64Key)]; ok {
			return v
		}
	}

	return nil
}
