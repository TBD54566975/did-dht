package bolt

import (
	"encoding/base64"
	"fmt"

	"github.com/TBD54566975/did-dht-method/pkg/storage/pkarr"
)

var encoding = base64.RawURLEncoding

type base64PkarrRecord struct {
	// Up to an 1000 byte base64URL encoded string
	V string `json:"v" validate:"required"`
	// 32 byte base64URL encoded string
	K string `json:"k" validate:"required"`
	// 64 byte base64URL encoded string
	Sig string `json:"sig" validate:"required"`
	Seq int64  `json:"seq" validate:"required"`
}

func encodeRecord(r pkarr.Record) base64PkarrRecord {
	return base64PkarrRecord{
		V:   encoding.EncodeToString(r.V),
		K:   encoding.EncodeToString(r.K),
		Sig: encoding.EncodeToString(r.Sig),
		Seq: r.Seq,
	}
}

func (b base64PkarrRecord) Decode() (pkarr.Record, error) {
	record := pkarr.Record{}

	v, err := encoding.DecodeString(b.V)
	if err != nil {
		return record, fmt.Errorf("error parsing pkarr value field: %v", err)
	}

	k, err := encoding.DecodeString(b.K)
	if err != nil {
		return record, fmt.Errorf("error parsing pkarr key field: %v", err)
	}

	sig, err := encoding.DecodeString(b.Sig)
	if err != nil {
		return record, fmt.Errorf("error parsing pkarr sig field: %v", err)
	}

	return pkarr.Record{
		V:   v,
		K:   k,
		Sig: sig,
		Seq: b.Seq,
	}, nil
}
