package pkarr

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/torrent/bencode"
)

type Record struct {
	Value          []byte   `json:"v" validate:"required"`
	Key            [32]byte `json:"k" validate:"required"`
	Signature      [64]byte `json:"sig" validate:"required"`
	SequenceNumber int64    `json:"seq" validate:"required"`
}

type Response struct {
	V   []byte   `validate:"required"`
	Seq int64    `validate:"required"`
	Sig [64]byte `validate:"required"`
}

func NewRecord(k []byte, v []byte, sig []byte, seq int64) (*Record, error) {
	record := Record{SequenceNumber: seq}

	if len(k) != 32 {
		return nil, errors.New("incorrect key length for pkarr record")
	}
	record.Key = [32]byte(k)

	if len(v) > 1000 {
		return nil, errors.New("pkarr record value too long")
	}
	record.Value = v

	if len(sig) != 64 {
		return nil, errors.New("incorrect sig length for pkarr record")
	}
	record.Signature = [64]byte(sig)

	return &record, nil
}

// Valid returns an error if the request is invalid; also validates the signature
func (r Record) Valid() error {
	if err := util.IsValidStruct(r); err != nil {
		return err
	}

	// validate the signature
	bv, err := bencode.Marshal(r.Value)
	if err != nil {
		return fmt.Errorf("error bencoding pkarr record: %v", err)
	}

	if !bep44.Verify(r.Key[:], nil, r.SequenceNumber, bv, r.Signature[:]) {
		return errors.New("signature is invalid")
	}
	return nil
}

func (r Record) Response() Response {
	return Response{
		V:   r.Value,
		Seq: r.SequenceNumber,
		Sig: r.Signature,
	}
}

func (r Record) Bep44() bep44.Put {
	return bep44.Put{
		V:   r.Value,
		K:   &r.Key,
		Sig: r.Signature,
		Seq: r.SequenceNumber,
	}
}

func (r Record) String() string {
	e := base64.RawURLEncoding
	return fmt.Sprintf("pkarr.Record{K=%s V=%s Sig=%s Seq=%d}", e.EncodeToString(r.Key[:]), e.EncodeToString(r.Value), e.EncodeToString(r.Signature[:]), r.SequenceNumber)
}

func RecordFromBep44(putMsg *bep44.Put) Record {
	return Record{
		Key:            *putMsg.K,
		Value:          putMsg.V.([]byte),
		Signature:      putMsg.Sig,
		SequenceNumber: putMsg.Seq,
	}
}
