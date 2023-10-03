import brotli from 'brotli-compress';
import b4a from 'b4a';

export class Codec {
  /**
     * @param records An array of arrays containing strings
     */
  async encode(records: string[][]): Promise<Uint8Array> {
    const string = JSON.stringify(records);
    const encoded = b4a.from(string);
    return await brotli.compress(encoded);
  }

  /**
     * @param encoded A Uint8Array containing the encoded data
     */
  async decode(encoded: Uint8Array): Promise<string[][]> {
    const decoded = await brotli.decompress(encoded);
    const string = b4a.toString(b4a.from(decoded));
    return JSON.parse(string);
  }
}
