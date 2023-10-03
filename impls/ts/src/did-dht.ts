import type { PrivateKeyJwk, PublicKeyJwk, JwkKeyPair } from '@web5/crypto';
import type {
  DidDocument,
  DidKeySetVerificationMethodKey,
  DidMethod,
  DidResolutionResult,
  DidService,
  PortableDid
} from '@web5/dids';
import {EcdsaAlgorithm, EdDsaAlgorithm, Jose, Web5Crypto} from '@web5/crypto';

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

  public static async create(options?: DidDhtCreateOptions): Promise<PortableDid> {
    let { publish, keySet, services } = options ?? { };

    // Begin constructing a PortableDid
    const did: Partial<PortableDid> = {};

    // Generate missing keys if not provided in the options
    did.keySet = await this.generateKeySet({ keySet });

    // Get the identifier
    throw new Error('Method not implemented.');
  }

  public static async publish(key: DidDhtKeySet, didDocument: DidDocument): Promise<DidResolutionResult | undefined> {
    throw new Error('Method not implemented.');
  }

  public static async resolve(did: string): Promise<DidDocument> {
    throw new Error('Method not implemented.');
  }

  public static async generateJwkKeyPair(options: {
        keyAlgorithm: typeof SupportedCryptoAlgorithms[number],
        keyId?: string
    }): Promise<JwkKeyPair> {
    const { keyAlgorithm, keyId } = options;

    let cryptoKeyPair: Web5Crypto.CryptoKeyPair;

    switch (keyAlgorithm) {
      case 'Ed25519': {
        cryptoKeyPair = await new EdDsaAlgorithm().generateKey({
          algorithm   : { name: 'EdDSA', namedCurve: 'Ed25519' },
          extractable : true,
          keyUsages   : ['sign', 'verify']
        });
        break;
      }

      case 'secp256k1': {
        cryptoKeyPair = await new EcdsaAlgorithm().generateKey({
          algorithm   : { name: 'ECDSA', namedCurve: 'secp256k1' },
          extractable : true,
          keyUsages   : ['sign', 'verify']
        });
        break;
      }

      default: {
        throw new Error(`Unsupported crypto algorithm: '${keyAlgorithm}'`);
      }
    }

    // Convert the CryptoKeyPair to JwkKeyPair.
    const jwkKeyPair = await Jose.cryptoKeyToJwkPair({ keyPair: cryptoKeyPair });

    // Set kid values.
    if (keyId) {
      jwkKeyPair.privateKeyJwk.kid = keyId;
      jwkKeyPair.publicKeyJwk.kid = keyId;
    } else {
      // If a key ID is not specified, generate RFC 7638 JWK thumbprint.
      const jwkThumbprint = await Jose.jwkThumbprint({ key: jwkKeyPair.publicKeyJwk });
      jwkKeyPair.privateKeyJwk.kid = jwkThumbprint;
      jwkKeyPair.publicKeyJwk.kid = jwkThumbprint;
    }

    return jwkKeyPair;
  }

  public static async generateKeySet(options?: {
        keySet?: DidDhtKeySet
    }): Promise<DidDhtKeySet> {
    let { keySet = {} } = options ?? { };

    if (!keySet.identityKey) {
      const identityKeyPair = await this.generateJwkKeyPair({
        keyAlgorithm : 'Ed25519',
        keyId        : '0'
      });
      keySet.identityKey = identityKeyPair;
      keySet.verificationMethodKeys = [{
        ...identityKeyPair,
        relationships: ['authentication', 'assertionMethod', 'capabilityInvocation', 'capabilityDelegation']
      }];
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
