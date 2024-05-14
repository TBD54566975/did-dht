package did

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// computeSHA256Hash computes the SHA-256 hash of a string and returns it as a hexadecimal string.
func computeSHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// has26LeadingZeros checks if the binary representation of the hash has 26 leading zeros.
func hasLeadingZeros(hash string, difficulty int) bool {
	// Convert hex hash to big.Int to handle binary conversion
	hashInt := new(big.Int)
	hashInt.SetString(hash, 16)

	// Convert to binary string
	binaryHash := fmt.Sprintf("%0256b", hashInt)

	target := strings.Repeat("0", difficulty)

	// Check if the first 26 characters are all zeros
	return strings.HasPrefix(binaryHash, target)
}

// solveRetentionChallenge generates the Retention Challenge Hash and checks if it meets the criteria.
func solveRetentionChallenge(didIdentifier, inputHash string, difficulty, nonce int) (string, bool) {
	// Concatenating the DID identifier with the retention value
	retentionValue := didIdentifier + (inputHash + fmt.Sprintf("%d", nonce))

	// Computing the SHA-256 hash
	hash := computeSHA256Hash(retentionValue)

	// Checking for the required number of leading zeros according to the difficulty
	return hash, hasLeadingZeros(hash, difficulty)
}

// validateRetentionSolution validates the Retention Solution.
func validateRetentionSolution(did, hash, retentionSolution string, difficulty int) bool {
	parts := strings.Split(retentionSolution, ":")
	if len(parts) != 2 {
		return false
	}

	nonce, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return false
	}

	retentionValue := did + hash + strconv.FormatUint(nonce, 10)
	computedHash := computeSHA256Hash(retentionValue)

	if !hasLeadingZeros(computedHash, difficulty) {
		return false
	}

	solutionHash := parts[0]
	return solutionHash == computedHash
}
