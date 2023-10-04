import type {JwkKeyPair, PublicKeyJwk} from '@web5/crypto';
import {EcdsaAlgorithm, EdDsaAlgorithm, Jose, Web5Crypto} from '@web5/crypto';
import type {
    DidDocument,
    DidKeySetVerificationMethodKey,
    DidMethod,
    DidResolutionResult,
    DidService,
    VerificationMethod
} from '@web5/dids';
import z32 from 'z32';

const SupportedCryptoAlgorithms = [
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

export class DidDhtMethod implements DidMethod {

    public static methodName = 'dht';

    public static async create(options?: DidDhtCreateOptions): Promise<{ did: DidDocument, keySet: DidDhtKeySet }> {
        let {publish, keySet, services} = options ?? {};

        // Begin constructing a PortableDid
        const did: Partial<DidDocument> = {};

        // Generate missing keys if not provided in the options
        keySet = await this.generateKeySet({keySet});

        // Get the identifier and set it
        const id = await this.getDidIdentifier({key: keySet.identityKey.publicKeyJwk});
        did.id = id;

        // add identity key to the verificationMethod and relationship arrays
        const identityKey: Partial<VerificationMethod> = {
            id: `${id}#0`,
            type: 'JsonWebKey2020',
            controller: id,
            publicKeyJwk: keySet.identityKey.publicKeyJwk
        };
        did.verificationMethod = [identityKey as VerificationMethod];

        // add all other keys to the verificationMethod and relationship arrays
        const relationshipsMap: { [key: string]: string[] } = {
            authentication: ['#0'],
            assertionMethod: ['#0'],
            capabilityInvocation: ['#0'],
            capabilityDelegation: ['#0'],
            keyAgreement: []
        };

        for (const key of keySet.verificationMethodKeys) {
            const vm: Partial<VerificationMethod> = {
                id: `${id}#${key.publicKeyJwk.kid}`,
                type: 'JsonWebKey2020',
                controller: id,
                publicKeyJwk: key.publicKeyJwk
            };

            for (const relationship of key.relationships) {
                if (relationshipsMap[relationship]) {
                    relationshipsMap[relationship].push(key.publicKeyJwk.kid);
                }
            }

            did.verificationMethod.push(vm as VerificationMethod);
        }

        // add all services
        if (services) {
            did.service = services;
        }
        return {did: did as DidDocument, keySet: keySet};
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
        keyAlgorithm: typeof SupportedCryptoAlgorithms[number],
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

        const identityKeySetVerificationMethod: DidKeySetVerificationMethodKey = {
            ...keySet.identityKey,
            relationships: ['authentication', 'assertionMethod', 'capabilityInvocation', 'capabilityDelegation']
        }

        // add verificationMethodKeys for the identity key
        if (keySet.verificationMethodKeys) {
            keySet.verificationMethodKeys.push(identityKeySetVerificationMethod);
        } else {
            keySet.verificationMethodKeys = [identityKeySetVerificationMethod];
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
