The DID DHT Method Specification 1.0
==================

**Specification Status**: Working Draft

**Latest Draft:** [https://did-dht.com](https://did-dht.com)

**Registry:** [https://did-dht.com/registry](https://did-dht.com/registry)

**Draft Created:** October 20, 2023

**Latest Update:** January 9, 2024

**Editors:**
~ [Gabe Cohen](https://github.com/decentralgabe)
~ [Daniel Buchner](https://github.com/csuwildcat)

**Contributors:**
~ [Moe Jangda](https://github.com/mistermoe)
~ [Frank Hinek](https://github.com/frankhinek)

**Participate:**
~ [GitHub repo](https://github.com/TBD54566975/did-dht-method)
~ [File a bug](https://github.com/TBD54566975/did-dht-method/issues)
~ [Commit history](https://github.com/TBD54566975/did-dht-method/commits/main)

## Abstract

A [DID Method](https://www.w3.org/TR/did-core/) based on [Pkarr](https://github.com/nuhvi/pkarr), identified by `did:dht`.

<style id="protocol-stack-styles">
  #protocol-stack-styles + table {
    display: table;
    width: 400px;
    border-radius: 4px;
    box-shadow: 0 1px 3px -1px rgb(0 0 0 / 80%);
    overflow: hidden;
  }
  #protocol-stack-styles + table tr, #protocol-stack-styles + table td {
    border: none;
  }
  #protocol-stack-styles + table tr {
    text-shadow: 0 1px 2px rgb(255 255 255 / 75%);
  }
  #protocol-stack-styles + table tr:nth-child(1) {
    background: hsl(149deg 100% 86%);
  }
  #protocol-stack-styles + table tr:nth-child(2) {
    background: hsl(0deg 100% 82%);
  }
  #protocol-stack-styles + table tr:nth-child(3) {
    background: hsl(198deg 100% 87%);
  }
</style>

:----: |
DID DHT |
[Pkarr](https://github.com/nuhvi/pkarr)|
[Mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT) |

[[ref:Pkarr]] stands for **P**ublic **K**ey **A**ddressable **R**esource **R**ecords. [[ref:Pkarr]] makes use of the
[Mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT), specifically [[ref:BEP44]] to store records. This DID
method uses [[ref:Pkarr]] to store _DID Documents_.

As [[ref:Pkarr]] states, mainline is used for the following reasons:

1. It has a proven track record of 15 years.
2. It is the biggest DHT in existence with an estimated 10 million nodes.
3. It is fairly generous in retaining data.
4. It has been implemented in most languages and is stable.

The syntax of the identifier and accompanying data model used by the protocol is conformant with the [[spec:DID-Core]]
specification and shall be registered with the [[spec:DID-Spec-Registries]].

## Conformance

The keywords MAY, MUST, MUST NOT, RECOMMENDED, SHOULD, and SHOULD NOT in this document are to be interpreted as described in [BCP 14](https://www.rfc-editor.org/info/bcp14) [[spec:RFC2119]] [[spec:RFC8174]] when, and only when, they appear in all capitals, as shown here.

## Terminology

[[def:Decentralized Identifier, Decentralized Identifier, DID, DIDs, DID Document, DID Documents]]
~ A [W3C specification](https://www.w3.org/TR/did-core/) describing an _identifier that enables verifiable, decentralized digital identity_. Associated with a document containing properties outlined in the specification.

[[def:DID Suffix, Suffix]]
~ The unique identifier string within a DID URI. e.g. The unique suffix of `did:dht:123` would be `123`.

[[def:DID Suffix Data, Suffix Data]]
~ Data required to deterministically generate a DID, the [[ref:Identity Key]].

[[def:Identity Key]]
~ An [[ref:Ed25519]] public key encoded with [[ref:z-base-32]] used to uniquely identify a `did:dht` document.

[[def:DID DHT Service]]
~ A service that provides a [[ref:DHT]] interface to the [[ref:Pkarr]] network, extended to support this [[ref:DID]] method.

[[def:Pkarr]]
~ An [open-source project](https://github.com/nuhvi/pkarr) that provides the "simplest possible streamlined integration
between the Domain Name System and peer-to-peer overlay networks, enabling self-issued public keys to function as
sovereign, publicly addressable domains."

[[def:Mainline DHT, DHT, Mainline, Mainline Node]]
~ [Mainline DHT](https://en.wikipedia.org/wiki/Mainline_DHT) is the name given to the DHT used by the BitTorrent
protocol. It is a distributed system for storing and finding data on a peer-to-peer network. It is based on
[Kademlia](https://en.wikipedia.org/wiki/Kademlia) and is primarily used to store and retrieve _torrent_ metadata.
It has between 16 and 28 million concurrent users.

[[def:Gateway, Gateways, Nodes, DID DHT Node, Bitcoin-anchored Gateway]]
~ A node that acts as a gateway to the DID DHT. The gateway may offer a set of APIs to interact with the DID DHT,
such as features providing guaranteed retention, historical resolution, and other features.

[[def:Registered Gateway, Registered Gateways]]
~ A gateway that has chosen to make itself discoverable via a [[ref:Gateway Registry]] such as the [DID DHT Registry](registry/index.html#gateways).

[[def:Gateway Registry, Gateway Registries]]
~ A system used to make [[ref:Gateways]], more specifically, [[ref:Registered Gateways]] discoverable.

[[def:Client, Clients]]
~ A client is a piece of software that is responsible for generating a `did:dht` and submitting it to a [[ref:Mainline]] node or [[ref:Gateway]].

[[def:Retained DID Set, Retained Set, Retention Set]]
~ The set of DIDs that a [[ref:Gateway]] is retaining and thus is responsible for republishing.

[[def:Retention Proof, Retention Proofs]]
~ A proof of work that is performed by the [[ref:DID]] controller to prove that they are still in control of the DID.
Nodes use this proof to determine how long they should retain a DID.

## DID DHT Method Specification

### Format

The format for `did:dht` conforms to the [[spec:DID Core]] specification. It consists of the
`did:dht` prefix followed by the [[ref:Pkarr]] identifier. The [[ref:Pkarr]] identifier is a [[ref:z-base-32]].
encoded [[ref:Ed25519]] public key.

```text
did-dht-format := did:dht:<z-base-32-value>
z-base-32-value := [a-km-uw-z13-9]+
```

Alternatively, one can interpret the encoding rules as a series of transformation functions on the raw public key bytes:

```text
did-dht-format := did:dht:Z-BASE-32(raw-public-key-bytes)
```

### DIDs as DNS Records

In this scheme, we encode the [[ref:DID Document]] as multiple [DNS TXT records](https://en.wikipedia.org/wiki/TXT_record).
Comprising a DNS packet [[spec:RFC1034]] [[spec:RFC1035]], which is then stored in the [[ref:DHT]].

| Name      | Type | TTL    | Rdata                                    |
| --------- | ---- | ------ | ---------------------------------------- |
| _did.     | TXT  |  7200  | vm=k0,k1,k2;auth=k0;asm=k1;inv=k2;del=k2;srv=s0,s1,s2 |
| _k0._did. | TXT  |  7200  | id=0;t=0;k=`<unpadded-b64url>`           |
| _k1._did. | TXT  |  7200  | id=1;t=1;k=`<unpadded-b64url>`           |
| _k2._did. | TXT  |  7200  | id=2;t=1;k=`<unpadded-b64url>`           |
| _s0._did. | TXT  |  7200  | id=domain;t=LinkedDomains;se=foo.com     |
| _s1._did. | TXT  |  7200  | id=dwn;t=DecentralizedWebNode;se=https://dwn.tbddev.org/dwn5 |

::: note
The recommended TTL value is 7200 seconds (2 hours), the default TTL for Mainline records.
:::

- A root `_did.` record identifies the [property mapping](#property-mapping) for the document.

- Additional records like `srv` (services), `vm` (verification methods), and verification relationships
(e.g., authentication, assertion, etc.) are represented as additional records in the format `<ID>._did.`.
These records contain the zero-indexed value of each `key` or `service` as attributes.

- All records ****MUST**** end in `_did.` or `_did.TLD.` if a TLD is associated with the record.

::: note
It might look like repeating `_did` is an overhead, but is compressed away using
[DNS packet compression](https://courses.cs.duke.edu/fall16/compsci356/DNS/DNS-primer.pdf) techniques.
:::

- The DNS packet ****MUST**** set the _Authoritative Answer_ flag since this is always an _Authoritative_ packet.

- The DID identifier [[ref:z-base-32]]-encoded key ****MUST**** be appended as the Origin of all records:

| Name                                                       | Type | TTL    | Rdata                                                 |
| ---------------------------------------------------------- | ---- | ------ | ----------------------------------------------------- |
| _did.o4dksfbqk85ogzdb5osziw6befigbuxmuxkuxq8434q89uj56uyy. | TXT  |  7200  | vm=k0,k1,k2;auth=k0;asm=k1;inv=k2;del=k2;srv=s0,s1,s2 |

### Property Mapping

The following section describes mapping a [[ref:DID Document]] to a DNS packet. To avoid repeating potentially
long identifiers in resource name fields, resource names are aliased with zero-indexed values (e.g. `k0`, `k1`, `s0`, `s1`).
The complete identifier is stored in the resource data field (e.g. `id=abcd;t=0;k=...`).

- The _root record_, `_did.` or `_did.TLD.` if a [TLD](https://en.wikipedia.org/wiki/Top-level_domain) is used,
contains a list of IDs of the keys and service endpoints used in different sections of the [[ref:DID Document]].

- Verification Methods, Verification Relationships, and Services are separated by a semicolon (`;`), while
values within each property are separated by a comma (`,`).

- Across all properties, distinct elements are separated by semicolons (`;`) while array elements are separated by
commas (`,`).

An example of a _root record_ is as follows:

| Name       | Type | TTL  | Rdata                                                 |
| ---------- | ---- | ---- | ----------------------------------------------------- |
| _did.TLD.  | TXT  | 7200 | vm=k1,k2,k3;auth=k1;asm=k2;inv=k3;del=k3;srv=s1,s2,s3 |

The following instructions serve as a reference for mapping DID Document properties to [DNS TXT records](https://en.wikipedia.org/wiki/TXT_record):

#### Identifiers

##### Controller

A [DID controller](https://www.w3.org/TR/did-core/#did-controller) ****MAY**** be present in a `did:dht` document.

If present, a DID controller ****MUST**** be represented as a `_cnt._did` record in the form of a comma-separated
list of controller DID identifiers.

An example is given as follows:

| Name       | Type | TTL  | Rdata            |
| ---------- | ---- | ---- | ---------------- |
| _cnt._did. | TXT  | 7200 | did:example:abcd |

##### Also Known As

A `did:dht` document ****MAY**** have multiple identifiers using the [alsoKnownAs](https://www.w3.org/TR/did-core/#also-known-as) property.

If present, alternate DID identifiers ****MUST**** be represented as `_aka_.did` record in the form of a
comma-separated list of DID identifiers.

An example is given as follows:

| Name       | Type | TTL  | Rdata                              |
| ---------- | ---- | ---- | ---------------------------------- |
| _aka._did. | TXT  | 7200 | did:example:efgh,did:example:ijkl  |

#### Verification Methods

- Each Verification Method **name** is represented as a `_kN._did` record where `N` is the zero-indexed positional index of
a given [Verification Method](https://www.w3.org/TR/did-core/#verification-methods) (e.g. `_k0`, `_k1`).

- Each [Verification Method](https://www.w3.org/TR/did-core/#verification-methods) **rdata** is represented with the form
`id=M;t=N;k=O` where `M` is the key's ID, `N` is the index of the key's type from [key type index](registry/index.html#key-type-index),
and `O` is the unpadded base64URL [[spec:RFC4648]] representation of the public key.

#### Verification Relationships

- Each [Verification Relationship](https://www.w3.org/TR/did-core/#verification-relationships) is represented as a part
of the root `_did.TLD.` record (see: [Property Mapping](#property-mapping)).

The following table acts as a map between Verification Relationship types and their record name:

##### Verification Relationship Index

| Verification Relationship  | Record Name  |
| -------------------------- | ------------ |
| Authentication             | `auth`       |
| Assertion                  | `asm`        |
| Key Agreement              | `agm`        |
| Capability Invocation      | `inv`        |
| Capability Delegation      | `del`        |

The record data is uniform across [Verification Relationships](https://www.w3.org/TR/did-core/#verification-relationships),
represented as a comma-separated list of key references.

An example is as follows:

| Verification Relationship            | Rdata in the Root Record                    |
|-------------------------------------|----------------------------------------------|
| "authentication": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"] | auth=0,HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ |
| "assertionMethod": ["#0", "#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"]| asm=0,HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ  |
| "keyAgreement": ["#1"]              | agm=1                                        |
| "capabilityInvocation": ["#0"]      | inv=0                                        |
| "capabilityDelegation": ["#0"]      | del=0                                        |

#### Services

- Each [Service](https://www.w3.org/TR/did-core/#services)'s **name** is represented as a `_sN._did.` record where `N` is
the zero-indexed positional index of the Service (e.g. `_s0`, `_s1`).
- Each [Service](https://www.w3.org/TR/did-core/#services)'s **data** is represented with the form `id=M;t=N;se=O`
where `M` is the Service's ID, `N` is the Service's Type and `O` is the Service's URI.
  - Multiple service endpoints can be represented as an array (e.g. `id=dwn;t=DecentralizedWebNodes;se=https://dwn.org/dwn1,https://dwn.org/dwn2`)
  - Additional properties ****MAY**** be present (e.g. `id=dwn;t=DecentralizedWebNode;se=https://dwn.org/dwn1;sig=1;enc=2`)

An example is given as follows:

| Name      | Type | TTL  | Rdata                                                    |
| --------- | ---- | ---- | -------------------------------------------------------- |
| _s0._did. | TXT  | 7200 | id=dwn;t=DecentralizedWebNode;se=https://example.com/dwn |

Each Service is represented as part of the root `_did.TLD.` record as a list under the key `srv=<ids>` where `ids`
is a comma-separated list of all IDs for each Service.

#### Example

A sample transformation of a fully-featured DID Document to a DNS packet is exemplified as follows:

**DID Document**

```json
{
  "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
  "controller": "did:example:abcd",
  "alsoKnownAs": ["did:example:efgh", "did:example:ijkl"],
  "verificationMethod": [
    {
      "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0",
      "type": "JsonWebKey",
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
      "type": "JsonWebKey",
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
  "authentication": [
    "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0",
    "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"
  ],
  "assertionMethod": [
    "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0",
    "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ"
  ],
  "capabilityInvocation": [
    "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0"
  ],
  "capabilityDelegation": [
    "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0"
  ],
  "service": [
    {
      "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#dwn",
      "type": "DecentralizedWebNode",
      "serviceEndpoint": ["https://example.com/dwn1", "https://example/dwn2"]
    }
  ]
}
```

**DNS Resource Records**

| Name       | Type | TTL   | Rdata                                                                       |
| ---------- | ---- | ----- | --------------------------------------------------------------------------- |
| _did.TLD.  | TXT  | 7200  | vm=k0,k1;auth=k0,k1;asm=k0,k1;inv=k0;del=k0;srv=s1                          |
| _cnt.did.  | TXT  | 7200  | did:example:abcd                                                            |
| _aka.did.  | TXT  | 7200  | did:example:efgh,did:example:ijkl                                           |
| _k0._did.  | TXT  | 7200  | id=0;t=0;h=afdea69c63605863a68edea0ff7ff49dde0a96ce7e9249eb7780dd3d6f2ab5fc |
| _k1._did.  | TXT  | 7200  | id=HTsY9aMkoDomPBhGcUxSOGP40F-W4Q9XCJV1ab8anTQ;t=1;k=BCiNAz7y-XBr853PBAzgAOU_c0Hyw0Gb69Hr9jTC3MQ80iSbXxZo0jIFLtW8vVnoWd8tEzUV2o22BVc_IjVTIt8 |
| _s0._did.  | TXT  | 7200  | id=dwn;t=DecentralizedWebNode;se=https://example.com/dwn1,https://example.com/dwn2 |

### Operations

Entries to the [[ref:DHT]] require a signed record. As such, the keypair used for the [[ref:Pkarr]]
identifier is also used to sign the [[ref:DHT]] record. This keypair ****MUST**** always be present in a `did:dht` document
and is referred to as the [[ref:Identity Key]].

#### Create

To create a `did:dht`, the process is as follows:

1. Generate an [[ref:Ed25519]] keypair and encode the public key using the format provided in the [format section](#format).

2. Construct a conformant JSON representation of a [[ref:DID Document]].

    a. The document ****MUST**** include a [Verification Method](https://www.w3.org/TR/did-core/#verification-methods) with
    the _identifier key_ encoded as a `publicKeyJwk` as per [[spec:RFC7517]] with an `id` of `#0` and `type` of
    `JsonWebKey` as per [[ref:VC-JOSE-COSE]].

    b. The document can include any number of other [core properties](https://www.w3.org/TR/did-core/#core-properties);
    always representing key material as a `JWK` as per [[spec:RFC7517]].

3. Map the output [[ref:DID Document]] to a DNS packet as outlined in [property mapping](#property-mapping).

4. Construct a signed [[ref:BEP44]] put message with the `v` value as a [[ref:bencode]]d DNS packet from the prior step.

5. Submit the result of to the [[ref:DHT]] via a [[ref:Pkarr]] relay, or a [[ref:Gateway]], with the identifier created in step 1.

::: note
This specification **does not** make use of JSON-LD. As such it is prohibited to inclue an `@context` property in the DID Document.
:::

#### Read

To read a `did:dht`, the process is as follows:

1. Take the suffix of the DID, that is, the _[[ref:z-base-32]] encoded identifier key_, and pass it to a [[ref:Pkarr]]
relay or a [[ref:Gateway]].
2. Decode the resulting [[ref:BEP44]] response's `v` value using [[ref:bencode]].
3. Reverse the DNS [property mapping](#property-mapping) process and re-construct a conformant [[ref:DID Document]].

::: note
As a fallback, if a `did:dht` value cannot be resolved via the network, it can be expanded to a conformant [[ref:DID Document]]
containing just the [[ref:Identity Key]].
:::

#### Update

Any valid [[ref:BEP44]] record written to the DHT is an update. As long as control of the
[[ref:Identity Key]] is retained any update is made possibly by signing and writing records with a unique incremental
sequence number with [mutable items](https://www.bittorrent.org/beps/bep_0044.html).

It is ****RECOMMENDED**** that updates are infrequent, as caching of the DHT is highly encouraged.

#### Deactivate

To deactivate a document, there are a couple options:

1. Let the DHT record expire and cease to publish it.
2. Publish a new DHT record where the `rdata` of the root DNS record is the string "deactivated."

| Name       | Type | TTL  | Rdata       |
| ---------- | ---- | ---- | ----------- |
| _did.TLD.  | TXT  | 7200 | deactivated |

::: note
If you have published your DID through a [[ref:Gateway]], you may need to contact the operator to have them remove the
record from their [[ref:Retained DID Set]].
:::

### Type Indexing

Type indexing is an **OPTIONAL** feature that enables DIDs to become **discoverable**, by flagging themselves as being of
a particular type. Types are not included as a part of the DID Document, but rather as part of the DNS packet. This allows
for DIDs to be indexed by type by [[ref:Gateways]], and for DIDs to be resolved by type.

DIDs can be indexed by type by adding a `_typ._did.` record to the DNS packet. A DID may have **AT MOST** one type index
record. This record is a TXT record with the following format:

- The record **name** is represented as a `_typ._did.` record.
- The record **data** is represented with the form `id=0,1,2` where the value is a comma-separated list of types from
the [type index](#type-indexing).

An example type record is as follows:

| Name       | Type | TTL  | Rdata     |
| ---------- | ---- | ---- | --------- |
| _typ._did. | TXT  | 7200 | id=0,1,2  |

Types can be found and registered in the [DID DHT Registry](registry/index.html#indexed-types).

::: note
Identifying entities through type-based indexing is a preliminary and relatively unreliable method. It serves merely
as an initial phase in recognizing the identity linked to a [[ref:DID]]. To validate identity claims more robustly,
it is essential to delve deeper, employing tools like verifiable credentials and examining related data.
:::

## Interoperability With Other DID Methods

As an **OPTIONAL** extension, some existing DID methods can leverage `did:dht` to expand their feature set. This
enhancement is most useful for DID methods that operate based on a single key and are compatible with the [[ref:Ed25519]]
key format. By adopting this optional extension, users can maintain their current DIDs without any changes. Additionally,
they gain the ability to add extra information to their DIDs. This is achieved by either publishing or retrieving
data from [[ref:Mainline]].

Interoperable DID methods ****MUST**** be registered in [the corresponding registry](registry/index.html#interoperable-did-methods).

## Gateways

Gateways serve as specialized nodes within the network, providing a range of DID-centric functionalities that extend
beyond the capabilities of a standard [[ref:Mainline DHT]] node. This section elaborates on these unique features,
outlines the operational prerequisites for managing a gateway, and discusses various other facets, including the
optional integration of these gateways into a registry system.

::: note
[[ref:Gateways]] may choose to support interoperable methods in addition to `did:dht` as outlined in the
[section on interoperability](#interoperability-with-other-did-methods).
:::

### Discovering Gateways

As an **OPTIONAL** feature of the DID DHT Method, operators of a [[ref:Gateway]] have the opportunity to make it to
a [[ref:Registered Gateway]]. A [[ref:Registered Gateway]] distinguishes itself by being discoverable through a
[[ref:Gateway Registry]]. This feature allows for easy location through various internet-based discovery mechanisms.
The primary purpose of [[ref:Registered Gateways]] is to simplify the process of finding [[ref:Gateways]], accessible
to any entity utilizing a [[ref:Gateway Registry]] to locate registered [[ref:Nodes]]. The [[ref:Gateway Registries]]
can vary in nature, encompassing a spectrum from centrally managed directories to diverse decentralized systems
including databases, ledgers, or other structures. [[ref:Registered Gateways]] are exposed through the
[Gateway Registry](registry/index.html#gateways).

### Retained DID Set

A [[ref:Retained DID Set]] refers to the set of DIDs a [[ref:Gateway]] retains and republishes to the DHT. A
[[ref:Gateway]] may choose to surface additional [APIs](#gateway-api) based on this set, such as providing a
[type index](#type-indexing).

To safeguard equitable access to the resources of [[ref:Gateways]], which are publicly accessible and potentially
subject to [a high volume of requests](#rate-limiting), we suggest an ****OPTIONAL**** mechanism aimed at upholding
fairness in the retention and republishing of record sets by [[ref:Gateways]]. This mechanism, referred to as a
[[ref:Retention Proof]], requires clients to generate a proof value for write requests. This process guarantees that
the amount of work done by a client is proportional to the duration of data retention and republishing a [[ref:Gateway]]
performs. This mechanism enhances the overall reliability and effectiveness of [[ref:Gateways]] in managing requests.

#### Generating a Retention Proof

A [[ref:Retention Proof]] is a form of [Proof of Work](https://en.bitcoin.it/wiki/Proof_of_work) performed over a DID
identifier concatenated with the `retention` value of a given DID operation. The `retention` value is composed of a
hash value specified [in the gateway registry](registry/index.html#gateways), and a random
[nonce](https://en.wikipedia.org/wiki/Cryptographic_nonce) using the [SHA-256 hashing algorithm](https://en.wikipedia.org/wiki/SHA-2).
The resulting _Retention Proof Hash_ is used to determine the retention duration based on the number of leading zeros
of the hash, referred to as the _difficulty_, which ****MUST**** be no less than 26 bits of the 256-bit hash value.
The algorithm, in detail, is as follows:

1. Obtain a did identifier and set it to `DID`.
2. Get the difficulty and recent hash from the server set to `DIFFICULTY` and `HASH`, respectively.
3. Generate a random 32-bit integer nonce value set to `NONCE`.
4. Compute the [SHA-256](https://en.wikipedia.org/wiki/SHA-2) hash over `ATTEMPT` where `ATTEMPT` = (`DID` + `HASH` + `NONCE`).
5. Inspect the result of `ATTEMPT`, and ensure it has >= `DIFFICULTY` bits of leading zeroes.
  a. If so, `ATTEMPT` = `RETENTION_PROOF`.
  b. Otherwise, regenerate `NONCE` and go to step 3.
6. Submit the `RETENTION_PROOF` to the [Gateway API](#register=or-update-a-did).

#### Managing the Retained DID Set

[[ref:Nodes]] following the [[ref:Retention Set]] rules ****SHOULD**** sort DIDs they are retaining by the number of
_leading 0s_ in their [[ref:Retention Proofs]] in descending order, followed by the block hash's index number in
descending order. When a [[ref:node]] needs to reduce its [[ref:retained set]] of DID entries, it ****SHOULD****
remove entries from the bottom of the list following this sort.

#### Reporting on Retention Status

Nodes ****MUST**** include the approximate time until retention fall-off in the
[DID Resolution Metadata](https://www.w3.org/TR/did-core/#did-resolution-metadata) of a resolved
[[ref:DID Document]], to aid [[ref:clients]] in being able to assess whether resubmission is required.

### Gateway API

At a minimum, a gateway ****MUST**** support the [Relay API defined by Pkarr](https://github.com/Nuhvi/pkarr/blob/main/design/relays.md).

Expanding on this API, a Gateway ****MUST**** support the following API, which is also made available via an
[OpenAPI document](#open-api-definition).

#### Get the Current Difficulty

Difficulty is exposed as an **OPTIONAL** endpoint based on support of [retention proofs](#retained-did-set).

- **Method:** `GET`
- **Path:** `/difficulty`
- **Returns:**
  - `200` - Success.
    - `hash` - **string** - The current hash.
    - `difficulty` - **integer** - The current difficulty.
  - `501` - Retention proofs not supported by this gateway.

```json
{
  "hash": "000000000000000000022be0c55caae4152d023dd57e8d63dc1a55c1f6de46e7",
  "difficulty": 26
}
```

#### Register or Update a DID

- **Method:** `PUT`
- **Path:** `/did/:id`
  - `id` - **string** - **REQUIRED** - ID of the DID to publish.
    - `did` - **string** - **REQUIRED** - The DID to register or update.
    - `sig` - **string** - **REQUIRED** - An unpadded base64URL-encoded signature of the [[ref:BEP44]] payload.
    - `seq` - **integer** - **REQUIRED** - A sequence number for the request. This number ****MUST**** be unique for each DID operation,
    recommended to be a Unix timestamp in seconds.
    - `v` - **string** - **REQUIRED** - An unpadded base64URL-encoded bencoded DNS packet containing the DID Document.
    - `retention_proof` - **string** – **OPTIONAL** - A retention proof calculated according to the [retention proof algorithm](#generating-a-retention-proof).
- **Returns:**
  - `202` - Accepted. The server has accepted the request as valid and will publish to the DHT.
  - `400` - Invalid request.
  - `401` - Invalid signature.
  - `409` - DID already exists with a higher sequence number. DID may be accepted if the [[ref:Gateway]] supports [historical resolution](#historical-resolution).

```json
{
  "did": "did:dht:example",
  "sig": "<unpadded-base64URL-encoded-signature>",
  "seq": 1234,
  "v": "<unpadded-base64URL-encoded bencoded DNS packet>"
}
```

Upon receiving a request to register a DID, the Gateway ****MUST**** verify the signature of the request and if valid
publish the DID Document to the DHT. If the DNS Packets contain a `_typ._did.` record, the Gateway ****MUST**** index the
DID by its type.

#### Resolving a DID

- **Method:** `GET`
- **Path:** `/did/:id`
  - `id` - **string** - **REQUIRED** - ID of the DID to resolve.
- **Returns:**
  - `200` - Success.
    - `did` - **object** - A JSON object representing the DID Document.
    - `types` - **array** - An array of [type strings](#type-indexing) for the DID.
    - `sequence_numbers` - **array** - An sorted array of seen sequence numbers, used with [historical resolution](#historical-resolution).
  - `400` - Invalid request.
  - `404` - DID not found.

```json
{
  "did": {
    "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
    "verificationMethod": [
      {
        "id": "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0",
        "type": "JsonWebKey",
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
    "authentication": [
      "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0"
    ],
    "assertionMethod": [
      "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y#0"
    ]
  },
  "types": [1, 4],
  "sequence_numbers": [1700356854, 1700461736]
}
```

Upon receiving a request to resolve a DID, the Gateway ****MUST**** query the DHT for the DID Document, and if found,
return the DID Document. If the records are not found in the DHT, the Gateway ****MAY**** fall back to its local storage.
If the DNS Packets contain a `_typ._did.` record, the Gateway ****MUST**** return the type index.

::: note
This API is not required to return the complete DNS packet but rather the DID Document and type index. If the full DNS
packet, with its signature data, is required it is ****RECOMMENDED**** to use the
[Relay API](https://github.com/Nuhvi/pkarr/blob/main/design/relays.md) directly.
:::

##### Historical Resolution

[[ref:Nodes]] ****MAY**** choose to support historical resolution, which is to surface different versions of the same [[ref:DID Document]],
sorted by sequence number, according to the rules set out in the section on [conflict resolution](#conflict-resolution).

Upon [resolving a DID](#resolving-a-did), the Gateway will return the parameter `sequence_numbers` if there exists
historical state for a given [[ref:DID]]. The following API can be used with specific sequence numbers to fetch historical state:

- **Method:** `GET`
- **Path:** `/did/:id?seq=:sequence_number`
  - `id` - **string** - **REQUIRED** - ID of the DID to resolve
  - `seq` - **integer** - **OPTIONAL** - Sequence number of the DID to resolve
- **Returns**:
  - `200` - Success.
    - `did` - **object** - A JSON object representing the DID Document.
    - `types` - **array** - An array of [type strings](#type-indexing) for the DID.
  - `400` - Invalid request.
  - `404` - DID not found for the given sequence number.
  - `501` - Historical resolution not supported by this gateway.

#### Deactivating a DID

To intentionally deactivate a DID, as opposed to letting the record cease being published to the DHT, a DID controller
follows the same process as [updating a DID](#register-or-update-a-did), but with a record format outlined in the
[section on deactivation](#deactivate).

Upon receiving a request to deactivate a DID, the Gateway ****MUST**** verify the signature of the request, and if valid,
stop republishing the DHT. If the DNS Packets contain a `_typ._did.` record, the Gateway ****MUST**** remove the type index.

#### Type Indexing

**Get Info**

- **Method:** `GET`
- **Path:** `/did/types`
- **Returns:**
  - `200` - Success.
    - **array** - An array of objects describing the known types.
  - `404` - Type indexing not supported.

```json
[
  {
    "type": 1,
    "description": "Organization"
  },
  {
    "type": 7,
    "description": "Financial Institution"
  }
]
```

**Get a Specific Type**

- **Method:** `GET`
- **Path:** `/did/types/:id`
  - `id` - **integer** - **REQUIRED** - The type to query from the index.
  - `offset` - **integer** - **OPTIONAL** - Specifies the starting position from where the type records should be retrieved (Default: `0`).
  - `limit` - **integer** - **OPTIONAL** - Specifies the maximum number of type records to retrieve (Default: `100`).
- **Returns:**
  - `200` - Success.
    - **array** - An array of DID Identifiers matching the associated type.
  - `400` - Invalid request.
  - `404` - Type not found.
  - `501` - Types not supported by this gateway.

```json
[
  "did:dht:i9xkp8ddcbcg8jwq54ox699wuzxyifsqx4jru45zodqu453ksz6y",
  "did:dht:uodqi99wuzxsz6yx445zxkp8ddwj9q54ocbcg8yifsqru45x63kj"
]
```

A query to the type index returns an array of DIDs matching the associated type. If the type is not found, a `404` is
returned. If no DIDs match the type, an empty array is returned.

## Implementation Considerations

### Conflict Resolution

According to [[ref:BEP44]] [[ref:Nodes]] can leverage the `seq` sequence number to handle conflicts:

> Storing nodes receiving a put request where seq is lower than or equal to what's already stored on the node,
****MUST**** reject the request. If the sequence number is equal, and the value is also the same, the node
****SHOULD**** reset its timeout counter.

When the sequence number is equal, but the value is different, nodes need to decide which value to accept and which
to reject. To make this determination nodes ****MUST**** compare the payloads lexicographically to determine a
[lexicographical order](https://en.wikipedia.org/wiki/Lexicographic_order), and reject the payload with a **lower**
lexicographical order.

### Size Constraints

[[ref:BEP44]] payload sizes are limited to 1000 bytes. Accordingly, we have defined [an efficient representation of a
DID Document](#dids-as-a-dns-packet) and leveraged DNS packet encoding to optimize our payload sizes. With this
encoding format, we recommend additional considerations to keep payload sizes minimal:

#### Representing Keys

Apart from the encoding of the cryptographic key itself, which cannot be further minimized in size, we ****RECOMMENDED****
the following representations of keys and their identifiers using `JsonWebKey`:

- The [[ref:Identity Key]]'s identifier ****MUST**** always be `0`.
- Key identifiers (`kid`s) ****MAY**** be omitted. If omitted, upon reconstruction of a DID Document, the JWK `kid`
is set to its JWK Thumbprint [[spec:RFC7638]].
- DID Document representations ****SHOULD**** always use fully qualified identifiers (e.g. `did:dht:uodqi99wuzxsz6yx445zxkp8ddwj9q54ocbcg8yifsqru45x63kj#0` as opposed to `0` or `#0`)

#### Historical Key State

However, key rotation is a commonly recommended security practice, which could lead to having many historically
necessary keys in a [[ref: DID Document]], increasing the size of the document. To address this concern and to
distinguish between keys that are currently active and keys that are no longer used but were once considered valid
users ****MAY**** make use of the [service property](https://www.w3.org/TR/did-core/#services) to store signed records
of historical key state, saving space in the [[ref:DID Document]] itself.

### Republishing Data

[[ref:Mainline]] offers a limited duration (approximately 2 hours) for retaining records in the DHT. To ensure the
verifiability of data signed by a [[ref:DID]], consistent republishing of [[ref:DID Document]] records is crucial. To
address this, it is ****RECOMMENDED**** to use [[ref:Gateways]] equipped with [[ref:Retention Proofs]] support.

### Rate Limiting

To reduce the risk of [Denial of Service Attacks](https://www.cisa.gov/news-events/news/understanding-denial-service-attacks),
spam, and other unwanted traffic, it is ****RECOMMENDED**** that [[ref:Gateways]] require [[ref:Retention Proofs]]. The
use of [[ref:Retention Proofs]] can act as an attack prevention measure, as it would be costly to scale retention proof
calculations. [[ref:Nodes]] ****MAY**** choose to explore other rate limiting techniques, such as IP-limiting, or an
access-token-based approach.

## Security and Privacy Considerations

When implementing and using the `did:dht` method, there are several security and privacy considerations to be aware of
to ensure expected and legitimate behavior.

### Data Conflicts

Malicious actors may try to force [[ref:Nodes]] into uncertain states by manipulating the sequence number associated
with a record set. There are three such cases to be aware of:

- **Low Sequence Number** - If a [[ref:Node]] has yet to see sequence numbers for a given record it ****MUST**** query
its peers to see if they have encountered the record. If a peer is found who has encountered the record, the record
with the latest sequence number must be selected. If the node has encountered greater sequence numbers before, the
node ****MAY**** reject the record set. If the node supports [historical resolution](#historical-resolution) it
****MAY**** choose to accept the request and insert the record into its historical ordered state.

- **Conflicting Sequence Number** - When a malicious actor publishes _valid but conflicting_ records to two different
[[ref:Mainline Nodes]] or [[ref:Gateways]]. Implementers are encouraged to follow the guidance outlined in [conflict
resolution](#conflict-resolution).

- **High Sequence Number** - Since sequence numbers ****MUST**** be second representations of [Unix
time](https://en.wikipedia.org/wiki/Unix_time), it is ****RECOMMENDED**** that nodes reject sequence numbers that
represent timestamps greater than **2 hours** into the future.

### Data Availability

Given the nature of decentralized distributed systems, there are no firm guarantees that all [[ref:Nodes]] have access
to the same state. It is ****RECOMMENDED**** to publish and read from multiple [[ref:Gateways]] to reduce such risks.
As an **optional** enhancement [[ref:Gateways]] ****MAY**** choose to share state amongst themselves via mechanisms
such as a [gossip protocol](https://en.wikipedia.org/wiki/Gossip_protocol).

### Data Authenticity

To enter into the DHT using [[ref:BEP44]] your records must be signed by an [[ref:Ed25519]] private key. When retrieving
records either through a [[ref:Mainline Node]] or a [[ref:Gateway]] is it ****RECOMMENDED**** that one verifies the
cryptographic integrity of the record themselves instead of trusting a node to have done the validation. Nodes that do
not return a signature value ****MUST NOT**** be trusted.

### Key Compromise

Since the `did:dht` uses a single, un-rotatable root key, there is a risk of root key compromise. Such a compromise
may be tough to detect without external assurances of identity. Implementers are encouraged to be aware of this
possibility and devise strategies that support entities transitioning to new [[ref:DIDs]] over time.

### Public Data

[[ref:Mainline]] is a public network. As such, there is risk in storing private, sensitive, or personally identifying
information (PII) on such a network. Storing such sensitive information on the network or in the contents of a `did:dht`
document is strongly discouraged.

### Data Retention

It is ****RECOMMENDED**** that [[ref:Gateways]] implement measures supporting the "[Right to be
Forgotten](https://en.wikipedia.org/wiki/Right_to_be_forgotten)," enabling precise control over the data retention duration.

### Cryptographic Risk

The security of data within the [[ref:Mainline DHT]] which relies on mutable records using [[ref:Ed25519]] keys—is
intrinsically tied to the strength of these keys and their underlying algorithms, as outlined in [[spec:RFC8032]].
Should vulnerabilities be discovered in [[ref:Ed25519]] or if advancements in quantum computing compromise its
cryptographic foundations, the [[ref:Mainline]] method could become obsolete.

## Appendix

### Test Vectors

#### Vector 1

A minimal DID Document.

**Identity Public Key JWK:**

```json
{
  "kty": "OKP",
  "crv": "Ed25519",
  "x": "YCcHYL2sYNPDlKaALcEmll2HHyT968M4UWbr-9CFGWE",
  "alg": "EdDSA",
  "kid": "0"
}
```

**DID Document:**

```json
{
  "id": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo",
  "verificationMethod": [
    {
      "id": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0",
      "type": "JsonWebKey",
      "controller": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo",
      "publicKeyJwk": {
        "kty": "OKP",
        "crv": "Ed25519",
        "x": "YCcHYL2sYNPDlKaALcEmll2HHyT968M4UWbr-9CFGWE",
        "alg": "EdDSA",
        "kid": "0"
      }
    }
  ],
  "authentication": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0"
  ],
  "assertionMethod": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0"
  ],
  "capabilityInvocation": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0"
  ],
  "capabilityDelegation": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0"
  ]
}
```

**DNS Resource Records:**

| Name      | Type | TTL  | Rdata       |
| --------- | ---- | ---- | ----------- |
| _did.     | TXT  | 7200 | vm=k0;auth=k0;asm=k0;inv=k0;del=k0 |
| _k0._did. | TXT  | 7200 | id=0;t=0;k=YCcHYL2sYNPDlKaALcEmll2HHyT968M4UWbr-9CFGWE |

#### Vector 2

A DID Document with two keys ([[ref:Identity Key]] and an uncompressed secp256k1 key), a service with multiple endpoints, two types to index, an aka, and controller properties.

**Identity Public Key JWK:**

```json
{
  "kty": "OKP",
  "crv": "Ed25519",
  "x": "YCcHYL2sYNPDlKaALcEmll2HHyT968M4UWbr-9CFGWE",
  "alg": "EdDSA",
  "kid": "0"
}
```

**secp256k1 Public Key JWK:**

```json
{
  "kty": "EC",
  "crv": "secp256k1",
  "x": "1_o0IKHGNamet8-3VYNUTiKlhVK-LilcKrhJSPHSNP0",
  "y": "qzU8qqh0wKB6JC_9HCu8pHE-ZPkDpw4AdJ-MsV2InVY",
  "alg": "ES256K",
  "kid": "0GkvkdCGu3DL7Mkv0W1DhTMCBT9-z0CkFqZoJQtw7vw"
}
```

**Key Purposes:** Assertion Method, Capability Invocation.

**Service:**

```json
{
  "id": "service-1",
  "type": "TestService",
  "serviceEndpoint": ["https://test-service.com/1", "https://test-service.com/2"]
}
```

**Types:** 1, 2, 3.

**DID Document:**

```json
{
  "id": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo",
  "controller": "did:example:abcd",
  "alsoKnownAs": ["did:example:efgh", "did:example:ijkl"],
  "verificationMethod": [
    {
      "id": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0",
      "type": "JsonWebKey",
      "controller": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo",
      "publicKeyJwk": {
        "kty": "OKP",
        "crv": "Ed25519",
        "x": "YCcHYL2sYNPDlKaALcEmll2HHyT968M4UWbr-9CFGWE",
        "alg": "EdDSA",
        "kid": "0"
      }
    },
    {
      "id": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0GkvkdCGu3DL7Mkv0W1DhTMCBT9-z0CkFqZoJQtw7vw",
      "type": "JsonWebKey",
      "controller": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo",
      "publicKeyJwk": {
        "kty": "EC",
        "crv": "secp256k1",
        "x": "1_o0IKHGNamet8-3VYNUTiKlhVK-LilcKrhJSPHSNP0",
        "y": "qzU8qqh0wKB6JC_9HCu8pHE-ZPkDpw4AdJ-MsV2InVY",
        "alg": "ES256K",
        "kid": "0GkvkdCGu3DL7Mkv0W1DhTMCBT9-z0CkFqZoJQtw7vw"
      }
    }
  ],
  "authentication": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0"
  ],
  "assertionMethod": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0",
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0GkvkdCGu3DL7Mkv0W1DhTMCBT9-z0CkFqZoJQtw7vw"
  ],
  "capabilityInvocation": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0",
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0GkvkdCGu3DL7Mkv0W1DhTMCBT9-z0CkFqZoJQtw7vw"
  ],
  "capabilityDelegation": [
    "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#0"
  ],
  "service": [
    {
      "id": "did:dht:cyuoqaf7itop8ohww4yn5ojg13qaq83r9zihgqntc5i9zwrfdfoo#service-1",
      "type": "TestService",
      "serviceEndpoint": ["https://test-service.com/1", "https://test-service.com/2"]
    }
  ]
}
```

**DNS Resource Records:**

| Name      | Type | TTL  | Rdata       |
| --------- | ---- | ---- | ----------- |
| _did.     | TXT  | 7200 | vm=k0,k1;svc=s0;auth=k0;asm=k0,k1;inv=k0,k1;del=k0                                                |
| _cnt.did. | TXT  | 7200 | did:example:abcd                                                                                  |
| _aka.did. | TXT  | 7200 | did:example:efgh,did:example:ijkl                                                                 |
| _k0.did.  | TXT  | 7200 | id=0;t=0;k=YCcHYL2sYNPDlKaALcEmll2HHyT968M4UWbr-9CFGWE                                            |
| _k1.did.  | TXT  | 7200 | id=0GkvkdCGu3DL7Mkv0W1DhTMCBT9-z0CkFqZoJQtw7vw;t=1;k=Atf6NCChxjWpnrfPt1WDVE4ipYVSvi4pXCq4SUjx0jT9 |
| _s0.did.  | TXT  | 7200 | id=service-1;t=TestService;se=https://test-service.com/1,https://test-service.com/2               |
| _typ.did. | TXT  | 7200 | id=1,2,3                                                                                          |

### Open API Definition

```yaml
[[insert: api.yaml]]
```

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

[[def:VC-JOSE-COSE]]
~ [Securing Verifiable Credentials using JOSE and COSE](https://www.w3.org/TR/vc-jose-cose/). O. Steele, M. Jones,
M. Prorock, G. Cohen; 04 December 2023. [W3C](https://www.w3.org/).

[[spec]]
