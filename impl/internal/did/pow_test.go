package did

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestPOW(t *testing.T) {
	// Example usage of solveRetentionChallenge
	didIdentifier := "did:dht:test"
	inputHash := "000000000000000000022be0c55caae4152d023dd57e8d63dc1a55c1f6de46e7"

	// 26 leading zeros
	difficulty := 26

	timer := time.Now()
	for nonce := 0; nonce < math.MaxInt; nonce++ {
		solution, isValid := solveRetentionChallenge(didIdentifier, inputHash, difficulty, nonce)
		if isValid {
			fmt.Printf("Solution: %s\n", solution)
			fmt.Printf("Valid Retention Solution: %v\n", isValid)
			fmt.Printf("Nonce: %d\n", nonce)

			isValidRetentionSolution := validateRetentionSolution(didIdentifier, inputHash, fmt.Sprintf("%s:%d", solution, nonce), difficulty)
			fmt.Printf("Validated Solution: %v\n", isValidRetentionSolution)
			break
		}
	}

	fmt.Printf("Time taken: %s\n", time.Since(timer))
}
