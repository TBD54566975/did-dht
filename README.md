[![godoc did-dht](https://img.shields.io/badge/godoc-did_dht-blue)](https://github.com/TBD54566975/did-dht/impl)
[![go version 1.22.5](https://img.shields.io/badge/go_version-1.22.5-brightgreen)](https://go.dev/)
[![license Apache 2](https://img.shields.io/badge/license-Apache%202-black)](https://github.com/TBD54566975/did-dht/blob/main/LICENSE)
[![issues](https://img.shields.io/github/issues/TBD54566975/did-dht)](https://github.com/TBD54566975/did-dht/issues)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/TBD54566975/did-dht/ci.yml)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTBD54566975%2Fdid-dht.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FTBD54566975%2Fdid-dht?ref=badge_shield)

# did-dht

The `did:dht` method. Home to the [DID DHT Method Specification](https://did-dht.com), and a reference
implementation of a gateway server in Go.

## 🎉 Hacktoberfest 2024 🎉

`did-dht` is a participating project in Hacktoberfest 2024! We’re so excited for your contributions, and have created a wide variety of issues so that anyone can contribute. Whether you're a seasoned developer or a first-time open source contributor, there's something for everyone.

### To get started:
1. Read the [contributing guide](https://github.com/taniashiba/did-dht/blob/main/CONTRIBUTING.md).
2. Read the [code of conduct](https://github.com/taniashiba/did-dht/blob/main/CODE_OF_CONDUCT.md).
3. Choose a task from this project's Hacktoberfest issues in our [Project Hub](https://github.com/TBD54566975/did-dht/issues/292) and follow the instructions. Each issue has the 🏷️ `hacktoberfest` label.

Have questions? Connecting with us in our [Discord community](https://discord.gg/tbd) in the `#hacktoberfest` project channel.

---

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
| Go         | Yes    | Yes    | [did-dht](./impl), [web5-go](https://github.com/TBD54566975/web5-go/tree/main/dids/diddht) |
| Typescript | Yes    | No     | [web5-js](https://github.com/TBD54566975/web5-js/blob/main/packages/dids/src/methods/did-dht.ts) |
| Kotlin     | Yes    | No     | [web5-kt](https://github.com/TBD54566975/web5-kt/tree/main/dids/src/main/kotlin/web5/sdk/dids/methods/dht) |
| Swift      | Yes    | No     | [web5-swift](https://github.com/TBD54566975/web5-swift/blob/main/Sources/Web5/Dids/Methods/DIDDHT.swift) |
| Dart       | Yes    | No     | [web5-dart](https://github.com/TBD54566975/web5-dart/tree/main/packages/web5/lib/src/dids/did_dht) |
| Rust       | Yes    | No     | [web5-rs](https://github.com/TBD54566975/web5-rs/tree/main/crates/dids/src/methods/dht) |

## Project Resources

| Resource                                   | Description                                                                    |
| ------------------------------------------ | ------------------------------------------------------------------------------ |
| [Specification](./spec.md)                  | The DID Method specification                                                    |
| [CODEOWNERS](./CODEOWNERS)                 | Outlines the project lead(s)                                                   |
| [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) | Expected behavior for project contributors, promoting a welcoming environment  |
| [CONTRIBUTING.md](./CONTRIBUTING.md)       | Developer guide to build, test, run, access CI, chat, discuss, file issues      |
| [GOVERNANCE.md](./GOVERNANCE.md)           | Project governance                                                             |
| [LICENSE](./LICENSE)                       | Apache License, Version 2.0                                                    |


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTBD54566975%2Fdid-dht.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FTBD54566975%2Fdid-dht?ref=badge_large)