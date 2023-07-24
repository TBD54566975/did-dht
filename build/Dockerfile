# Start with a base image containing Golang build environment
FROM golang:1.20.6-alpine AS build-stage

# Set the current working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the binary
RUN go build -tags jwx_es256k -o /did-dht ./cmd

# The second stage of Dockerfile (the runtime image)
FROM alpine:latest

# Set the current working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build-stage /did-dht .

# Expose the service on the specified port
EXPOSE 8305

# Command to run the binary
CMD [ "/app/did-dht" ]
