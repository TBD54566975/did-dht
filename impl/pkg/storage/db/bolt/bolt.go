package bolt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/TBD54566975/did-dht-method/pkg/storage/pkarr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

const (
	pkarrNamespace = "pkarr"
)

type boltdb struct {
	db *bolt.DB
}

func NewBolt(path string) (*boltdb, error) {
	if path == "" {
		return nil, errors.New("path is required")
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, err
	}

	return &boltdb{db: db}, nil
}

// WriteRecord writes the given record to the storage
// TODO: don't overwrite existing records, store unique seq numbers
func (s *boltdb) WriteRecord(_ context.Context, record pkarr.Record) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.write(pkarrNamespace, record.K, recordBytes)
}

// ReadRecord reads the record with the given id from the storage
func (s *boltdb) ReadRecord(_ context.Context, id string) (*pkarr.Record, error) {
	recordBytes, err := s.read(pkarrNamespace, id)
	if err != nil {
		return nil, err
	}
	if len(recordBytes) == 0 {
		return nil, nil
	}
	var record pkarr.Record
	if err = json.Unmarshal(recordBytes, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// ListRecords lists all records in the storage
func (s *boltdb) ListRecords(_ context.Context) ([]pkarr.Record, error) {
	recordsMap, err := s.readAll(pkarrNamespace)
	if err != nil {
		return nil, err
	}
	var records []pkarr.Record
	for _, recordBytes := range recordsMap {
		var record pkarr.Record
		if err = json.Unmarshal(recordBytes, &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *boltdb) Close() error {
	return s.db.Close()
}

func (s *boltdb) write(namespace string, key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		if err = bucket.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})
}

func (s *boltdb) read(namespace, key string) ([]byte, error) {
	var result []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.Infof("namespace[%s] does not exist", namespace)
			return nil
		}
		result = bucket.Get([]byte(key))
		return nil
	})
	return result, err
}

func (s *boltdb) readAll(namespace string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.Warnf("namespace[%s] does not exist", namespace)
			return nil
		}
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			result[string(k)] = v
		}
		return nil
	})
	return result, err
}
