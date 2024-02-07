package pkarr

import (
	"encoding/base64"
	"fmt"
)

type Record struct {
	// Up to an 1000 bytes
	V []byte `json:"v" validate:"required"`
	// 32 bytes
	K []byte `json:"k" validate:"required"`
	// 64 bytes
	Sig []byte `json:"sig" validate:"required"`
	Seq int64  `json:"seq" validate:"required"`
}

func (r Record) String() string {
	encoding := base64.RawURLEncoding
	return fmt.Sprintf("pkarr.Record{K=%s V=%s Sig=%s Seq=%d}", encoding.EncodeToString(r.K), encoding.EncodeToString(r.V), encoding.EncodeToString(r.Sig), r.Seq)
}
