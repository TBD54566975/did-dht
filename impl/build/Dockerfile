FROM golang:1.23.2-alpine

# Create directory for our app inside the container
WORKDIR /app

# Prepare dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy code /to the container image.
COPY . ./

# Use ARG to declare the variable
ARG VERSION

# Use ENV to set the environment variable
ENV VERSION=$VERSION

# Build using the environment variable
RUN go build -ldflags="-X github.com/TBD54566975/did-dht/config.Version=$VERSION" -tags jwx_es256k -o /did-dht ./cmd

EXPOSE 8305
EXPOSE 6881/udp

CMD [ "/did-dht" ]