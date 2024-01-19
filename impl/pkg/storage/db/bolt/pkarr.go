package bolt

import (
	"context"
	"encoding/json"

	"github.com/TBD54566975/did-dht-method/pkg/storage/pkarr"
)

const (
	pkarrNamespace = "pkarr"
)

// WriteRecord writes the given record to the storage
// TODO: don't overwrite existing records, store unique seq numbers
func (s *Storage) WriteRecord(_ context.Context, record pkarr.PkarrRecord) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.Write(pkarrNamespace, record.K, recordBytes)
}

// ReadRecord reads the record with the given id from the storage
func (s *Storage) ReadRecord(_ context.Context, id string) (*pkarr.PkarrRecord, error) {
	recordBytes, err := s.Read(pkarrNamespace, id)
	if err != nil {
		return nil, err
	}
	if len(recordBytes) == 0 {
		return nil, nil
	}
	var record pkarr.PkarrRecord
	if err = json.Unmarshal(recordBytes, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// ListRecords lists all records in the storage
func (s *Storage) ListRecords(_ context.Context) ([]pkarr.PkarrRecord, error) {
	recordsMap, err := s.ReadAll(pkarrNamespace)
	if err != nil {
		return nil, err
	}
	var records []pkarr.PkarrRecord
	for _, recordBytes := range recordsMap {
		var record pkarr.PkarrRecord
		if err = json.Unmarshal(recordBytes, &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
