import {expect} from 'chai';
import {DidDht} from "../src/dht.js";
import {DidDhtMethod} from "../src/did-dht.js";
import {Jose} from "@web5/crypto";
describe('DHT', async function () {
    this.timeout(100000); // 15 seconds

    const dht = new DidDht();
    after(() => {
        dht.destroy();
    });

    it('should put and get data from DHT', async () => {
        const {did, keySet} = await DidDhtMethod.create();
        const publicCryptoKey = await Jose.jwkToCryptoKey({key: keySet.identityKey.publicKeyJwk});
        const privateCryptoKey = await Jose.jwkToCryptoKey({key: keySet.identityKey.privateKeyJwk});

        const request = await dht.createPutDidRequest({
            publicKey: publicCryptoKey,
            privateKey: privateCryptoKey
        }, did);

        const hash = await dht.put(request);

        const retrievedValue = await dht.get(hash);

        const gotDid = await dht.parseGetDidResponse(retrievedValue);
        expect(gotDid).to.deep.equal(did);
    });
});
