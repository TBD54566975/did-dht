# Server Implementation

## Config

### TOML Config File

Config is managed using a [TOML](https://toml.io/en/) [file](../../config/dev.toml). There are sets of configuration value
s for the server (e.g. which port to listen on) and each sub-component (e.g. which database to use).

## Usage

How it works:

1. On startup the service loads default values into the `ServiceConfig`
2. Checks for a TOML config file:
   - If exists, load toml file
   - If does not exist, it uses a default config defined in the code inline
3. Loads the `config/.env` file and adds the env variables defined in this file to the final `ServiceConfig`

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

Run:

```sh
docker build \
  --build-arg VERSION=$(git describe --always --tags) \
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

### Postgres

To use a postgres database as the storage backend, set configuration option `storage_uri` to a `postgres://` URI with
the database connection string. The schema will be created or updated as needed while the program starts.