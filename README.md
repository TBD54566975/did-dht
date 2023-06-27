# DID DHT

A DID-based network relying upon the IPFS DHT which facilitates the discovery of DID Documents and associated endpoints.

![arch](docs/arch.png)

## Design

The service has three components: the DHT, crawler, and indexer.

### DHT

The DHT is a distributed hash table which stores the DID Documents and associated endpoints. 
To enter into the DHT, we must receive a message signed by a DID Document containing the endpoint(s) it wishes
to advertise.

The DHT handles queries for DID Documents and endpoints and re-publishes them to the network on a regular interval.

### Crawler

The crawler is a service which crawls the endpoints made available via the DHT. It will attempt to uncover new data
and make the data available for indexing via the indexer.

It is anticipated that there are custom crawlers and indexers for different types of data. For example, a crawler
can be specifically written for the ION network, which understands how to traverse the network and re-assemble the current
state of any DID Document on it.

### Indexer

The indexer is responsible for indexing all content made available by the crawler. It will store the data in a
database and make it available for query via a REST API.

The indexer will be used to back a UI which allows users to search for DID Documents, their endpoints, and associated
semantic data.

## Project Resources

| Resource                                                                               | Description                                                                   |
|----------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| [CODEOWNERS](https://github.com/TBD54566975/did-dht/blob/main/CODEOWNERS)              | Outlines the project lead(s)                                                  |
| [CODE_OF_CONDUCT](https://github.com/TBD54566975/did-dht/blob/main/CODE_OF_CONDUCT.md) | Expected behavior for project contributors, promoting a welcoming environment |
| [CONTRIBUTING](https://github.com/TBD54566975/did-dht/blob/main/CONTRIBUTING.md)       | Developer guide to build, test, run, access CI, chat, discuss, file issues    |
| [GOVERNANCE](https://github.com/TBD54566975/did-dht/blob/main/GOVERNANCE.md)           | Project governance                                                            |
| [SECURITY](https://github.com/TBD54566975/did-dht/blob/main/SECURITY.md)               | Vulnerability and bug reporting                                               |
| [LICENSE](https://github.com/TBD54566975/did-dht/blob/main/LICENSE)                    | Apache License, Version 2.0                                                   |
