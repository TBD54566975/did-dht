package dht

type DDTMessage struct {
	PublisherID string `json:"publisherId,omitempty"`
	Record      Record `json:"record,omitempty"`
}

type Record struct {
	DID      string `json:"did,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	JWS      string `json:"jws,omitempty"`
}
