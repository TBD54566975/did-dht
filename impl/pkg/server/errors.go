package server

type InvalidSignatureError struct{}

func (e *InvalidSignatureError) Error() string {
	return "invalid signature"
}

type HigherSequenceNumberError struct{}

func (e *HigherSequenceNumberError) Error() string {
	return "DID already exists with a higher sequence number"
}
