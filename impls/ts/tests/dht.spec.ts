import {expect} from 'chai';
import {DHTWrapper} from "../src/dht.js";

describe('DHT',  async function() {
    this.timeout(10000); // 10 seconds

    const dht = new DHTWrapper();
    after(() => {
        dht.destroy();
    });

    it('should put and get data from DHT', async () => {
        const value = Buffer.from('Hello, DHT!');
        const hash = await dht.put(value);

        const retrievedValue = await dht.get(hash);
        expect(retrievedValue.toString()).to.equal(value.toString());
        console.log(retrievedValue.toString());
    });
});
