import chai from 'chai';
import {expect} from 'chai';
import chaiAsPromised from 'chai-as-promised';
import {DidDhtMethod} from '../src/did-dht.js';
import {DidKeySetVerificationMethodKey} from "@web5/dids";

chai.use(chaiAsPromised);

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
            expect(keySet).to.have.property('verificationMethodKeys');
            expect(keySet).to.not.have.property('recoveryKey');
            expect(keySet).to.not.have.property('updateKey');
            expect(keySet).to.not.have.property('signingKey');
            expect(keySet.verificationMethodKeys).to.have.lengthOf(1);
            expect(keySet.verificationMethodKeys[0].publicKeyJwk.kid).to.equal('0');
        });

        it('should generate a keyset with an identity keyset passed in (wrong kid)', async () => {
            const ed25519KeyPair = await DidDhtMethod.generateJwkKeyPair({keyAlgorithm: 'Ed25519'});
            expect(DidDhtMethod.generateKeySet({keySet: {identityKey: ed25519KeyPair}})).to.be.rejectedWith('The identity key must have a kid of 0');
        });

        it('should generate a keyset with an identity keyset passed in (correct kid)', async () => {
            const ed25519KeyPair = await DidDhtMethod.generateJwkKeyPair({keyId: '0', keyAlgorithm: 'Ed25519'});
            const keySet = await DidDhtMethod.generateKeySet({keySet: {identityKey: ed25519KeyPair}});
            expect(keySet).to.exist;
            expect(keySet).to.have.property('identityKey');
            expect(keySet).to.have.property('verificationMethodKeys');
            expect(keySet).to.not.have.property('recoveryKey');
            expect(keySet).to.not.have.property('updateKey');
            expect(keySet).to.not.have.property('signingKey');
            expect(keySet.verificationMethodKeys).to.have.lengthOf(1);
            expect(keySet.verificationMethodKeys[0].publicKeyJwk.kid).to.equal('0');
        });

        it('should generate a keyset with a non identity keyset passed in', async () => {
            const ed25519KeyPair = await DidDhtMethod.generateJwkKeyPair({keyAlgorithm: 'Ed25519'});
            const vm: DidKeySetVerificationMethodKey = {
                publicKeyJwk: ed25519KeyPair.publicKeyJwk,
                privateKeyJwk: ed25519KeyPair.privateKeyJwk,
                relationships: ['authentication', 'assertionMethod', 'capabilityInvocation', 'capabilityDelegation']
            }

            const keySet = await DidDhtMethod.generateKeySet({keySet: {verificationMethodKeys: [vm]}});
            expect(keySet).to.exist;
            expect(keySet).to.have.property('identityKey');
            expect(keySet).to.have.property('verificationMethodKeys');
            expect(keySet).to.not.have.property('recoveryKey');
            expect(keySet).to.not.have.property('updateKey');
            expect(keySet).to.not.have.property('signingKey');
            expect(keySet.verificationMethodKeys).to.have.lengthOf(2);

            if (keySet.verificationMethodKeys[0].publicKeyJwk.kid === '0') {
                expect(keySet.verificationMethodKeys[1].publicKeyJwk.kid).to.not.equal('0');
            } else {
                expect(keySet.verificationMethodKeys[1].publicKeyJwk.kid).to.equal('0');
            }
        });
    });

    describe('dids', async () => {
        it('should generate a did identifier given a public key jwk', async () => {
            const ed25519KeyPair = await DidDhtMethod.generateJwkKeyPair({keyAlgorithm: 'Ed25519'});
            const did = await DidDhtMethod.getDidIdentifier({key: ed25519KeyPair.publicKeyJwk});
            expect(did).to.exist;
            expect(did).to.contain('did:dht:');
        });

        it('should create a did document without options', async () => {
            const {did, keySet} = await DidDhtMethod.create();

            console.log(did);

            expect(did).to.exist;
            expect(did.id).to.contain('did:dht:');
            expect(did.verificationMethod).to.exist;
            expect(did.verificationMethod).to.have.lengthOf(1);
            expect(did.verificationMethod[0].id).to.equal(`${did.id}#0`);
            expect(did.verificationMethod[0].publicKeyJwk).to.exist;
            expect(did.verificationMethod[0].publicKeyJwk.kid).to.equal('0');

            expect(keySet).to.exist;
            expect(keySet.identityKey).to.exist;
            expect(keySet.identityKey.publicKeyJwk).to.exist;
            expect(keySet.identityKey.privateKeyJwk).to.exist;
            expect(keySet.identityKey.publicKeyJwk.kid).to.equal('0');
        });
    });
});
