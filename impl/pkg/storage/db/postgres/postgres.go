package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/TBD54566975/did-dht-method/pkg/storage/pkarr"
	pgx "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
)

//go:embed migrations
var migrations embed.FS

type postgres string

// NewPostgres creates a PostgresQL-based implementation of storage.Storage
func NewPostgres(uri string) (postgres, error) {
	db := postgres(uri)
	if err := db.migrate(); err != nil {
		return db, fmt.Errorf("error migrating postgres database: %v", err)
	}

	return db, nil
}

func (p postgres) migrate() error {
	db, err := sql.Open("pgx/v5", string(p))
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

func (p postgres) connect(ctx context.Context) (*Queries, *pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, string(p))
	if err != nil {
		return nil, nil, err
	}

	return New(conn), conn, nil
}

func (p postgres) WriteRecord(ctx context.Context, record pkarr.Record) error {
	queries, db, err := p.connect(ctx)
	if err != nil {
		return err
	}
	defer db.Close(ctx)

	err = queries.WriteRecord(ctx, WriteRecordParams{
		Key:   record.K,
		Value: record.V,
		Sig:   record.Sig,
		Seq:   record.Seq,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p postgres) ReadRecord(ctx context.Context, id []byte) (*pkarr.Record, error) {
	queries, db, err := p.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close(ctx)

	record, err := queries.ReadRecord(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pkarr.Record{
		K:   record.Key,
		V:   record.Value,
		Sig: record.Sig,
		Seq: record.Seq,
	}, nil
}

func (p postgres) ListAllRecords(ctx context.Context) ([]pkarr.Record, error) {
	queries, db, err := p.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close(ctx)

	rows, err := queries.ListAllRecords(ctx)
	if err != nil {
		return nil, err
	}

	var records []pkarr.Record
	for _, row := range rows {
		records = append(records, pkarr.Record{
			K:   row.Key,
			V:   row.Value,
			Sig: row.Sig,
			Seq: row.Seq,
		})
	}

	return records, nil
}

func (p postgres) ListRecords(ctx context.Context, nextPageToken []byte, limit int) ([]pkarr.Record, []byte, error) {
	queries, db, err := p.connect(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close(ctx)

	var rows []PkarrRecord
	if nextPageToken == nil {
		rows, err = queries.ListRecordsFirstPage(ctx, int32(limit))
	} else {
		rows, err = queries.ListRecords(ctx, ListRecordsParams{
			Key:   nextPageToken,
			Limit: int32(limit),
		})
	}
	if err != nil {
		return nil, nil, err
	}

	var records []pkarr.Record
	for _, row := range rows {
		records = append(records, pkarr.Record{
			K:   row.Key,
			V:   row.Value,
			Sig: row.Sig,
			Seq: row.Seq,
		})
	}

	if len(rows) == limit {
		nextPageToken = rows[len(rows)-1].Key
	} else {
		nextPageToken = nil
	}

	return records, nextPageToken, nil
}

func (p postgres) Close() error {
	// no-op, postgres connection is closed after each request
	return nil
}
