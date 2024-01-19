package pkarr

type PkarrRecord struct {
	// Up to an 1000 byte base64URL encoded string
	V string `json:"v" validate:"required"`
	// 32 byte base64URL encoded string
	K string `json:"k" validate:"required"`
	// 64 byte base64URL encoded string
	Sig string `json:"sig" validate:"required"`
	Seq int64  `json:"seq" validate:"required"`
}

type PKARRStorage interface {
	WriteRecord(record PkarrRecord) error
	ReadRecord(id string) (*PkarrRecord, error)
	ListRecords() ([]PkarrRecord, error)
}
