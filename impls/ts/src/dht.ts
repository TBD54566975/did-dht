import DHT from 'bittorrent-dht'
import ed from 'bittorrent-dht-sodium'

import crypto from 'crypto'

export class DHTWrapper {
    private dht: DHT;

    constructor() {
        this.dht = new DHT();
        this.dht.listen(20000, () => {
            console.log('DHT is listening on port 20000');
        });
    }

    put(value: Buffer): Promise<string> {
        return new Promise((resolve, reject) => {
            this.dht.put({ v: value }, (err, hash) => {
                if (err) {
                    reject(err);
                } else {
                    resolve(hash.toString('hex'));
                }
            });
        });
    }

    get(hash: string): Promise<Buffer> {
        return new Promise((resolve, reject) => {
            this.dht.get(Buffer.from(hash, 'hex'), (err, res) => {
                if (err) {
                    reject(err);
                } else {
                    resolve(res.v);
                }
            });
        });
    }

    destroy(): void {
        this.dht.destroy();
    }
}

/**
 * @param {Uint8Array} input
 */
function hash(input: Uint8Array): Uint8Array {
    return crypto.createHash('sha1').update(input).digest()
}
