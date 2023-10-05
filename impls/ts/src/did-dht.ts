import type {JwkKeyPair, PublicKeyJwk} from '@web5/crypto';
import {EcdsaAlgorithm, EdDsaAlgorithm, Jose, Web5Crypto} from '@web5/crypto';
import type {
    DidDocument,
    DidKeySetVerificationMethodKey,
    DidMethod,
    DidResolutionResult,
    DidService,
    VerificationRelationship
} from '@web5/dids';
import z32 from 'z32';

const SupportedCryptoKeyTypes = [
    'Ed25519',
    'secp256k1'
] as const;

export type DidDhtCreateOptions = {
    publish?: boolean;
    keySet?: DidDhtKeySet;
    services?: DidService[];
}

export type DidDhtKeySet = {
    identityKey?: JwkKeyPair;
    verificationMethodKeys?: DidKeySetVerificationMethodKey[];
}

export type DidDhtContainer = {
    did: DidDocument;
    keySet: DidDhtKeySet;
}

export class DidDhtMethod implements DidMethod {

    public static methodName = 'dht';

    public static async create(options?: DidDhtCreateOptions): Promise<DidDhtContainer> {
        const {publish, keySet: initialKeySet, services} = options ?? {};

        // Generate missing keys if not provided in the options
        const keySet = await this.generateKeySet({keySet: initialKeySet});

        // Get the identifier and set it
        const id = await this.getDidIdentifier({key: keySet.identityKey.publicKeyJwk});

        // add all other keys to the verificationMethod and relationship arrays
        const relationshipsMap: Partial<Record<VerificationRelationship, string[]>> = {};
        const verificationMethods = keySet.verificationMethodKeys.map(key => {
            for (const relationship of key.relationships) {
                if (relationshipsMap[relationship]) {
                    relationshipsMap[relationship].push(`#${key.publicKeyJwk.kid}`);
                } else {
                    relationshipsMap[relationship] = [`#${key.publicKeyJwk.kid}`]
                }
            }

            return {
                id: `${id}#${key.publicKeyJwk.kid}`,
                type: 'JsonWebKey2020',
                controller: id,
                publicKeyJwk: key.publicKeyJwk
            };
        });

        const did: DidDocument = {
            id,
            verificationMethod: [...verificationMethods],
            ...relationshipsMap,
            ...services && {service: services}
        };
        return {did, keySet};
    }

    public static async publish(key: DidDhtKeySet, didDocument: DidDocument): Promise<DidResolutionResult | undefined> {
        throw new Error('Method not implemented.');
    }

    public static async resolve(did: string): Promise<DidDocument> {
        throw new Error('Method not implemented.');
    }

    public static async getDidIdentifier(options: {
        key: PublicKeyJwk
    }): Promise<string> {
        const {key} = options;

        const cryptoKey = await Jose.jwkToCryptoKey({key});
        const identifier = z32.encode(cryptoKey.material);
        return 'did:dht:' + identifier;
    }

    public static async generateJwkKeyPair(options: {
        keyAlgorithm: typeof SupportedCryptoKeyTypes[number],
        keyId?: string
    }): Promise<JwkKeyPair> {
        const {keyAlgorithm, keyId} = options;

        let cryptoKeyPair: Web5Crypto.CryptoKeyPair;

        switch (keyAlgorithm) {
            case 'Ed25519': {
                cryptoKeyPair = await new EdDsaAlgorithm().generateKey({
                    algorithm: {name: 'EdDSA', namedCurve: 'Ed25519'},
                    extractable: true,
                    keyUsages: ['sign', 'verify']
                });
                break;
            }

            case 'secp256k1': {
                cryptoKeyPair = await new EcdsaAlgorithm().generateKey({
                    algorithm: {name: 'ECDSA', namedCurve: 'secp256k1'},
                    extractable: true,
                    keyUsages: ['sign', 'verify']
                });
                break;
            }

            default: {
                throw new Error(`Unsupported crypto algorithm: '${keyAlgorithm}'`);
            }
        }

        // Convert the CryptoKeyPair to JwkKeyPair.
        const jwkKeyPair = await Jose.cryptoKeyToJwkPair({keyPair: cryptoKeyPair});

        // Set kid values.
        if (keyId) {
            jwkKeyPair.privateKeyJwk.kid = keyId;
            jwkKeyPair.publicKeyJwk.kid = keyId;
        } else {
            // If a key ID is not specified, generate RFC 7638 JWK thumbprint.
            const jwkThumbprint = await Jose.jwkThumbprint({key: jwkKeyPair.publicKeyJwk});
            jwkKeyPair.privateKeyJwk.kid = jwkThumbprint;
            jwkKeyPair.publicKeyJwk.kid = jwkThumbprint;
        }

        return jwkKeyPair;
    }

    public static async generateKeySet(options?: {
        keySet?: DidDhtKeySet
    }): Promise<DidDhtKeySet> {
        let {keySet = {}} = options ?? {};

        if (!keySet.identityKey) {
            keySet.identityKey = await this.generateJwkKeyPair({
                keyAlgorithm: 'Ed25519',
                keyId: '0'
            });


        } else if (keySet.identityKey.publicKeyJwk.kid !== '0') {
            throw new Error('The identity key must have a kid of 0');
        }

        // add verificationMethodKeys for the identity key
        const identityKeySetVerificationMethod: DidKeySetVerificationMethodKey = {
            ...keySet.identityKey,
            relationships: ['authentication', 'assertionMethod', 'capabilityInvocation', 'capabilityDelegation']
        }

        if (!keySet.verificationMethodKeys) {
            keySet.verificationMethodKeys = [identityKeySetVerificationMethod];
        } else if (keySet.verificationMethodKeys.filter(key => key.publicKeyJwk.kid === '0').length === 0) {
            keySet.verificationMethodKeys.push(identityKeySetVerificationMethod);
        }

        // Generate RFC 7638 JWK thumbprints if `kid` is missing from any key.
        if (keySet.verificationMethodKeys) {
            for (const key of keySet.verificationMethodKeys) {
                if (key.publicKeyJwk) key.publicKeyJwk.kid ??= await Jose.jwkThumbprint({key: key.publicKeyJwk});
                if (key.privateKeyJwk) key.privateKeyJwk.kid ??= await Jose.jwkThumbprint({key: key.privateKeyJwk});
            }
        }

        return keySet;
    }
}
