package internal

// Identities is the mainline id mapped to its identity data
type Identities map[string]Identity

type Identity struct {
	// The public key of the identity.
	Base58PublicKey string `json:"publicKey"`
	// The private key of the identity.
	Base58PrivateKey string `json:"privateKey"`
	// Records is a slice of slices of strings, such as: [["foo", "bar"]].
	Records [][]any `json:"records"`
}
