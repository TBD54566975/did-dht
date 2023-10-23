# Server Implementation

- Heavily a work-in-progress
- Designed to be run as a single instance

## Config

# TOML Config File

Config is managed using a [TOML](https://toml.io/en/) [file](../../config/dev.toml). There are sets of configuration values for the server
(e.g. which port to listen on), the services (e.g. which database to use), and each service.

Each service may define specific configuration, such as which DID methods are enabled for the DID service.

A full config example is [provided here](../../config/kitchensink.toml).

## Usage

How it works:

1. On startup the service loads default values into the `ServiceConfig`
2. Checks for a TOML config file:
   - If exists, load toml file
   - If does not exist, it uses a default config defined in the code inline
3. Loads the `config/.env` file and adds the env variables defined in this file to the final `ServiceConfig`

## Build & Run

Run:

```
docker build . -t did-dht -f build/Dockerfile
```

and then

```
docker run -p8305:8305 did-dht
```


