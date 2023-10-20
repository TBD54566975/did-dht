# Pkarr DID Method Specification

# Introduction
A [DID Method](https://www.w3.org/TR/did-core/) based on [Pkarr](https://github.com/nuhvi/pkarr), identified by `did:dht`.

Pkarr stands for **P**ublic **K**ey **A**ddressable **R**esource **R**ecords. Pkarr makes use of the
[Mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT), specifically [BEP44](https://www.bittorrent.org/beps/bep_0044.html)
to store records. This DID method uses Pkarr to store _DID Documents_.

As Pkarr states, mainline is chosen for the following reasons (paraphased):

> 1. It has a proven track record of 15 years.
> 2. It is the biggest DHT in existence with an estimated 10 million nodes.
> 3. It is fairly generous in retaining data.
> 4. It has implementation in most languagues, and is stable.

# Format

The format for `did:dht` conforms to the [DID Core](https://www.w3.org/TR/did-core/) specification. It consists of the
`did:dht` prefix followed by the Pkarr identifier. The Pkarr identifier is a [z-base-32](https://en.wikipedia.org/wiki/Base32#z-base-32)
encoded [Ed25519](https://ed25519.cr.yp.to/) public key.

```
did-dht-format := did:dht:<z-base-32-value>
z-base-32-value := [a-km-uw-z13-9]+
```

Alternatively, the encoding rules can be thought of as a series of transformation functions on the raw public key bytes:

```
did-dht-format := did:dht:z-base-32(raw-public-key-bytes)
```

# Operations

Entries to the DHT require a signed record. As such, the keypair that is generated for the pkarr identifier is used to
sign the DHT record. This keypair **MUST** always be present in a `did:dht` document and is referred to as the _identifier key_.

## Create

To create a `did:dht`, the process is as follows:

1. Create an [Ed25519](https://ed25519.cr.yp.to/) keypair and encode the public key using the format provided in
the [prior section](#Format).

2. Construct a compliant JSON representation of a DID Document
 a. The document **MUST** include a [verification method](https://www.w3.org/TR/did-core/#verification-methods) with
 the _identifier key_ encoded as a `publicKeyJwk` with an `id` of ``#0`` and `type` of
 [`JsonWebKey2020`](https://www.w3.org/community/reports/credentials/CG-FINAL-lds-jws2020-20220721/#json-web-key-2020).
 b. The document can include any number of other [core properties](https://www.w3.org/TR/did-core/#core-properties);
 always representing key material as a [JWK](https://datatracker.ietf.org/doc/html/rfc7517).

3. Map the output DID Document to a DNS packet as outlined in the [section below](#DNS-Packet-DID-document). 

4. Construct a [BEP44 put message](https://www.bittorrent.org/beps/bep_0044.html) with the `v` value as a DNS packet
from the previous step.

5. Submit the result of (3) to the DHT, a Pkarr, or DID DHT service.

### DIDs as a DNS Packet

 | name             | type      | TTL | rdata                                     |
 |------------------|-----------|-----|-------------------------------------------|
 | `_did`           | TXT       |  -  | `id=o4dksfbqk85ogzdb5osziw6befigbuxmuxkuxq8434q89uj56uyy` |
 | `_key1._did`     | TXT       |  -  | `t=0,k=<b64url>`                            |
 | `_key2._did`     | TXT       |  -  | `t=1,k=<b64url>`                            |
 | `_key3._did`     | TXT       |  -  | `t=1,k=<b64url>`                            |
 | `_srv1._did`     | TXT       |  -  | `t=LinkedDomains;uri=foo.com;...`           |
 | `_srv2._did`     | TXT       |  -  | `t=DWN;uri=https://dwn.tbddev.org/dwn5;...` |

    
In this scheme, we encode the DID document as multiple TXT records.

A root `_did` includes the values of `id` (the DID). Other records such as `srv` (services), `vm` (verification methods),
and verification relationships are surfaced as additional records of the format `<ID>._did`, which contains the value of
each `key` or `service` as attributes.
    
All records must end in `_did` or `_did.TLD` if a TLD is being used with the record.

It might look like repeating `_did` is an overhead, but these can be compressed away using normal DNS standard
[packet compression](https://courses.cs.duke.edu/fall16/compsci356/DNS/DNS-primer.pdf) techniques.

The DNS packet must set the `Authoritative Answer` flag, since this is always an `Authoritative` packet.

The DID identifier z-base32 key should be appended as the Origin of all records, which won't cost much thanks to the
name compression in DNS packets, so `_did` should be saved as `_did.o4dksfbqk85ogzdb5osziw6befigbuxmuxkuxq8434q89uj56uyy`,
this should make caching and responding with to DNS requests faster as they are already in the shape of a DNS packet,
albeit for all types and subdomains.

#### Property Mapping

The _root record_, `_did` or `_did.TLD` if a TLD is being utilized contains a list of IDs of the keys and service
endpoints used in different sections of the DID Document.

An example is as follows:

 | name        | rdata                                                    |
 |-------------|----------------------------------------------------------|
 | `_did.TLD`  | `vm=k1,k2,k3;auth=k1;asr=k2;cin=k3;cdel=k3;srv=s1,s2,s3` |
 
-------------------------------------------------------------------------
The following instructions serve as a reference of mapping DID Document properties to DNS `TXT` records:

**1. Verification Methods**

Each Verification Method _name_ is represented as a `_kN._did` record where `N` is the positional index of the
Verification Method (e.g. `_k0`, `_k1`).

Each Verification Method _data_ is represented with the form `id=M,t=N,k=O` where `M` is the key's ID, `N` is the index
of the key's type from the table below, and `O` is the base64URL representation of the public key.

| index | key type    |
| ----- | ----------- |
| 0     | `ed25519`   |
| 1     | `secp256k1` |
| 2     | `secp256r1` |


An example is as follows:
    
| name       | rdata                                                       |
|------------|-------------------------------------------------------------|
| `_k1._did` | `id=abcd,t=0,k=r96mnGNgWGOmjt6g_3_0nd4Kls5-kknrd4DdPW8qtfw` |

-------------------------------------------------------------------------

**2. Verification Relationships**

Each Verification Relationship, if utilized, is represented as a part of the root `_did.TLD` record. 

The following table maps Verification Relationship types to their record name.

| Verification Relationship  | Record Name |
| ------------------------- | ---------- |
| Authentication            | `auth`     |
| Assertion                 | `asm`      |
| Key Agreement             | `agm`      |
| Capability Invocation     | `inv`      |
| Capability Delegation     | `del`      |

The record data is uniform across verification relationships, a comma separated list of key references:

An example is as follows:
 | Verification Relationship  |  rdata in the root record                    |
 |---------------------------|----------------------------------------------|
 | `"authentication": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"]` | `auth=0,HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ` |
 | `"assertionMethod": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"]`| `asm=0,HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ`  |
 | `"keyAgreement": ["#1"]`              | `agm=1`      |
 | `"capabilityInvocation": ["#0"]`      | `inv=0`      |
 | `"capabilityDelegation": ["#0"]`      | `del=0`      |

    
-----------------------------------------------------

**3. Services**

Each Service, if utilized, is represented  name is represented as a `_sN._did` record where N is the positional index
of the Service (e.g. `_s0`, `_s1`). Each Service _data_ is represented with the form `id=M,t=N,uri=O` where `M` is the
Service's ID, `N` is the Service's Type and `O` is the Service's URI.
    
An example is given as follows:

 | name       |  rdata                                                      |
 |------------|-------------------------------------------------------------|
 | `_s0._did` | `id=dwn,t=DecentralizedWebNode,uri=https://example.com/dwn` |

    
Each Service is also represented as part of the root `_did.TLD` record as a list under the key `srv=<ids>` where `ids`
is a comma separate list of all IDs for each Service.

-----------------------------------------------------

A sample transformation is provided of the following DID Document:

```json
{
  "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
  "verificationMethod": [
    {
      "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0",
      "type": "JsonWebKey2020",
      "controller": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
      "publicKeyJwk": {
        "alg": "EdDSA",
        "crv": "Ed25519",
        "kty": "OKP",
        "ext": "true",
        "key_ops": [
          "verify"
        ],
        "x": "r96mnGNgWGOmjt6g_3_0nd4Kls5-kknrd4DdPW8qtfw",
        "kid": "0"
      }
    },
    {
      "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ",
      "type": "JsonWebKey2020",
      "controller": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
      "publicKeyJwk": {
        "alg": "ES256K",
        "crv": "secp256k1",
        "kty": "EC",
        "ext": "true",
        "key_ops": [
          "verify"
        ],
        "x": "KI0DPvL5cGvznc8EDOAA5T9zQfLDQZvr0ev2NMLcxDw",
        "y": "0iSbXxZo0jIFLtW8vVnoWd8tEzUV2o22BVc_IjVTIt8",
        "kid": "HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"
      }
    }
  ],
  "authentication": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"],
  "assertionMethod": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"],
  "capabilityInvocation": ["#0"],
  "capabilityDelegation": ["#0"],
  "service": [
    {
      "id": "#dwn",
      "type": "DecentralizedWebNode",
      "serviceEndpoint": "https://example.com/dwn"
    }
  ]
}
```

All records are of type `TXT` with an expiry of `7200` to align with the DHT's standard 2-hour expiry window.

 | name       | rdata                                                                          |
 |------------|--------------------------------------------------------------------------------|
 | `_did.TLD` | `vm=k0,k1;auth=k0,k1;asm=k0,k1;inv=k0;del=k0;srv=s1`                           |
 | `_k0._did` | `id=0,t=0,h=afdea69c63605863a68edea0ff7ff49dde0a96ce7e9249eb7780dd3d6f2ab5fc`  |
 | `_k1._did` | `id=HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ,t=1,k=BCiNAz7y-XBr853PBAzgAOU_c0Hyw0Gb69Hr9jTC3MQ80iSbXxZo0jIFLtW8vVnoWd8tEzUV2o22BVc_IjVTIt8` |
 | `_s0._did` | `id=dwn,t=DecentralizedWebNode,uri=https://example.com/dwn`                    |
 
## Read

To read a `did:dht`, the process is as follows:

1. Take the suffix of the DID, that is, the _encoded identifier key_, and pass it to a Pkarr resolver.
2. Decode the resulting [BEP44 response](https://www.bittorrent.org/beps/bep_0044.html)'s `v` value using [bencode](https://en.wikipedia.org/wiki/Bencode).
3. Reverse the DNS packet process outlined above and re-construct a DID Document.


## Update

Each write to the DHT is considered an update. As long as control of the _identity key_ is retained any update is
possible with a unique sequence number with [mutable items](https://www.bittorrent.org/beps/bep_0044.html) using BEP44.

## Deactivate

To deactivate a document there are two options:
    
1. Let the DHT record expire and cease to publish it.
2. Publish a new DHT record where the `rdata` of the root DNS record is the string "deactivated"

    
 | name      | rdata        |
 |-----------|--------------|
 | _did.TLD  | deactivated  |


# Bitcoin-anchored Gateways

**(TODO): service endpoints / api**

To be recognized as a DID DHT retention gateway, the gateway operator must anchor a transaction on Bitcoin that
timelocks Bitcoin value proportional to the number of DIDs they introduce into the _Retained DID Set_.

The amount of value locked must be no less than the mean value of the upper half of UTXOs for the block in which the
timelock takes effect, and the lock must be a *relative timelock* set to 1000 blocks.

# Retained DID Set

### Generating a Retention Proof

Perform Proof of Work over the DID's identifier + the `retention` value of a given DID operation (composed of the
selected bitcoin block hash and nonce). The resulting Retention Proof Hash determines the duration of retention based
 on the number of leading zeros of the hash, which must be no less than 26.

### Managing the Retained DID Set

Nodes following the Retention Set rules SHOULD sort DIDs they are retaining by age of retention proof, followed by
number of retention proof leading 0s. When a node needs to reduce its retained set of DID entries, it SHOULD remove
entries from the bottom of the list in accordance with this sort.

### Reporting on Retention Status

Nodes MUST include the approximate time until retention fall-off in the Method-specific metadata of a resolved DID
Document, to aid in Identity Agents (wallets) being able to assess whether resubmission is required.

# Implementation Considerations

Data needs to be republished.

# Security and Privacy Considerations

## Security

## Privacy
    
# Implementation Considerations

Data needs to be republished.

# Security and Privacy Considerations

## Security

## Privacy
