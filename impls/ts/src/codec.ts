import brotli from 'brotli-compress';
import b4a from 'b4a';

export class Codec {
  /**
     * @param records An array of arrays containing strings to compress
     */
  async compress(records: string[][]): Promise<Uint8Array> {
    const string = JSON.stringify(records);
    const toCompress = b4a.from(string);
    return await brotli.compress(toCompress);
  }

  /**
     * @param compressed A Uint8Array containing the compressed data
     */
  async decompress(compressed: Uint8Array): Promise<string[][]> {
    const decompressed = await brotli.decompress(compressed);
    const string = b4a.toString(b4a.from(decompressed));
    return JSON.parse(string);
  }
}
