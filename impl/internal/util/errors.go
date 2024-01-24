package util

type InvalidSignatureError struct{}

func (e *InvalidSignatureError) Error() string {
	return "invalid signature"
}

type HigherSequenceNumberError struct{}

func (e *HigherSequenceNumberError) Error() string {
	return "DID already exists with a higher sequence number"
}

type TypeNotFoundError struct{}

func (e *TypeNotFoundError) Error() string {
	return "type not found"
}
