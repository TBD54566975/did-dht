# Build

From the root directory of the project you can run the following command to build a docker image:

```sh
docker build . -t did-dht -f build/Dockerfile
```

# Run

After building you can run the image with the following command:

```sh
docker run -itd -p8503:8503 -p8305:8305 did-dht
```