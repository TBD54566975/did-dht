import {expect} from 'chai';
import {Codec} from "../src/codec.js";
import b4a from 'b4a';

describe('Codec', async () => {
    const codec = new Codec();

    it('compresses and decompresses an uncompressable value', async () => {
        const uncompressable = [['2'], ['3']];
        const compressed = await codec.compress(uncompressable);
        const decompressed = await codec.decompress(compressed);

        const uncompressableBuffer = Buffer.from(JSON.stringify(uncompressable));
        const compressedBuffer = Buffer.from(compressed);
        expect(compressedBuffer.length).to.be.greaterThanOrEqual(uncompressableBuffer.length);
        expect(decompressed).to.deep.equal(uncompressable);
    });

    it('compresses and decompresses a compressable value', async () => {
        const records = [['did', '{"@context":"https://w3id.org/did/v1", "id": "did:example:123"}']];
        const compressed = await codec.compress(records);
        const decompressed = await codec.decompress(compressed);

        const uncompressableBuffer = Buffer.from(JSON.stringify(records));
        const compressedBuffer = Buffer.from(compressed);
        expect(compressedBuffer.length).to.be.lessThan(uncompressableBuffer.length);
        expect(decompressed).to.deep.equal(records);
    });
});
