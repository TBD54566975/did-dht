package bolt

import (
	"bytes"
	"context"
	"encoding/binary"
	"time"

	"github.com/goccy/go-json"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"

	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/telemetry"
)

const (
	dhtNamespace    = "pkarr"
	failedNamespace = "failed"
)

type Bolt struct {
	db *bolt.DB
}

type boltRecord struct {
	key, value []byte
}

// NewBolt creates a BoltDB-based implementation of storage.Storage
func NewBolt(path string) (*Bolt, error) {
	if path == "" {
		return nil, errors.New("path is required")
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, err
	}

	return &Bolt{db: db}, nil
}

// WriteRecord writes the given record to the storage
// TODO: don't overwrite existing records, store unique seq numbers
func (b *Bolt) WriteRecord(ctx context.Context, record dht.BEP44Record) error {
	ctx, span := telemetry.GetTracer().Start(ctx, "bolt.WriteRecord")
	defer span.End()

	encoded := encodeRecord(record)
	recordBytes, err := json.Marshal(encoded)
	if err != nil {
		return err
	}

	return b.write(ctx, dhtNamespace, record.ID(), recordBytes)
}

// ReadRecord reads the record with the given id from the storage
func (b *Bolt) ReadRecord(ctx context.Context, id string) (*dht.BEP44Record, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "bolt.ReadRecord")
	defer span.End()

	recordBytes, err := b.read(ctx, dhtNamespace, id)
	if err != nil {
		return nil, err
	}
	if len(recordBytes) == 0 {
		return nil, nil
	}

	var b64record base64BEP44Record
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
func (b *Bolt) ListRecords(ctx context.Context, nextPageToken []byte, pageSize int) ([]dht.BEP44Record, []byte, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "bolt.ListRecords")
	defer span.End()

	boltRecords, err := b.readSeveral(ctx, dhtNamespace, nextPageToken, pageSize)
	if err != nil {
		return nil, nil, err
	}

	var records []dht.BEP44Record
	for _, recordBytes := range boltRecords {
		var encodedRecord base64BEP44Record
		if err = json.Unmarshal(recordBytes.value, &encodedRecord); err != nil {
			return nil, nil, err
		}

		record, err := encodedRecord.Decode()
		if err != nil {
			return nil, nil, err
		}

		records = append(records, *record)
	}

	if len(boltRecords) == pageSize {
		nextPageToken = boltRecords[len(boltRecords)-1].key
	} else {
		nextPageToken = nil
	}

	return records, nextPageToken, nil
}

func (b *Bolt) Close() error {
	return b.db.Close()
}

func (b *Bolt) write(ctx context.Context, namespace string, key string, value []byte) error {
	_, span := telemetry.GetTracer().Start(ctx, "bolt.write")
	defer span.End()

	return b.db.Update(func(tx *bolt.Tx) error {
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

func (b *Bolt) read(ctx context.Context, namespace, key string) ([]byte, error) {
	_, span := telemetry.GetTracer().Start(ctx, "bolt.read")
	defer span.End()

	var result []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.WithContext(ctx).WithField("namespace", namespace).Info("namespace does not exist")
			return nil
		}
		result = bucket.Get([]byte(key))
		return nil
	})
	return result, err
}

func (b *Bolt) readAll(namespace string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := b.db.View(func(tx *bolt.Tx) error {
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

func (b *Bolt) readSeveral(ctx context.Context, namespace string, after []byte, count int) ([]boltRecord, error) {
	_, span := telemetry.GetTracer().Start(ctx, "bolt.readSeveral")
	defer span.End()

	var result []boltRecord
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.WithContext(ctx).WithField("namespace", namespace).Warn("namespace does not exist")
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

// RecordCount returns the number of records in the storage for the mainline namespace
func (b *Bolt) RecordCount(ctx context.Context) (int, error) {
	_, span := telemetry.GetTracer().Start(ctx, "bolt.RecordCount")
	defer span.End()

	var count int
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dhtNamespace))
		if bucket == nil {
			logrus.WithContext(ctx).WithField("namespace", dhtNamespace).Warn("namespace does not exist")
			return nil
		}
		count = bucket.Stats().KeyN
		return nil
	})
	return count, err
}

func (b *Bolt) WriteFailedRecord(ctx context.Context, id string) error {
	_, span := telemetry.GetTracer().Start(ctx, "bolt.WriteFailedRecord")
	defer span.End()

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(failedNamespace))
		if err != nil {
			return err
		}

		count := 1
		v := bucket.Get([]byte(id))
		if v != nil {
			if err = json.Unmarshal(v, &count); err != nil {
				return err
			}
			count++
		}

		buf := new(bytes.Buffer)
		if err = binary.Write(buf, binary.LittleEndian, count); err != nil {
			return err
		}
		return bucket.Put([]byte(id), buf.Bytes())
	})
}

func (b *Bolt) ListFailedRecords(ctx context.Context) ([]dht.FailedRecord, error) {
	_, span := telemetry.GetTracer().Start(ctx, "bolt.ListFailedRecords")
	defer span.End()

	var result []dht.FailedRecord
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(failedNamespace))
		if bucket == nil {
			logrus.WithField("namespace", failedNamespace).Warn("namespace does not exist")
			return nil
		}

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var count int
			if err := binary.Read(bytes.NewReader(v), binary.LittleEndian, &count); err != nil {
				return err
			}
			result = append(result, dht.FailedRecord{ID: string(k), Count: count})
		}
		return nil
	})
	return result, err
}

func (b *Bolt) FailedRecordCount(ctx context.Context) (int, error) {
	_, span := telemetry.GetTracer().Start(ctx, "bolt.FailedRecordCount")
	defer span.End()

	var count int
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(failedNamespace))
		if bucket == nil {
			logrus.WithField("namespace", failedNamespace).Warn("namespace does not exist")
			return nil
		}
		count = bucket.Stats().KeyN
		return nil
	})
	return count, err
}
