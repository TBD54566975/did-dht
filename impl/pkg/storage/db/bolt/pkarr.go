package bolt

import (
	"encoding/base64"
	"fmt"

	"github.com/TBD54566975/ssi-sdk/util"

	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
)

var (
	encoding = base64.RawURLEncoding
)

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
		V:   encoding.EncodeToString(r.Value[:]),
		K:   encoding.EncodeToString(r.Key[:]),
		Sig: encoding.EncodeToString(r.Signature[:]),
		Seq: r.SequenceNumber,
	}
}

func (b base64PkarrRecord) Decode() (*pkarr.Record, error) {
	v, err := encoding.DecodeString(b.V)
	if err != nil {
		return nil, fmt.Errorf("error parsing pkarr value field: %v", err)
	}

	k, err := encoding.DecodeString(b.K)
	if err != nil {
		return nil, fmt.Errorf("error parsing pkarr key field: %v", err)
	}

	sig, err := encoding.DecodeString(b.Sig)
	if err != nil {
		return nil, fmt.Errorf("error parsing pkarr sig field: %v", err)
	}

	record, err := pkarr.NewRecord(k, v, sig, b.Seq)
	if err != nil {
		// TODO: do something useful if this happens
		return nil, util.LoggingErrorMsg(err, "error loading record from database, skipping")
	}
	return record, nil
}
