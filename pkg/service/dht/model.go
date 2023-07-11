package dht

type Record struct {
	DID      string `json:"did,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	JWS      string `json:"jws,omitempty"`
}
