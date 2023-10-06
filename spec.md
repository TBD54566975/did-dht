# DID DHT Method Specification

# Introduction
A [DID Method](https://www.w3.org/TR/did-core/) based on [Mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT) and inspired by [Pkarr](https://github.com/nuhvi/pkarr), this DID method uses Mainline [BEP44](https://www.bittorrent.org/beps/bep_0044.html) DHT entries to store _DID Documents_.

Mainline has the following attributes that make it attractive for use as a foundation for circulation of DHT entries:

> 1. It has a proven track record of 15 years.
> 2. It is the biggest DHT in existence with an estimated 10 million nodes.
> 3. It is fairly generous in retaining data.
> 4. It has implementation in most languages, and is stable.

# Format

The format for `did:dht` conforms to the [DID Core](https://www.w3.org/TR/did-core/) specification. It consists of the `did:dht` prefix followed by the Method-specific identifier, which is a [z-base-32](https://en.wikipedia.org/wiki/Base32#z-base-32) encoded [Ed25519](https://ed25519.cr.yp.to/) public key.

```
did-dht-format := did:dht:<z-base-32-value>
z-base-32-value := [a-km-uw-z13-9]+
```

Alternatively, the encoding rules can be thought of as a series of transformation functions on the raw public key bytes:

```
did-dht-format := did:dht:z-base-32(raw-public-key-bytes)
```

# Operations

Entries to the DHT require a signed record. As such, the keypair that is generated for the Method-specific identifier is used to sign the DHT record. This keypair **MUST** always be present in a `did:dht` document and is referred to as the _identifier key_.

## Create

To create a `did:dht`, the process is as follows:

1. Create an [Ed25519](https://ed25519.cr.yp.to/) keypair and encode the public key using the format provided in the [prior section](#Format).


2. Construct a JSON object composed of the following properties:
   - **types:** array of type strings corresponding with the Entity Type Table (DIF-hosted document)
   - **keys:** an array of key objects of the following format:
      ```javascript
      {
        "id": "0",
        "uses": ["signing"],
        "jwk": {...}
      }
      ```
    - **services**: an array of DID Document-compliant endpoint objects
    - **retention**: optional Retention Proof string, which is a comma separated list of values composed as follows:

      - 32 byte SHA-256 Bitcoin block hash
      - <= 32 byte nonce

      Computing the retention hash: perform Proof of Work over the DID’s identifier + the selected bitcoin block hash + a nonce to compute a value that secures greater retention time in the Retained DID Set maintained by the Bitcoin-anchored gateways.

      See the section Retained DID Set for more information.

> NOTE: The document **MUST** include a `signing` key used for [verification method](https://www.w3.org/TR/did-core/#verification-methods) with the _identifier key_ encoded as a `publicKeyJwk` with an `id` of `0`. All keys will be assumed to be the type [`JsonWebKey2020`](https://www.w3.org/community/reports/credentials/CG-FINAL-lds-jws2020-20220721/#json-web-key-2020).


3. Construct a [BEP44 PUT message](https://www.bittorrent.org/beps/bep_0044.html) with the `v` value set to a [bencoded](https://en.wikipedia.org/wiki/Bencode) dict representation of the JSON object from Step 1.

(TODO) **brotli encode before bencode** - to get max compression
(TODO) Seq number as a timestamp

4. Submit the result to one of the Bitcoin-anchored gateways or another Mainline gateway.

## Read

To read a `did:dht`, the process is as follows:

1. Retrieve the value for the provided DID's identifier from the DHT.
2. Decode the resulting [BEP44 response](https://www.bittorrent.org/beps/bep_0044.html)'s `v` value using [bencode](https://en.wikipedia.org/wiki/Bencode) and transform the bencoded dict to JSON.
3. Project all `keys`, `services`, and other relevant values into a DID Document structure with all its required properties set.

## Update

Repeat the Create process to generate a new Mainline payload with an incremented `seq` value greater than the previous version.

## Deactivate

Since updates are not supported, deactivation can only be accomplished by failing to republish an identifier or setting a new update's `v` value to null

# Bitcoin-anchored Gateways

**(TODO): service endpoints / api**

To be recognized as a DID DHT retention gateway, the gateway operator must anchor a transaction on Bitcoin that timelocks Bitcoin value proportional to the number of DIDs they introduce into the _Retained DID Set_.

The amount of value locked must be no less than the mean value of the upper half of UTXOs for the block in which the timelock takes effect, and the lock must be a *relative timelock* set to 1000 blocks.

# Retained DID Set

### Generating a Retention Proof

Perform Proof of Work over the DID's identifier + the `retention` value of a given DID operation (composed of the selected bitcoin block hash and nonce). The resulting Retention Proof Hash determines the duration of retention based on the number of leading zeros of the hash, which must be no less than 26.

### Managing the Retained DID Set

Nodes following the Retention Set rules SHOULD sort DIDs they are retaining by age of retention proof, followed by number of retention proof leading 0s. When a node needs to reduce its retained set of DID entries, it SHOULD remove entries from the bottom of the list in accordance with this sort.

### Reporting on Retention Status

Nodes MUST include the approximate time until retention fall-off in the Method-specific metadata of a resolved DID Document, to aid in Identity Agents (wallets) being able to assess whether resubmission is required.

# Implementation Considerations

Data needs to be republished.

# Security and Privacy Considerations

## Security

## Privacy