package did

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestPOW(t *testing.T) {
	// Example usage of computeRetentionProof
	didIdentifier := "did:dht:test"
	bitcoinBlockHash := "000000000000000000022be0c55caae4152d023dd57e8d63dc1a55c1f6de46e7"

	// 26 leading zeros
	difficulty := 26

	timer := time.Now()
	for nonce := 0; nonce < math.MaxInt; nonce++ {
		hash, isValid := computeRetentionProof(didIdentifier, bitcoinBlockHash, difficulty, nonce)
		if isValid {
			fmt.Printf("Hash: %s\n", hash)
			fmt.Printf("Valid Retention Proof: %v\n", isValid)
			fmt.Printf("Nonce: %d\n", nonce)
			break
		}
	}
	fmt.Printf("Time taken: %s\n", time.Since(timer))
}
