import {expect} from 'chai';
import {DidDht} from "../src/dht.js";
import {DidDhtMethod} from "../src/did-dht.js";
import {Jose} from "@web5/crypto";

;
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

        const request = await dht.createPutRequest({
            publicKey: publicCryptoKey,
            privateKey: privateCryptoKey
        }, did);

        const hash = await dht.put(request);
        console.log("DID: ", did.id)
        console.log("HASH: ", hash);

        const retrievedValue = await dht.get(hash);
        console.log(retrievedValue);
    });

    it('should get' , async () => {
        const got = await dht.get('f8c55b5e0c4ff220a6f2be7f4feb01ecf3cdf358');
        console.log(got);
    })
});
