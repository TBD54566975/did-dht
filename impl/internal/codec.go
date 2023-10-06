package internal

import (
	"bytes"

	"github.com/andybalholm/brotli"
	"github.com/sirupsen/logrus"
)

const (
	// Version is the version of the encoding
	Version = 0
)

// Encode compresses the input data using Brotli compression
// the input data is expected to be a byte representation of JSON data
// the result is a byte array prepended with the version
func Encode(data []byte) ([]byte, error) {
	// Create a buffer to store the compressed data
	var compressedBuffer bytes.Buffer

	// Create a Brotli writer with the buffer
	writer := brotli.NewWriterLevel(&compressedBuffer, brotli.BestCompression)

	// Write the input data to the Brotli writer
	if _, err := writer.Write(data); err != nil {
		logrus.WithError(err).Error("failed to write data to brotli writer")
		return nil, err
	}

	// Close the writer to finalize the compression
	if err := writer.Close(); err != nil {
		logrus.WithError(err).Error("failed to close brotli writer")
		return nil, err
	}

	// Append the version to the beginning of the compressed data
	compressedBytes := compressedBuffer.Bytes()
	return append([]byte{Version}, compressedBytes...), nil
}

// Decode decompresses the input data using Brotli decompression
// the input data is expected to be a byte array prepended with the version
// the result is expected to be a byte array of JSON data after decompression
func Decode(encoded []byte) ([]byte, error) {
	// Ensure the version is correct
	// if encoded[0] != Version {
	// 	return nil, fmt.Errorf("invalid version: got[%d], expected [%d]", encoded[0], Version)
	// }

	// Create a buffer to store the decompressed data
	var decompressedBuffer bytes.Buffer

	// Create a Brotli reader with the compressed data
	reader := brotli.NewReader(bytes.NewReader(encoded))

	// Copy the decompressed data from the reader to the buffer
	if _, err := decompressedBuffer.ReadFrom(reader); err != nil {
		logrus.WithError(err).Error("failed to read from brotli reader")
		return nil, err
	}

	// Convert the decompressed data to bytes
	return decompressedBuffer.Bytes(), nil
}
