package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db *bolt.DB
}

func NewStorage(path string) (*Storage, error) {
	if path == "" {
		return nil, errors.New("path is required")
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

// URI return filepath of boltDB,
func (s *Storage) URI() string {
	return s.db.Path()
}

// IsOpen return if db was opened
func (s *Storage) IsOpen() bool {
	if s.db == nil {
		return false
	}
	return s.db.Path() != ""
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Exists(_ context.Context, namespace, key string) (bool, error) {
	exists := true
	var result []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			exists = false
			return nil
		}
		result = bucket.Get([]byte(key))
		return nil
	})

	if result == nil {
		exists = false
	}

	return exists, err
}

func (s *Storage) Write(namespace string, key string, value []byte) error {
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

func (s *Storage) Read(namespace, key string) ([]byte, error) {
	var result []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.Infof("namespace<%s> does not exist", namespace)
			return nil
		}
		result = bucket.Get([]byte(key))
		return nil
	})
	return result, err
}

// ReadPrefix does a prefix query within a dhtNamespace.
func (s *Storage) ReadPrefix(namespace, prefix string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			errMsg := fmt.Sprintf("namespace<%s> does not exist", namespace)
			logrus.Error(errMsg)
			return errors.New(errMsg)
		}
		cursor := bucket.Cursor()
		for k, v := cursor.Seek([]byte(prefix)); k != nil && bytes.HasPrefix(k, []byte(prefix)); k, v = cursor.Next() {
			result[string(k)] = v
		}
		return nil
	})
	return result, err
}

func (s *Storage) ReadAll(namespace string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.Warnf("namespace<%s> does not exist", namespace)
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

func (s *Storage) ReadAllKeys(_ context.Context, namespace string) ([]string, error) {
	var result []string
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			logrus.Warnf("namespace<%s> does not exist", namespace)
			return nil
		}
		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			result = append(result, string(k))
		}
		return nil
	})
	return result, err
}

func (s *Storage) Update(namespace string, key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			return fmt.Errorf("namespace<%s> does not exist", namespace)
		}
		if bucket.Get([]byte(key)) == nil {
			return fmt.Errorf("key<%s> does not exist", key)
		}
		if err := bucket.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})
}

func (s *Storage) Delete(namespace, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			return fmt.Errorf("namespace<%s> does not exist", namespace)
		}
		return bucket.Delete([]byte(key))
	})
}

func (s *Storage) DeleteNamespace(namespace string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(namespace)); err != nil {
			return errors.Wrapf(err, "could not delete namespace<%s>, n", namespace)
		}
		return nil
	})
}
