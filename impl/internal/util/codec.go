package util

import (
	"bytes"

	"github.com/andybalholm/brotli"
	"github.com/sirupsen/logrus"
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

	return compressedBuffer.Bytes(), nil
}

// Decode decompresses the input data using Brotli decompression
// the input data is expected to be a byte array
// the result is expected to be a byte array of JSON data after decompression
func Decode(encoded []byte) ([]byte, error) {
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
