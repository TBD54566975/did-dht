package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht/pkg/dht"
	"github.com/TBD54566975/did-dht/pkg/storage/db/postgres"
	"github.com/TBD54566975/did-dht/pkg/storage/db/sqlite"
)

type Storage interface {
	WriteRecord(ctx context.Context, record dht.BEP44Record) error
	ReadRecord(ctx context.Context, id string) (*dht.BEP44Record, error)
	ListRecords(ctx context.Context, nextPageToken []byte, pageSize int) (records []dht.BEP44Record, nextPage []byte, err error)
	RecordCount(ctx context.Context) (int, error)

	WriteFailedRecord(ctx context.Context, id string) error
	ListFailedRecords(ctx context.Context) ([]dht.FailedRecord, error)
	FailedRecordCount(ctx context.Context) (int, error)

	Close() error
}

func NewStorage(uri string) (Storage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "sqlite", "":
		filename := u.Host
		if u.Path != "" {
			filename = fmt.Sprintf("%s/%s", filename, u.Path)
		}
		logrus.WithField("file", filename).Info("using sqlite for storage")
		return sqlite.NewSQLite(filename)

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
