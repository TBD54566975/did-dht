package storage

import (
	"context"
	"fmt"
	"net/url"

	"github.com/TBD54566975/did-dht-method/pkg/storage/db/bolt"
	"github.com/TBD54566975/did-dht-method/pkg/storage/db/postgres"
	"github.com/TBD54566975/did-dht-method/pkg/storage/pkarr"
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
		return bolt.NewBolt(filename)
	case "postgres":
		return postgres.NewPostgres(uri)
	default:
		return nil, fmt.Errorf("unsupported db type %s (from uri %s)", u.Scheme, uri)
	}
}
