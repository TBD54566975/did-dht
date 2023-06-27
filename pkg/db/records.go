package db

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	namespace = "did-dht"
)

type DDTRecord struct {
	Requester Requester `json:"requester,omitempty"`
	Publisher Publisher `json:"publisher,omitempty"`
	Record    Record    `json:"record,omitempty"`
	CreatedAt string    `json:"createdAt,omitempty"`
}

type Requester struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Publisher struct {
	ID   string `json:"id,omitempty"`
	DID  string `json:"did,omitempty"`
	Name string `json:"name,omitempty"`
}

type Record struct {
	DID      string `json:"did,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type DDTStorage interface {
	WriteRecord(record DDTRecord) error
	ReadRecord(id string) (*DDTRecord, error)
	ReadRecords() ([]DDTRecord, error)
}

func (s *Storage) WriteRecord(record DDTRecord) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal record")
	}
	return s.Write(namespace, record.Record.DID, recordBytes)
}

func (s *Storage) ReadRecord(id string) (*DDTRecord, error) {
	record, err := s.Read(namespace, id)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read record")
	}
	if len(record) == 0 {
		return nil, nil
	}
	var recordResult DDTRecord
	if err = json.Unmarshal(record, &recordResult); err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal record")
	}
	return &recordResult, nil
}

func (s *Storage) ReadRecords() ([]DDTRecord, error) {
	records, err := s.ReadAll(namespace)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read records")
	}
	var recordResults []DDTRecord
	for _, record := range records {
		var recordResult DDTRecord
		if err = json.Unmarshal(record, &recordResult); err != nil {
			return nil, errors.WithMessage(err, "failed to unmarshal record")
		}
		recordResults = append(recordResults, recordResult)
	}
	return recordResults, nil
}
