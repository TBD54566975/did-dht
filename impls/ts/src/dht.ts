import DHT from 'bittorrent-dht'
import crypto from 'crypto'

const DEFAULT_BOOTSTRAP = [
    'router.magnets.im:6881',
    'router.bittorrent.com:6881',
    'router.utorrent.com:6881',
    'dht.transmissionbt.com:6881',
    'router.nuh.dev:6881'
].map(addr => {
    const [host, port] = addr.split(':')
    return { host, port: Number(port) }
})

export class DHTWrapper {
    private dht: DHT;

    constructor(options?: { bootstrap?: { host: string, port: number }[] }) {
        options.bootstrap = options?.bootstrap ?? DEFAULT_BOOTSTRAP;
        this.dht = new DHT(options);
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
