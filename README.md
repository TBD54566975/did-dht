# did-dht-method

The `did:dht` method. Home to the [DID DHT Method Specification](https://did-dht.com), and a reference implementation of a 
gateway server in Go.

## Build & Run

To build and run the gateway server, from the `impl` directory run:

```
docker build . -t did-dht -f build/Dockerfile
```

and then

```
docker run -p8305:8305 did-dht
```

## Implementations

| Language   | Client | Server | Link |
| ---------- | ------ | ------ | ---- |
| Go         | Yes    | Yes    | [did-dht-go](./impl) |
| Typescript | Yes    | No     | [did-dht-ts](https://github.com/TBD54566975/web5-js/blob/main/packages/dids/src/did-dht.ts) |
| Kotlin     | Yes    | No     | [did-dht-kotlin](https://github.com/TBD54566975/web5-kt/tree/main/dids/src/main/kotlin/web5/sdk/dids/methods/dht) |
| Swift      | Yes    | No     | Coming soon! |
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
