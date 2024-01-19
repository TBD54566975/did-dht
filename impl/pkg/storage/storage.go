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
	WriteRecord(ctx context.Context, record pkarr.PkarrRecord) error
	ReadRecord(ctx context.Context, id string) (*pkarr.PkarrRecord, error)
	ListRecords(ctx context.Context) ([]pkarr.PkarrRecord, error)
}

func NewStorage(uri string) (Storage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "bolt":
		filename := u.Host
		if u.Path != "" {
			filename = fmt.Sprintf("%s/%s", filename, u.Path)
		}
		return bolt.NewStorage(filename)
	case "postgres":
		return postgres.NewPostgres(uri)
	default:
		return nil, fmt.Errorf("unsupported db type %s", u.Scheme)
	}
}
