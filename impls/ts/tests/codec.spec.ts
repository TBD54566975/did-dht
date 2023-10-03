import { expect } from 'chai';
import {Codec} from "../src/codec.js";

describe('Codec', async () => {
  const codec = new Codec();

  it('encodes and decodes an uncompressable value', async () => {
    const uncompressable = [['2'], ['3']];
    const encoded = await codec.encode(uncompressable);
    const decoded = await codec.decode(encoded);
    expect(decoded).to.deep.equal(uncompressable);
  });

  it('encodes and decodes a compressable value', async() =>{
  const records = [['did', '{"@context":"https://w3id.org/did/v1", "id": "did:example:123"}']];
    const encoded = await codec.encode(records);
    const decoded = await codec.decode(encoded);
    expect(decoded).to.deep.equal(records);
  });
});
