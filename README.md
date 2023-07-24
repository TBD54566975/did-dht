# DID DHT

A DID-based network relying upon the IPFS DHT which facilitates the discovery of DID Documents and associated endpoints.

![arch](docs/arch.png)


# What Problems Are We Sovling?

1. DIDs don't have a way to be _addressable_. 

- No way to advertise how to be contacted outside of an optional service property. Not all DID methods make use of this property.
- Knowing a DID is a distinct problem from knowing how to resolve a DID (e.g. you may need a [universal resolver](https://dev.uniresolver.io/))

2. DID-based content discovery.

- DIDs will have content associated with them, but no standard way to access that content.
- [DWNs](https://identity.foundation/decentralized-web-node/spec/) aim to be one of these mechanisms, however even within DWNs there are no guarantees about the content a DWN has or that content’s shape (i.e. there’s no universal [web server directory index](https://en.wikipedia.org/wiki/Web_server_directory_index) and crawl-able file structure).
- There needs to be a way to both crawl and index DID-relative content.

3. Global DID messaging.

- There’s no standard way to send DID-authored messages in a broadcast-style
- There’s value in name-spaced broadcast communications channels, authenticated by DIDs
- Different levels of gating (verification, PoW, cost, trust-based) can be used to reduce spam


## Use Cases

Some sample use cases for the service. Feel free to propose more!

1. General content discovery: a search engine for decentralized data.
2. A single way to discover how to contact DIDs for any type of messaging.
3. For PFIs in the case of tbDEX: broadcasting offers; having a unified way to discover PFIs and their offerings.
4. A method to discover credential issuers.

# Design

[Decentralization is a spectrum](https://decentralgabe.xyz/a-journey-into-decentralization/). The DID DHT service enables a version of decentralization that allows anyone to run their own node and participate in the permission-less, blockchain-less network that is DID DHT. Given the requirements of a DHT to store content, validate it, and republish it, there will likely end up being a few large centralized nodes that provide the backbone of the network. This is not viewed as a problem, since the content on the network must always be independently verifiable, and anyone can choose to run their own node at any time. The design intentionally trades off centralizing forces for ephemerality in order to provide decentralization.

The most centralized piece of the architecture is a necessary bootstrap peers list for initial connection. Once bootstrapping has taken place, there’s no further need to rely on this peer.

-----------------------------------------

The service has three main components: the DHT, gossiper, and crawler/indexer.

## DHT

The DHT is a distributed hash table which stores the DID Documents and associated endpoints. It does not make use of the IPFS DHT, due to its constraints on keys/values (either public keys or IPNS, neither of which are suitable here since we want to support all DID methods and key types). This means we’re creating our own DHT where DIDs are keys and values are our own custom authenticated record type.

To enter into the DHT, the service must receive a message signed by a DID Document containing the endpoint(s) it wishes
to advertise. If the message is not signed, the service will sign the message with its own DID and publish it to the
network. Future enhancements will enable strategies to only accept self-signed messages, prevent spam, and be able to 
determine which record(s) should be overwritten in the event of a collision.

In addition to writing the record to the DHT, the service will gossip the record over gossipsub to its
peers. In this manner a node may become aware of a new entry in the DHT without having to query the DHT directly.

The DHT handles queries for DID Documents and endpoints and re-publishes them to the network on a regular interval.

## Crawler & Indexer

### Crawler

The crawler is a service which crawls the endpoints made available via the DHT. It will attempt to uncover new data
and make the data available for indexing via the indexer.

It is anticipated that there are custom crawlers and indexers for different types of data. For example, a crawler
can be specifically written for the ION network, which understands how to traverse the network and re-assemble the current
state of any DID Document on it, and leverage its [type registry](https://github.com/decentralized-identity/sidetree/blob/master/docs/type-registry.md)
to discover new data.

Another crawler can be specifically written for the DWNs that speak certain protocols, such as a protocol for indexing a personal blog, or a person’s messages a-la Twitter tweets.

The code will be modular and adaptable for different types of crawlers.

### Indexer

The indexer is responsible for indexing all content made available by the crawler. It will store the data in a
database and make it available for query via a REST API.

The indexer will be used to back a UI which allows users to search for DID Documents, their endpoints, and associated
semantic data. Similar to the crawler, it is anticipated that there are custom indexers for different types of data.

The code will be modular and adaptable for different types of indexers.

#### Verifiable Index

Users may want to become discoverable with human-readable names. To do this we plan on designing and implementing a
protocol for creating a verifiable index. This index will be created based on specific verifiable credential types
which provide name/account information from reputable sources. For example, a user may want to be discoverable by
their Twitter handle. To do this, they would create a verifiable credential which contains their Twitter handle, sign it
with their DID, and post this as a tweet to prove control over their account. The crawler would then discover this 
credential and index it. The indexer would then make this information available for query.

## Gossiper

Similar to the gossip service used by the DHT, we will add a component for a generic gossiper service. This allows anyone to define a `topic` within the `diddht` namespace and engage in publishing and subscribing, similar to a globally distributed [Kafka](https://kafka.apache.org/) instance. You can imagine topics for credential issuers, PFIs, verifiable reviews / trust score interactions, tweets, or any other category of messaging.

As an anti-spam prevention measure, content must be signed by a DID with a valid signature. This may not be a sufficient anti-spam prevention measure and a Proof of Work challenge will need to be implemented before publishing to a topic. More research here is needed.

## Bonus Ideas

1. A `did-dht` DID Method

A new DID method that has a special record type in the DHT which is an authenticated DID Document. This allows the DHT to store identifiers in addition to records about identifiers.

2. A CLI resource

Publish, read, query and listen to the service via a CLI. No need to run the service in a managed environment!

3. DID-DHT client library

A light client library that supports interacting with the service, the DHT, and gossip network. Supported in multiple languages for wide adoption.

4. A Decentralized Trust Registry

Build a trust registry based on the gossip service. A topic for trust with a well-known data schema for sending verifiable trust signals for DID-identified parties.

# Project Resources

| Resource                                                                               | Description                                                                   |
|----------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| [CODEOWNERS](https://github.com/TBD54566975/did-dht/blob/main/CODEOWNERS)              | Outlines the project lead(s)                                                  |
| [CODE_OF_CONDUCT](https://github.com/TBD54566975/did-dht/blob/main/CODE_OF_CONDUCT.md) | Expected behavior for project contributors, promoting a welcoming environment |
| [CONTRIBUTING](https://github.com/TBD54566975/did-dht/blob/main/CONTRIBUTING.md)       | Developer guide to build, test, run, access CI, chat, discuss, file issues    |
| [GOVERNANCE](https://github.com/TBD54566975/did-dht/blob/main/GOVERNANCE.md)           | Project governance                                                            |
| [SECURITY](https://github.com/TBD54566975/did-dht/blob/main/SECURITY.md)               | Vulnerability and bug reporting                                               |
| [LICENSE](https://github.com/TBD54566975/did-dht/blob/main/LICENSE)                    | Apache License, Version 2.0                                                   |
