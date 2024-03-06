package bolt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"

	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/TBD54566975/did-dht-method/pkg/telemetry"
)

const (
	pkarrNamespace = "pkarr"
)

type BoltDB struct {
	db *bolt.DB
}

type boltRecord struct {
	key, value []byte
}

// NewBolt creates a BoltDB-based implementation of storage.Storage
func NewBolt(path string) (*BoltDB, error) {
	if path == "" {
		return nil, errors.New("path is required")
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, err
	}

	return &BoltDB{db: db}, nil
}

// WriteRecord writes the given record to the storage
// TODO: don't overwrite existing records, store unique seq numbers
func (s *BoltDB) WriteRecord(ctx context.Context, record pkarr.Record) error {
	ctx, span := telemetry.GetTracer("pkg/storage/bolt").Start(ctx, "WriteRecord")
	defer span.End()

	encoded := encodeRecord(record)
	recordBytes, err := json.Marshal(encoded)
	if err != nil {
		return err
	}

	return s.write(ctx, pkarrNamespace, encoded.K, recordBytes)
}

// ReadRecord reads the record with the given id from the storage
func (s *BoltDB) ReadRecord(ctx context.Context, id []byte) (*pkarr.Record, error) {
	ctx, span := telemetry.GetTracer("pkg/storage/bolt").Start(ctx, "ReadRecord")
	defer span.End()

	recordBytes, err := s.read(ctx, pkarrNamespace, encoding.EncodeToString(id))
	if err != nil {
		return nil, err
	}
	if len(recordBytes) == 0 {
		return nil, nil
	}

	var b64record base64PkarrRecord
	if err = json.Unmarshal(recordBytes, &b64record); err != nil {
		return nil, err
	}

	record, err := b64record.Decode()
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ListRecords lists all records in the storage
func (s *BoltDB) ListRecords(ctx context.Context, nextPageToken []byte, pagesize int) ([]pkarr.Record, []byte, error) {
	ctx, span := telemetry.GetTracer("pkg/storage/bolt").Start(ctx, "ListRecords")
	defer span.End()

	boltRecords, err := s.readSeveral(ctx, pkarrNamespace, nextPageToken, pagesize)
	if err != nil {
		return nil, nil, err
	}

	var records []pkarr.Record
	for _, recordBytes := range boltRecords {
		var encodedRecord base64PkarrRecord
		if err = json.Unmarshal(recordBytes.value, &encodedRecord); err != nil {
			return nil, nil, err
		}

		record, err := encodedRecord.Decode()
		if err != nil {
			return nil, nil, err
		}

		records = append(records, *record)
	}

	if len(boltRecords) == pagesize {
		nextPageToken = boltRecords[len(boltRecords)-1].key
	} else {
		nextPageToken = nil
	}

	return records, nextPageToken, nil
}

func (s *BoltDB) Close() error {
	return s.db.Close()
}

func (s *BoltDB) write(ctx context.Context, namespace string, key string, value []byte) error {
	_, span := telemetry.GetTracer("pkg/storage/bolt").Start(ctx, "write")
	defer span.End()

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

func (s *BoltDB) read(ctx context.Context, namespace, key string) ([]byte, error) {
	_, span := telemetry.GetTracer("pkg/storage/bolt").Start(ctx, "read")
	defer span.End()

	var result []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.WithField("namespace", namespace).Info("namespace does not exist")
			return nil
		}
		result = bucket.Get([]byte(key))
		return nil
	})
	return result, err
}

func (s *BoltDB) readAll(namespace string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.WithField("namespace", namespace).Warn("namespace does not exist")
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

func (s *BoltDB) readSeveral(ctx context.Context, namespace string, after []byte, count int) ([]boltRecord, error) {
	_, span := telemetry.GetTracer("pkg/storage/bolt").Start(ctx, "readSeveral")
	defer span.End()

	var result []boltRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.WithField("namespace", namespace).Warn("namespace does not exist")
			return nil
		}

		cursor := bucket.Cursor()

		var k []byte
		var v []byte
		if after != nil {
			cursor.Seek(after)
			k, v = cursor.Next()
		} else {
			k, v = cursor.First()
		}

		for ; k != nil; k, v = cursor.Next() {
			result = append(result, boltRecord{key: k, value: v})
			if len(result) >= count {
				break
			}
		}
		return nil
	})
	return result, err
}
