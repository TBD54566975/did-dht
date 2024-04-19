package dht

import (
	"bytes"
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

type BEP44Response struct {
	V   []byte   `validate:"required"`
	Seq int64    `validate:"required"`
	Sig [64]byte `validate:"required"`
}

// Equals returns true if the response is equal to the other response
func (r BEP44Response) Equals(other BEP44Response) bool {
	return r.Seq == other.Seq && bytes.Equal(r.V, other.V) && r.Sig == other.Sig
}

// BEP44Record represents a record in the DHT
type BEP44Record struct {
	Value          []byte   `json:"v" validate:"required"`
	Key            [32]byte `json:"k" validate:"required"`
	Signature      [64]byte `json:"sig" validate:"required"`
	SequenceNumber int64    `json:"seq" validate:"required"`
}

// NewBEP44Record returns a new BEP44Record with the given key, value, signature, and sequence number
func NewBEP44Record(k []byte, v []byte, sig []byte, seq int64) (*BEP44Record, error) {
	record := BEP44Record{SequenceNumber: seq}

	if len(k) != 32 {
		return nil, errors.New("incorrect key length for bep44 record")
	}
	record.Key = [32]byte(k)

	if len(v) > 1000 {
		return nil, errors.New("bep44 record value too long")
	}
	record.Value = v

	if len(sig) != 64 {
		return nil, errors.New("incorrect sig length for bep44 record")
	}
	record.Signature = [64]byte(sig)

	if err := record.IsValid(); err != nil {
		return nil, err
	}

	return &record, nil
}

// IsValid returns an error if the request is invalid; also validates the signature
func (r BEP44Record) IsValid() error {
	if err := util.IsValidStruct(r); err != nil {
		return err
	}

	// validate the signature
	bv, err := bencode.Marshal(r.Value)
	if err != nil {
		return fmt.Errorf("error bencoding bep44 record: %v", err)
	}

	if !bep44.Verify(r.Key[:], nil, r.SequenceNumber, bv, r.Signature[:]) {
		return errors.New("signature is invalid")
	}
	return nil
}

// Response returns the record as a BEP44Response
func (r BEP44Record) Response() BEP44Response {
	return BEP44Response{
		V:   r.Value,
		Seq: r.SequenceNumber,
		Sig: r.Signature,
	}
}

// Put returns the record as a bep44.Put message
func (r BEP44Record) Put() bep44.Put {
	return bep44.Put{
		V:   r.Value,
		K:   &r.Key,
		Sig: r.Signature,
		Seq: r.SequenceNumber,
	}
}

// String returns a string representation of the record
func (r BEP44Record) String() string {
	e := base64.RawURLEncoding
	return fmt.Sprintf("dht.BEP44Record{K=%s V=%s Sig=%s Seq=%d}", zbase32.EncodeToString(r.Key[:]), e.EncodeToString(r.Value), e.EncodeToString(r.Signature[:]), r.SequenceNumber)
}

// ID returns the base32 encoded key as a string
func (r BEP44Record) ID() string {
	return zbase32.EncodeToString(r.Key[:])
}

// Hash returns the SHA256 hash of the record as a string
func (r BEP44Record) Hash() (string, error) {
	recordBytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(sha256.New().Sum(recordBytes)), nil
}

// RecordFromBEP44 returns a BEP44Record from a bep44.Put message
func RecordFromBEP44(putMsg *bep44.Put) BEP44Record {
	return BEP44Record{
		Key:            *putMsg.K,
		Value:          putMsg.V.([]byte),
		Signature:      putMsg.Sig,
		SequenceNumber: putMsg.Seq,
	}
}
