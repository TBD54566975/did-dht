package storage

import (
	"encoding/json"
)

const (
	pkarrNamespace = "pkarr"
)

type PKARRRecord struct {
	// Up to an 1000 byte base64URL encoded string
	V string `json:"v" validate:"required"`
	// 32 byte base64URL encoded string
	K string `json:"k" validate:"required"`
	// 64 byte base64URL encoded string
	Sig string `json:"sig" validate:"required"`
	Seq int64  `json:"seq" validate:"required"`
}

type PKARRStorage interface {
	WriteRecord(record PKARRRecord) error
	ReadRecord(id string) (*PKARRRecord, error)
	ListRecords() ([]PKARRRecord, error)
}

// WriteRecord writes the given record to the storage
// TODO: don't overwrite existing records, store unique seq numbers
func (s *Storage) WriteRecord(record PKARRRecord) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.Write(pkarrNamespace, record.K, recordBytes)
}

// ReadRecord reads the record with the given id from the storage
func (s *Storage) ReadRecord(id string) (*PKARRRecord, error) {
	recordBytes, err := s.Read(pkarrNamespace, id)
	if err != nil {
		return nil, err
	}
	if len(recordBytes) == 0 {
		return nil, nil
	}
	var record PKARRRecord
	if err = json.Unmarshal(recordBytes, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// ListRecords lists all records in the storage
func (s *Storage) ListRecords() ([]PKARRRecord, error) {
	recordsMap, err := s.ReadAll(pkarrNamespace)
	if err != nil {
		return nil, err
	}
	var records []PKARRRecord
	for _, recordBytes := range recordsMap {
		var record PKARRRecord
		if err = json.Unmarshal(recordBytes, &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
