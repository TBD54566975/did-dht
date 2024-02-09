package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/TBD54566975/did-dht-method/pkg/storage/db/bolt"
	"github.com/TBD54566975/did-dht-method/pkg/storage/db/postgres"
)

type Storage interface {
	WriteRecord(ctx context.Context, record pkarr.Record) error
	ReadRecord(ctx context.Context, id []byte) (*pkarr.Record, error)
	ListRecords(ctx context.Context, nextPageToken []byte, pagesize int) (records []pkarr.Record, nextPage []byte, err error)
	ListAllRecords(ctx context.Context) ([]pkarr.Record, error)
	Close() error
}

func NewStorage(uri string) (Storage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "bolt", "":
		filename := u.Host
		if u.Path != "" {
			filename = fmt.Sprintf("%s/%s", filename, u.Path)
		}
		logrus.WithField("file", filename).Info("using boltdb for storage")
		return bolt.NewBolt(filename)
	case "postgres":
		logrus.WithFields(logrus.Fields{
			"host":     u.Host,
			"database": strings.TrimPrefix(u.Path, "/"),
		}).Info("using postgres for storage")
		return postgres.NewPostgres(uri)
	default:
		return nil, fmt.Errorf("unsupported db type %s (from uri %s)", u.Scheme, uri)
	}
}
