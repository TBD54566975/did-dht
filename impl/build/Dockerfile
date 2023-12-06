FROM golang:1.21.5-alpine

# Create directory for our app inside the container
WORKDIR /app

# Prepare dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy code /to the container image.
COPY . ./

# Use ARG to declare the variable
ARG GIT_COMMIT_HASH

# Use ENV to set the environment variable
ENV GIT_COMMIT_HASH=$GIT_COMMIT_HASH

# Build using the environment variable
RUN go build -ldflags="-X main.commitHash=$GIT_COMMIT_HASH" -tags jwx_es256k -o /did-dht ./cmd

EXPOSE 8305

CMD [ "/did-dht" ]