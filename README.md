[![godoc did-dht-method](https://img.shields.io/badge/godoc-did_dht_method-blue)](https://github.com/TBD54566975/did-dht-method/impl)
[![go version 1.22.1](https://img.shields.io/badge/go_version-1.22.1-brightgreen)](https://go.dev/)
[![license Apache 2](https://img.shields.io/badge/license-Apache%202-black)](https://github.com/TBD54566975/did-dht-method/blob/main/LICENSE)
[![issues](https://img.shields.io/github/issues/TBD54566975/did-dht-method)](https://github.com/TBD54566975/did-dht-method/issues)
![push](https://github.com/TBD54566975/did-dht-method/workflows/did-dht-ci/badge.svg?branch=main&event=push)

# did-dht-method

The `did:dht` method. Home to the [DID DHT Method Specification](https://did-dht.com), and a reference implementation of a
gateway server in Go.

## Build & Run

### Quickstart

To build and run in a single command `./scripts/quickstart.sh`.

```
Usage: ./scripts/quickstart.sh [options]

Builds and runs the did-dht server

Options
  -h, --help          show this help message and exit
  -c, --commit=<hash> commit hash for `docker build` (default: HEAD)
  -t, --tag=<tag>     tag name for `docker build` (default: did-dht:latest)
  -d, --detach        run the container in the background (default: false)
  -k, --keep          keep the container after it exits (default: false)
  -n, --name=<name>   name to give the container (default: did-dht-server)
  -p, --port=<port>   ports to publish the host/container (default: 8305:8305)
  --skip-run          skip running the container (default: false)
 ```

### `docker`

To build and run the gateway server, from the `impl` directory run:

```sh
docker build \
  --build-arg GIT_COMMIT_HASH=$(git rev-parse head) \
  --tag did-dht \
  --file build/Dockerfile .
```

and then

```sh
docker run \
    --publish 8305:8305 \
    --publish 6881:6881/udp \
    did-dht
```

## Implementations

| Language   | Client | Server | Link |
| ---------- | ------ | ------ | ---- |
| Go         | Yes    | Yes    | [did-dht-method](./impl), [web5-go](https://github.com/TBD54566975/web5-go/tree/main/dids/diddht) |
| Typescript | Yes    | No     | [web5-js](https://github.com/TBD54566975/web5-js/blob/main/packages/dids/src/methods/did-dht.ts) |
| Kotlin     | Yes    | No     | [web5-kt](https://github.com/TBD54566975/web5-kt/tree/main/dids/src/main/kotlin/web5/sdk/dids/methods/dht) |
| Swift      | Yes    | No     | [web5-swift](https://github.com/TBD54566975/web5-swift/blob/main/Sources/Web5/Dids/Methods/DIDDHT.swift) |
| Rust       | Yes    | No     | Coming soon! |

## Project Resources

| Resource                                   | Description                                                                    |
| ------------------------------------------ | ------------------------------------------------------------------------------ |
| [Specification](./spec.md)                  | The DID Method specification                                                    |
| [CODEOWNERS](./CODEOWNERS)                 | Outlines the project lead(s)                                                   |
| [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) | Expected behavior for project contributors, promoting a welcoming environment  |
| [CONTRIBUTING.md](./CONTRIBUTING.md)       | Developer guide to build, test, run, access CI, chat, discuss, file issues      |
| [GOVERNANCE.md](./GOVERNANCE.md)           | Project governance                                                             |
| [LICENSE](./LICENSE)                       | Apache License, Version 2.0                                                    |
