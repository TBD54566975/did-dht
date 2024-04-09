package pkarr

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/torrent/bencode"
	"github.com/goccy/go-json"
	"github.com/tv42/zbase32"
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

// NewRecord returns a new Record with the given key, value, signature, and sequence number
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

	if err := record.IsValid(); err != nil {
		return nil, err
	}

	return &record, nil
}

// IsValid returns an error if the request is invalid; also validates the signature
func (r Record) IsValid() error {
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

// Response returns the record as a Response
func (r Record) Response() Response {
	return Response{
		V:   r.Value,
		Seq: r.SequenceNumber,
		Sig: r.Signature,
	}
}

// BEP44 returns the record as a BEP44 Put message
func (r Record) BEP44() bep44.Put {
	return bep44.Put{
		V:   r.Value,
		K:   &r.Key,
		Sig: r.Signature,
		Seq: r.SequenceNumber,
	}
}

// String returns a string representation of the record
func (r Record) String() string {
	e := base64.RawURLEncoding
	return fmt.Sprintf("pkarr.Record{K=%s V=%s Sig=%s Seq=%d}", zbase32.EncodeToString(r.Key[:]), e.EncodeToString(r.Value), e.EncodeToString(r.Signature[:]), r.SequenceNumber)
}

// ID returns the base32 encoded key as a string
func (r Record) ID() string {
	return zbase32.EncodeToString(r.Key[:])
}

// Hash returns the SHA256 hash of the record as a string
func (r Record) Hash() (string, error) {
	recordBytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(sha256.New().Sum(recordBytes)), nil
}

// RecordFromBEP44 returns a Record from a BEP44 Put message
func RecordFromBEP44(putMsg *bep44.Put) Record {
	return Record{
		Key:            *putMsg.K,
		Value:          putMsg.V.([]byte),
		Signature:      putMsg.Sig,
		SequenceNumber: putMsg.Seq,
	}
}
