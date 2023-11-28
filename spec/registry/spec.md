The DID DHT Method Specification Registry 1.0
==================

**Specification Status**: Working Draft

**Latest Draft:** [tbd54566975.github.io/did-dht-method](https://tbd54566975.github.io/did-dht-method)

**Latest Update:** November 28, 2023

**Editors:**
~ [Gabe Cohen](https://github.com/decentralgabe)
~ [Daniel Buchner](https://github.com/csuwildcat)

**Participate:**
~ [GitHub repo](https://github.com/TBD54566975/did-dht-method)
~ [File a bug](https://github.com/TBD54566975/did-dht-method/issues)
~ [Commit history](https://github.com/TBD54566975/did-dht-method/commits/main)

## Abstract

This document serves as an official registry for all known extensions to the [DID DHT Method specification](../index.html).

All are welcome to propose changes to this registry.

## Extensions

### Key Type Index

Corresponds to the mapping, for a DID Document's DNS packet representation, of a cryptograhpic key type to its index value.

| Index | Key Type                                               |
| ----- | ------------------------------------------------------ |
| 0     | [Ed25519](https://ed25519.cr.yp.to/)                   |
| 1     | [secp256k1](https://en.bitcoin.it/wiki/Secp256k1)      |
| 2     | [secp256r1](https://neuromancer.sk/std/secg/secp256r1) |


An example [Verification Method](https://www.w3.org/TR/did-core/#verification-methods) record represented as a DNS TXT
record is as follows:

| Name      | Type | TTL  | Rdata                                                     |
| --------- | ---- | ---- | --------------------------------------------------------- |
| _k0._did. | TXT  | 7200 | id=abcd,t=0,k=r96mnGNgWGOmjt6g_3_0nd4Kls5-kknrd4DdPW8qtfw |

### Indexed Types

Represents an optional extension to a DID Document's DNS packet representation exposed as a [type index](../index.html#type-indexing).

::: note
The type `0` is reserved for DIDs that do not wish to associate themselves with a specific type, but wish to make
themselves discoverable via a [[ref:Gateway]]'s API.
:::

| Type Name               | Schema                                    | Type Integer |
|-------------------------|-------------------------------------------| ------------ |
| Discoverable            | -                                         | 0            |
| Organization            | https://schema.org/Organization           | 1            |
| Government Organization | https://schema.org/GovernmentOrganization | 2            |
| Corporation             | https://schema.org/Corporation            | 3            |
| Local Business          | https://schema.org/LocalBusiness          | 4            |
| Software Package        | https://schema.org/SoftwareSourceCode     | 5            |
| Web App                 | https://schema.org/WebApplication         | 6            |
| Financial Institution   | https://schema.org/FinancialService       | 7            |

### Interoperable DID Methods

As an **OPTIONAL** extension, some existing DID methods can leverage `did:dht` to broaden their feature set. This registry serves
to define such DID methods that are interoperable with `did:dht`. To enter into the registry each DID method ****MUST**** define
sections on _publishing_ and _resolving_ interoperable DIDs.

#### did:key

DID DHT is interoperable with the [[ref:DID Key method]] when using [[ref:Ed25519]] keys represented as JWKs [[spec:RFC7517]].

##### Publishing

To publish a [[ref:DID Key]] to the DHT, the process is as follows:

1. Verify the `did:key` value begins with the prefix `z6Mk`.
2. Decode the [[ref:Ed25519]] key in the `did:key` identifier, and re-encode it using [[ref:z-base-32]].
3. Expand the `did:key` using the [process outlined in the did:key spec](https://w3c-ccg.github.io/did-method-key/#read), 
with `options.publicKeyFormat` set to `JsonWebKey2020`.
4. Optionally, amend the [[ref:DID Document]] with additional properties (Verification Methods, Services, etc.).
5. Follow steps 3 onward in as outlined in the [create section of the spec](../index.html#create), using the identifier from step 2.

::: todo
[](https://github.com/w3c-ccg/did-method-key/issues/66)

[](https://github.com/TBD54566975/did-dht-method/issues/57)

Update this algorithm after the `did:key` spec is updated to support `did:dht`.
::: 

##### Resolving

To resolve a DID Key, the process is as follows:

1. Verify the `did:key` value begins with the prefix `z6Mk`.
2. Decode the [[ref:Ed25519]] key in the `did:key` identifier, and re-encode it using [[ref:z-base-32]].
3. Follow the process outlined in the [read section of the spec](../index.html#read) using the identifier from the prior step.
4. If lookup fails, fallback to the [guidance provided in the did:key spec](https://w3c-ccg.github.io/did-method-key/#read).

#### did:jwk

DID DHT is interoperable with the [[ref:DID JWK method]] when using JWKS [[spec:RFC7517]] representing [[ref:Ed25519]] keys.

##### Publishing

To publish a [[ref:DID JWK]] to the DHT, the process is as follows:

1. Expand the `did:jwk` using the [process outlined in the did:jwk spec](https://github.com/quartzjer/did-jwk/blob/main/spec.md#read).
2. Verify that the JWK represents an [[ref:Ed25519]] key.
3. Transform the [[ref:Ed25519]] key to its bytes representation and re-encode it using [[ref:z-base-32]].
4. Optionally, amend the [[ref:DID Document]] with additional properties (Verification Methods, Services, etc.).
5. Follow steps 3 onward in as outlined in the [create section](#create) above, using the identifier from step 3.

##### Resolving

1. Expand the `did:jwk` using the [process outlined in the did:jwk spec](https://github.com/quartzjer/did-jwk/blob/main/spec.md#read).
2. Verify that the JWK represents an [[ref:Ed25519]] key.
3. Transform the [[ref:Ed25519]] key to its bytes representation and re-encode it using [[ref:z-base-32]].
4. Follow the process outlined in the [Read section](#read) above using the identifier from the prior step.
5. If lookup fails, fallback to the [guidance provided in the did:jwk spec](https://github.com/quartzjer/did-jwk/blob/main/spec.md#read).

### Gateways

As an **OPTIONAL** feature of the DID DHT Method, Gateway operators have the opportunity to make their [gateways](../index.html#gateways) discoverable. This serves as a registry for such gateways.

Gateways ****MUST**** specify the source of their `HASH` value, and the mechanism by which their gateway(s) can be discovered.

#### Bitcoin Anchored Gateways

Bitcoin is used as a gateway discovery mechanism. By crawling the Bitcoin blockchain, one can discover gateways that are considered to be active. To be considered an active gateway, you ****MUST**** follow the steps outlined below, in order to have an active timelock, which acts as a proof-of-legitimacy for the gateway.

1. Generate a relative [[ref:timelock]] transaction for the Bitcoin blockchain with the following attributes:
    - Set the lock duration to 1000
    - Add locked value locked must be no less than the mean value of the upper quintile of [UTXOs](https://en.wikipedia.org/wiki/Unspent_transaction_output) as of a block that is no more than 10 blocks earlier from the block the locking transaction is included in (this effectively provides a 10 block grace period for the transaction to make it into the chain).
    - Add an `OP_RETURN` string composed of the following comma-separated values:
        - The block number used to compute the mean value of the upper quintile of [UTXOs](https://en.wikipedia.org/wiki/Unspent_transaction_output).
        - The `URI` where your node can be addressed
2. Include the [[ref:timelock]] transaction within 10 blocks of the block number specified for the average UTXO value calculation.
3. If this is a relocking transaction that refreshes an existing registration of a node:
    - The relocking transaction ****MUST**** spend the outputs of the lock it replaces.
    - If the operator wants to prevent other nodes and clients using the decentralized registry from dropping the Registered Gateway from their Registered Gateway list, the relocking transaction ****MUST**** be included in the blockchain within ten blocks of the previous lock's expiration.

##### Hash

The hash source to be used is [Bitcoin block hashes](https://csrc.nist.gov/glossary/term/block_header#:~:text=Definitions%3A,cryptographic%20nonce%20(if%20needed).). It is ****RECOMMENDED**** to use the most recent block hash value.

##### Discovery

To discover Bitcoin Anchored Gateways one must follow the following steps:

1. Starting at block height of `817714` traverse the chain, checking the value of the `OP_RETURN` field of transactions with _at least_ **6 confirmations**.
2. Find transactions that match the form `block number + uri`, as per the [steps outlined in the section above](#bitcoin-anchored-gateways).

## References

[[def:Timelock]]
~ [Timelock](https://github.com/bitcoin/bips/blob/master/bip-0065.mediawiki). P. Todd. 01 October 2014.
[Bitcoin](https://github.com/bitcoin).

[[def:DID Key, DID Key Method]]
~ [The did:key Method v0.7](https://w3c-ccg.github.io/did-method-key/). A DID Method for Static Cryptograhpic Keys.
D. Longley, D. Zagidulin, M. Sporny. [W3C CCG](https://w3c-ccg.github.io/).

[[def:DID JWK, DID JWK Method]]
~ [did:jwk](https://github.com/quartzjer/did-jwk/blob/main/spec.md). did:jwk is a deterministic transformation of a
JWK into a DID Document. J. Miller.

[[def:Ed25519]]
~ [Ed25519](https://ed25519.cr.yp.to/). D. J. Bernstein, N. Duif, T. Lange, P. Schwabe, B.-Y. Yang; 26 September 2011.
[ed25519.cr.yp.to](https://ed25519.cr.yp.to/).

[[def:z-base-32]]
~ [z-base-32](https://philzimmermann.com/docs/human-oriented-base-32-encoding.txt). Human-oriented base-32 encoding.
Z. O'Whielacronx; November 2002.

[[spec]]