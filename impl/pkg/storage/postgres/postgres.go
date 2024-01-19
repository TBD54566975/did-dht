package postgres

import (
	"database/sql"
	"embed"

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
