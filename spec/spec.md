The DID DHT Method Specification 1.0
==================

**Specification Status: Working Draft**

**Latest Draft:** [tbd54566975.github.io/did-dht-method](https://tbd54566975.github.io/did-dht-method)

**Editors:**
~ [Gabe Cohen](https://github.com/decentralgabe)
~ [Daniel Buchner](https://github.com/csuwildcat)

**Participate:**
~ [GitHub repo](https://github.com/TBD54566975/did-dht-method)
~ [File a bug](https://github.com/TBD54566975/did-dht-method/issues)
~ [Commit history](https://github.com/TBD54566975/did-dht-method/commits/main)

## Abstract
A [DID Method](https://www.w3.org/TR/did-core/) based on [Pkarr](https://github.com/nuhvi/pkarr), identified by `did:dht`.

[[ref:Pkarr]] stands for **P**ublic **K**ey **A**ddressable **R**esource **R**ecords. [[ref:Pkarr]] makes use of the
[Mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT), specifically [[ref:BEP44]] to store records. This DID
method uses [[ref:Pkarr]] to store _DID Documents_.

As [[ref:Pkarr]] states, mainline is chosen for the following reasons:

1. It has a proven track record of 15 years.
2. It is the biggest DHT in existence with an estimated 10 million nodes.
3. It is fairly generous in retaining data.
4. It has implementation in most languages, and is stable.

The syntax of the identifier and accompanying data model used by the protocol is conformant with the [[spec:DID-Core]]
specification and shall be registered with the [[spec:DID-Spec-Registries]].

## Terminology

[[def:Decentralized Identifier, Decentralized Identifier, DID]]
~ A [W3C specification](https://www.w3.org/TR/did-core/) describing an _identifier that enables verifiable, decentralized
digital identity_.

[[def:DID Suffix, Suffix]]
~ The unique identifier string within a DID URI. e.g. The unique suffix of `did:dht:123` would be `123`.

[[def:DID Suffix Data, Suffix Data]]
~ Data required to deterministically generate a DID, the [[ref:Identity Key]].

[[def:Identity Key]]
~ An [[ref:Ed25519]] public key encoded with [[ref:z-base-32]] used to uniquely identify a `did:dht` document.

[[def:DID DHT Service]]
~ A service that provides a [[ref:DHT]] interface to the [[ref:Pkarr]] network, extended to support this [[ref:DID]] method.

[[def:Pkarr]]
~ An [open source project](https://github.com/nuhvi/pkarr) which provides "simplest possible streamlined integration
between the Domain Name System and peer-to-peer overlay networks, enabling self-issued public keys to function as
sovereign, publicly addressable domains."

[[def:Mainline DHT, DHT, Mainline]]
~ [Mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT) is the name given to the DHT used by the BitTorrent
protocol. It is a distributed system for storing and finding data on a peer-to-peer network. It is based on
[Kademlia](https://en.wikipedia.org/wiki/Kademlia) and is primarily used to store and retrieve _torrent_ metadata.
It has between 16 and 28 million concurrent users.

[[def:Gateway, DID DHT Node, Bitcoin-anchored Gateway]]
~ A node that acts as a gateway to the DID DHT. The gateway may offer a set of APIs to interact with the DID DHT, such
as features providing guaranteed retention, historical resolution, and other features.

[[def:Retained DID Set]]
~ The set of DIDs that a [[ref:Gateway]] is retaining, and thus is responsible for republishing.

[[def:Retention Proof]]
~ A proof of work that is performed by the [[ref:DID]] controller to prove that they are still in control of the DID. This
proof is used by nodes to determine how long they should retain a DID.

## DID DHT Method Specification

### Format

The format for `did:dht` conforms to the [[spec:DID Core]] specification. It consists of the
`did:dht` prefix followed by the [[ref:Pkarr]] identifier. The [[ref:Pkarr]] identifier is a [[ref:z-base-32]].
encoded [[ref:Ed25519]] public key.

```text
did-dht-format := did:dht:<z-base-32-value>
z-base-32-value := [a-km-uw-z13-9]+
```

Alternatively, the encoding rules can be thought of as a series of transformation functions on the raw public key bytes:

```text
did-dht-format := did:dht:Z-BASE-32(raw-public-key-bytes)
```

### DIDs as a DNS Packet

In this scheme we encode the [[ref:DID Document]] as multiple [DNS TXT records](https://en.wikipedia.org/wiki/TXT_record).
Comprising a DNS packet [[spec:RFC1034]] [[spec:RFC1035]], which is then stored in the [[ref:DHT]].

| Name     | Type | TTL    | Rdata                                     |
| -------- | ---- | ------ | ----------------------------------------- |
| _did     | TXT  |  7200  | vm=k0,k1,k2;auth=k0;asm=k1;inv=k2;del=k2;srv=s0,s1,s2 |
| _k0._did | TXT  |  7200  | t=0,k=`<b64url>`                          |
| _k1._did | TXT  |  7200  | t=1,k=`<b64url>`                          |
| _k2._did | TXT  |  7200  | t=1,k=`<b64url>`                          |
| _s0._did | TXT  |  7200  | t=LinkedDomains;uri=foo.com;...           |
| _s1._did | TXT  |  7200  | t=DWN;uri=https://dwn.tbddev.org/dwn5;... |

::: note
The recommended TTL value is 7200 seconds (2 hours), which is the default TTL for Mainline records.
:::

* A root `_did` record identifies the [property mapping](#property-mapping) for the document. Other records such as `srv`
(services), `vm` (verification methods), and verification relationships (e.g. authentication, assertion, etc.) are surfaced
as additional records of the format `<ID>._did`, which contains the zero-indexed value of each `key` or `service` as
attributes.

* All records **MUST** end in `_did` or `_did.TLD` if a TLD is being used with the record.

::: note
It might look like repeating `_did` is an overhead, but these can be compressed away using normal DNS standard
[packet compression](https://courses.cs.duke.edu/fall16/compsci356/DNS/DNS-primer.pdf) techniques.
:::

* The DNS packet **MUST** set the _Authoritative Answer_ flag, since this is always an _Authoritative_ packet.

* The DID identifier [[ref:z-base-32]]-encoded key **MUST** be appended as the Origin of all records:

| Name                                                     | Type | TTL    | Rdata                                     |
| -------------------------------------------------------- | ---- | ------ | ----------------------------------------- |
| _did.o4dksfbqk85ogzdb5osziw6befigbuxmuxkuxq8434q89uj56uyy | TXT  |  7200  | vm=k0,k1,k2;auth=k0;asm=k1;inv=k2;del=k2;srv=s0,s1,s2 |

### Property Mapping

The following section describes how a [[ref:DID Document]] is mapped to a DNS packet. To avoid repeating potentially
long identifiers in resource name fields, resources are aliased with zero-indexed values (e.g. `k0`, `k1`, `s0`, `s1`).
The full identifier is then stored in the resource data field (e.g. `id=abcd,t=0,k=...`).

* The _root record_, `_did` or `_did.TLD` if a [TLD](https://en.wikipedia.org/wiki/Top-level_domain) is being utilized
contains a list of IDs of the keys and service endpoints used in different sections of the [[ref:DID Document]].

An example is as follows:

| Name      | Type | TTL  | Rdata                                                 |
| --------- | ---- | ---- | ----------------------------------------------------- |
| _did.TLD  | TXT  | 7200 | vm=k1,k2,k3;auth=k1;asm=k2;inv=k3;del=k3;srv=s1,s2,s3 |

The following instructions serve as a reference of mapping DID Document properties to [DNS TXT records](https://en.wikipedia.org/wiki/TXT_record):

#### Verification Methods

* Each Verification Method **name** is represented as a `_kN._did` record where `N` is the zero-indexed positional index of
a given [Verification Method](https://www.w3.org/TR/did-core/#verification-methods) (e.g. `_k0`, `_k1`).

* Each [Verification Method](https://www.w3.org/TR/did-core/#verification-methods) **rdata** is represented with the form
`id=M,t=N,k=O` where `M` is the key's ID, `N` is the index of the key's type from [key type index](#key-type-index),
and `O` is the base64URL [[spec:RFC4648]] representation of the public key.

##### Key Type Index

| Index | Key Type                                               |
| ----- | ------------------------------------------------------ |
| 0     | [Ed25519](https://ed25519.cr.yp.to/)                   |
| 1     | [secp256k1](https://en.bitcoin.it/wiki/Secp256k1)      |
| 2     | [secp256r1](https://neuromancer.sk/std/secg/secp256r1) |

An example [Verification Method]((https://www.w3.org/TR/did-core/#verification-methods)) record represented as a DNS TXT
record is as follows:

| Name     | Type | TTL  | Rdata                                                     |
| -------- | ---- | ---- | --------------------------------------------------------- |
| _k0._did | TXT  | 7200 | id=abcd,t=0,k=r96mnGNgWGOmjt6g_3_0nd4Kls5-kknrd4DdPW8qtfw |

#### Verification Relationships

* Each [Verification Relationship](https://www.w3.org/TR/did-core/#verification-relationships) is represented as a part
of the root `_did.TLD` record (see: [Property Mapping](#property-mapping)).

The following table maps Verification Relationship types to their record name:

##### Verification Relationship Index

| Verification Relationship  | Record Name |
| ------------------------- | ----------- |
| Authentication            | auth        |
| Assertion                 | asm         |
| Key Agreement             | agm         |
| Capability Invocation     | inv         |
| Capability Delegation     | del         |

The record data is uniform across [Verification Relationships](https://www.w3.org/TR/did-core/#verification-relationships),
represented as a comma separated list of key references.

An example is as follows:

| Verification Relationship            | Rdata in the Root Record                     |
|-------------------------------------|----------------------------------------------|
| "authentication": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"] | auth=0,HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ |
| "assertionMethod": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"]| asm=0,HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ  |
| "keyAgreement": ["#1"]              | agm=1                                        |
| "capabilityInvocation": ["#0"]      | inv=0                                        |
| "capabilityDelegation": ["#0"]      | del=0                                        |

#### Services

* Each [Service](https://www.w3.org/TR/did-core/#services)'s **name** is represented as a `_sN._did` record where `N` is
the zero-indexed positional index of the Service (e.g. `_s0`, `_s1`).
* Each [Service](https://www.w3.org/TR/did-core/#services)'s **data** is represented with the form `id=M,t=N,uri=O`
where `M` is the Service's ID, `N` is the Service's Type and `O` is the Service's URI.

An example is given as follows:

| Name     | Type | TTL  | Rdata                                                     |
| -------- | ---- | ---- | --------------------------------------------------------- |
| _s0._did | TXT  | 7200 | id=dwn,t=DecentralizedWebNode,uri=https://example.com/dwn |

Each Service is also represented as part of the root `_did.TLD` record as a list under the key `srv=<ids>` where `ids`
is a comma separate list of all IDs for each Service.

#### Example

A sample transformation is provided of a fully-featured DID Document to a DNS packet:

**DID Document**

```json
{
  "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
  "verificationMethod": [
    {
      "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0",
      "type": "JsonWebKey2020",
      "controller": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
      "publicKeyJwk": {
        "kid": "0",
        "alg": "EdDSA",
        "crv": "Ed25519",
        "kty": "OKP",
        "x": "r96mnGNgWGOmjt6g_3_0nd4Kls5-kknrd4DdPW8qtfw"
      }
    },
    {
      "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ",
      "type": "JsonWebKey2020",
      "controller": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
      "publicKeyJwk": {
        "kid": "HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ",
        "alg": "ES256K",
        "crv": "secp256k1",
        "kty": "EC",
        "x": "KI0DPvL5cGvznc8EDOAA5T9zQfLDQZvr0ev2NMLcxDw",
        "y": "0iSbXxZo0jIFLtW8vVnoWd8tEzUV2o22BVc_IjVTIt8"
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

**DNS Resource Records**

| Name     | Type | TTL   | Rdata                                                                        |
| -------- | ---- | ----- | ---------------------------------------------------------------------------- |
| _did.TLD | TXT  | 7200  | vm=k0,k1;auth=k0,k1;asm=k0,k1;inv=k0;del=k0;srv=s1                           |
| _k0._did | TXT  | 7200  | id=0,t=0,h=afdea69c63605863a68edea0ff7ff49dde0a96ce7e9249eb7780dd3d6f2ab5fc  |
| _k1._did | TXT  | 7200  | id=HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ,t=1,k=BCiNAz7y-XBr853PBAzgAOU_c0Hyw0Gb69Hr9jTC3MQ80iSbXxZo0jIFLtW8vVnoWd8tEzUV2o22BVc_IjVTIt8 |
| _s0._did | TXT  | 7200  | id=dwn,t=DecentralizedWebNode,uri=https://example.com/dwn                    |

### Operations

Entries to the [[ref:DHT]] require a signed record. As such, the keypair that is generated for the [[ref:Pkarr]]
identifier is used to sign the [[ref:DHT]] record. This keypair **MUST** always be present in a `did:dht` document
and is referred to as the [[ref:Identity Key]].

#### Create

To create a `did:dht`, the process is as follows:

1. Generate an [[ref:Ed25519]] keypair and encode the public key using the format provided in the [format section](#format).

2. Construct a compliant JSON representation of a [[ref:DID Document]].

    a. The document **MUST** include a [Verification Method](https://www.w3.org/TR/did-core/#verification-methods) with
 the _identifier key_ encoded as a `publicKeyJwk` as per [[spec:RFC7517]] with an `id` of ``#0`` and `type` of
 `JsonWebKey2020` as per [[ref:VC-JWS-2020]].

    b. The document can include any number of other [core properties](https://www.w3.org/TR/did-core/#core-properties);
 always representing key material as a `JWK` as per [[spec:RFC7517]].

3. Map the output [[ref:DID Document]] to a DNS packet as outlined in [property mapping](#property-mapping).

4. Construct a signed [[ref:BEP44]] put message with the `v` value as a [[ref:bencode]]d DNS packet from the prior step.

5. Submit the result of to the [[ref:DHT]] via a [[ref:Pkarr]] relay, or a [[ref:Gateway]].
 
#### Read

To read a `did:dht`, the process is as follows:

1. Take the suffix of the DID, that is, the _encoded identifier key_, and pass it to a [[ref:Pkarr]] relay or a [[ref:Gateway]].
2. Decode the resulting [[ref:BEP44]] response's `v` value using [[ref:bencode]].
3. Reverse the DNS [property mapping](#property-mapping) process and re-construct a compliant [[ref:DID Document]].

#### Update

Any valid [[ref:BEP44]] record written to the DHT is considered an update. This means as long as control of the
[[ref:Identity Key]] is retained any update is made possibly by signing and writing records with a unique incremental
sequence number with [mutable items](https://www.bittorrent.org/beps/bep_0044.html).

It is **RECOMMENDED** that updates are made infrequently as caching of the DHT is highly encouraged.

#### Deactivate

To deactivate a document there are a couple options:
    
1. Let the DHT record expire and cease to publish it.
2. Publish a new DHT record where the `rdata` of the root DNS record is the string "deactivated."

| Name      | Type | TTL  | Rdata       |
| --------- | ---- | ---- | ----------- |
| _did.TLD  | TXT  | 7200 | deactivated |

::: note
If you have published your DID through a [[ref:Gateway]] you may need to contact the operator to have them remove the
record from their [[ref:Retained DID Set]].
:::

### Type Indexing

Type indexing is an **OPTIONAL** feature that enables DIDs to flag themselves as being of a certain type. Types are not
included as a part of the DID Document, but rather included as part of the DNS packet. This allows for DIDs to be
indexed by type by [[ref:Gateway]]s, and for DIDs to be resolved by type.

DIDs can be indexed by type by adding a `_typ._did` record to the DNS packet. A DID may have **AT MOST** one type index
record. This record is a TXT record with the following format:

* The record **name** represented as a `_typ._did`.
* The record **data** is represented with the form `id=0,1,2` where the value is a comma separated list of types from
the [type index](#type-index).

An example type record is as follows:

| Name      | Type | TTL  | Rdata       |
| --------- | ---- | ---- | ----------- |
| _typ._did | TXT  | 7200 | id=0,1,2    |

#### Type Index

| Type Name               | Schema                                    | Type Integer |
|-------------------------|-------------------------------------------| ------------ |
| Organization            | https://schema.org/Organization           | 1            |
| Government Organization | https://schema.org/GovernmentOrganization | 2            |
| Corporation             | https://schema.org/Corporation            | 3            |
| Local Business          | https://schema.org/LocalBusiness          | 4            |
| Software Package        | https://schema.org/SoftwareSourceCode     | 5            |
| Web App                 | https://schema.org/WebApplication         | 6            |
| Financial Institution   | https://schema.org/FinancialService       | 7            |

## Bitcoin-anchored Gateways

::: issue
[](https://github.com/TBD54566975/did-dht-method/issues/10)
Fully define the BTC interactions.
:::

As an **OPTIONAL** feature of the DID DHT Method, nodes may choose to act as a gateway to the DID DHT for a fee. This
fee is paid in Bitcoin, and the gateway operator must anchor a transaction on Bitcoin that [[ref:timelock]]s
Bitcoin value proportional to the number of DIDs they introduce into the [[ref:Retained DID Set]]. Notably, this is not a fee
for the creation of a DID, but rather a fee for the retention of a DID paid by the node operator. DID controllers "pay"
for retention by providing a [retention proof](#generating-a-retention-proof).

The amount of value locked must be no less than the mean value of the upper half of [UTXOs](https://en.wikipedia.org/wiki/Unspent_transaction_output)
for the block in which the timelock takes effect, and the lock must be a *relative timelock* set to **1000 blocks**.

### Retained DID Set

#### Generating a Retention Proof

Perform Proof of Work over the DID's identifier + the `retention` value of a given DID operation (composed of the
selected bitcoin block hash and nonce). The resulting Retention Proof Hash determines the duration of retention based
 on the number of leading zeros of the hash, which must be no less than 26.

#### Managing the Retained DID Set

Nodes following the Retention Set rules SHOULD sort DIDs they are retaining by age of retention proof, followed by
number of retention proof leading 0s. When a node needs to reduce its retained set of DID entries, it SHOULD remove
entries from the bottom of the list in accordance with this sort.

#### Reporting on Retention Status

Nodes MUST include the approximate time until retention fall-off in the Method-specific metadata of a resolved DID
Document, to aid in Identity Agents (wallets) being able to assess whether resubmission is required.

### Gateway API

At a minimum, a gateway **MUST** support the
[Relay API defined by Pkarr](https://github.com/Nuhvi/pkarr/blob/main/design/relays.md).

Expanding on this API, a Gateway **MUST** support the following API endpoints:

#### Register or Update a DID

- **Method:** `PUT`
- **Path:** `/did`
- **Request Body:** A JSON payload constructed as follows...
    - `did` - **string** - The DID to register.
    - `sig` - **string** - A base64URL-encoded signature of the [[ref:BEP44]] payload
    - `seq` - **integer** - A sequence number for the DID. This number **MUST** be unique for each DID operation,
    recommended to be a unix timestamp.
    - `v` - **string** - A base64URL-encoded bencoded DNS packet containing the DID Document.
- **Returns**:
    - `200` - Success.
    - `400` - Invalid request body.
    - `401` - Invalid signature.
    - `409` - DID already exists with a higher sequence number.

```json
{
    "did": "did:dht:example",
    "sig": "<base64URL-encoded-signature>",
    "seq": 1234,
    "v": "<base64URL-encoded bencoded DNS packet>"
}
```

Upon receiving a request to register a DID, the Gateway **MUST** verify the signature of the request, and if valid,
publish the DID Document to the DHT. If the DNS Packets contains a `_typ._did` record, the Gateway **MUST** index the
DID by its type.

#### Resolving a DID

- **Method:** `GET`
- **Path:** `/did/:id`
- **Returns**:
    - `200` - Success.
        - `did` - **object** - A JSON object representing the DID's Document.
        - `types` - **array** - An array of type strings for the DID.
    - `404` - DID not found.

```json
{
    "did": {
        "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
        "verificationMethod": [
            {
                "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0",
                "type": "JsonWebKey2020",
                "controller": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
                "publicKeyJwk": {
                    "kid": "0",
                    "alg": "EdDSA",
                    "crv": "Ed25519",
                    "kty": "OKP",
                    "x": "r96mnGNgWGOmjt6g_3_0nd4Kls5-kknrd4DdPW8qtfw"
                }
            }
        ],
        "authentication": ["#0"],
        "assertionMethod": ["#0"]
    },
    "types": [1, 4]
}
```

Upon receiving a request to resolve a DID, the Gateway **MUST** query the DHT for the DID Document, and if found,
return the DID Document. If the records are not found in the DHT, the Gateway **MAY** fall back to its local storage.
If the DNS Packets contains a `_typ._did` record, the Gateway **MUST** return the type index.

::: note
This API is not required to return the full DNS packet, but rather the DID Document and type index. If the full DNS
packet, with its signature data is wanted, it is recommended to use the
[Relay API](https://github.com/Nuhvi/pkarr/blob/main/design/relays.md) directly.
:::

#### Deactivating a DID

To intentionally deactivate a DID, as opposed to letting the record cease being republished to the DHT, a DID controller
follows the same process as [updating a DID](#register-or-update-a-did), but with a record format outlined in the
[section on deactivation](#deactivate).

Upon receiving a request to deactivate a DID, the Gateway **MUST** verify the signature of the request, and if valid,
stop republishing the DHT. If the DNS Packets contains a `_typ._did` record, the Gateway **MUST** remove the type index.

#### Type Indexing

- **Method:** `GET`
- **Path:** `/did/types?id=:id`
    - `id` - **string** - The type to query from the index.
- **Returns**:
    - `200` - Success.
        - `dids` - **array** - An array of DIDs matching the associated type.
    - `404` - Type not found.

```json
{
    "dids": [
        "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
        "did:dht:uodqi99wuzxsz6yx445zxkp8ddwj9q54ocbcg8yifsqru45x63kj"
    ]
}
```

A query to the type index returns an array of DIDs matching the associated type. If the type is not found, a `404` is
returned. If there are no DIDs matching the type, an empty array is returned.

## Implementation Considerations

::: issue
[](https://github.com/TBD54566975/did-dht-method/issues/27)

Write this section.
:::

Data needs to be republished.

## Security and Privacy Considerations

::: issue
[](https://github.com/TBD54566975/did-dht-method/issues/28)

Write this section.
:::

### Security

### Privacy

## References

[[def:Ed25519]]
~ [Ed25519](https://ed25519.cr.yp.to/). D. J. Bernstein, N. Duif, T. Lange, P. Schwabe, B.-Y. Yang; 26 September 2011.
[ed25519.cr.yp.to](https://ed25519.cr.yp.to/).

[[def:BEP44]]
~ [BEP44](https://www.bittorrent.org/beps/bep_0044.html). Storing arbitrary data in the DHT. A. Norberg, S. Siloti;
19 December 2014. [Bittorrent.org](https://www.bittorrent.org/).

[[def:Bencode]]
~ [Bencode](https://wiki.theory.org/BitTorrentSpecification#Bencoding). A way to specify and organize data in a terse
format. [Bittorrent.org](https://www.bittorrent.org/).

[[def:z-base-32]]
~ [z-base-32](https://philzimmermann.com/docs/human-oriented-base-32-encoding.txt). Human-oriented base-32 encoding.
Z. O'Whielacronx; November 2002.

[[def:VC-JWS-2020]]
~ [Verifiable Credentials JSON Web Signature Suite 2020](https://www.w3.org/TR/vc-jws-2020/). O. Steele, M. Jones; 29
June 2023. [W3C](https://www.w3.org/).

[[def:Timelock]]
~ [Timelock](https://github.com/bitcoin/bips/blob/master/bip-0065.mediawiki). P. Todd. 01 October 2014.
[Bitcoin](https://github.com/bitcoin).

[[spec]]