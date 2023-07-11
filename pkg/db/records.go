package db

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	dhtNamespace = "dht"
)

type Record struct {
	DID      string `json:"did,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	JWS      string `json:"jws,omitempty"`
}

type DHTRecordStorage interface {
	WriteRecord(record Record) error
	ReadRecord(id string) (*Record, error)
	ListRecords() ([]Record, error)
	DeleteRecord(id string) error
}

func (s *Storage) WriteRecord(record Record) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal record")
	}
	return s.Write(dhtNamespace, record.DID, recordBytes)
}

func (s *Storage) ReadRecord(id string) (*Record, error) {
	record, err := s.Read(dhtNamespace, id)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read record")
	}
	if len(record) == 0 {
		return nil, nil
	}
	var recordResult Record
	if err = json.Unmarshal(record, &recordResult); err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal record")
	}
	return &recordResult, nil
}

func (s *Storage) ListRecords() ([]Record, error) {
	records, err := s.ReadAll(dhtNamespace)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read records")
	}
	var recordResults []Record
	for _, record := range records {
		var recordResult Record
		if err = json.Unmarshal(record, &recordResult); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal record")
		}
		recordResults = append(recordResults, recordResult)
	}
	return recordResults, nil
}

func (s *Storage) DeleteRecord(id string) error {
	return s.Delete(dhtNamespace, id)
}
