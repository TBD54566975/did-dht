package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/TBD54566975/did-dht-method/pkg/storage/pkarr"
	_ "github.com/jackc/pgx/v5"
	goose "github.com/pressly/goose/v3"
)

//go:embed migrations
var migrations embed.FS

type postgres string

func NewPostgres(uri string) (postgres, error) {
	db := postgres(uri)
	if err := db.migrate(); err != nil {
		return db, err
	}

	return db, nil
}

func (p postgres) migrate() error {
	db, err := sql.Open("postgres", string(p))
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}

func (p postgres) WriteRecord(ctx context.Context, record pkarr.PkarrRecord) error {
	return errors.New("not yet implemented")
}

func (p postgres) ReadRecord(ctx context.Context, id string) (*pkarr.PkarrRecord, error) {
	return nil, errors.New("not yet implemented")
}

func (p postgres) ListRecords(ctx context.Context) ([]pkarr.PkarrRecord, error) {
	return nil, errors.New("not yet implemented")
}
