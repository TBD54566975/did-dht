package pkarr

type Record struct {
	// Up to an 1000 bytes
	V []byte `json:"v" validate:"required"`
	// 32 bytes
	K []byte `json:"k" validate:"required"`
	// 64 bytes
	Sig []byte `json:"sig" validate:"required"`
	Seq int64  `json:"seq" validate:"required"`
}
