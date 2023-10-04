import {expect} from 'chai';
import {DidDhtMethod} from '../src/did-dht.js';

describe('did-dht', async () => {
  describe('keypairs', async () => {
    it('should generate a keypair', async () => {
      const ed25519KeyPair = await DidDhtMethod.generateJwkKeyPair({keyAlgorithm: 'Ed25519'});
      expect(ed25519KeyPair).to.exist;
      expect(ed25519KeyPair).to.have.property('privateKeyJwk');
      expect(ed25519KeyPair).to.have.property('publicKeyJwk');
      expect(ed25519KeyPair.publicKeyJwk.kid).to.exist;
      expect(ed25519KeyPair.publicKeyJwk.alg).to.equal('EdDSA');
      expect(ed25519KeyPair.publicKeyJwk.kty).to.equal('OKP');

      const secp256k1KeyPair = await DidDhtMethod.generateJwkKeyPair({keyAlgorithm: 'secp256k1'});
      expect(secp256k1KeyPair).to.exist;
      expect(secp256k1KeyPair).to.have.property('privateKeyJwk');
      expect(secp256k1KeyPair).to.have.property('publicKeyJwk');
      expect(secp256k1KeyPair.publicKeyJwk.kid).to.exist;
      expect(secp256k1KeyPair.publicKeyJwk.alg).to.equal('ES256K');
      expect(secp256k1KeyPair.publicKeyJwk.kty).to.equal('EC');
    });
  });

  describe('keysets', async () => {
    it('should generate a keyset with no keyset passed in', async () => {
        const keySet = await DidDhtMethod.generateKeySet();
        expect(keySet).to.exist;
        expect(keySet).to.have.property('identityKey');
        expect(keySet).to.not.have.property('recoveryKey');
        expect(keySet).to.not.have.property('updateKey');
        expect(keySet).to.not.have.property('signingKey');
    });

    it('should generate a keyset with a keyset passed in', async () => {

    });
  });

  describe('dids', async () => {
    it('should generate a did identifier', async () => {

    });

    it('should create a did document', async () => {

    });
  });
});
