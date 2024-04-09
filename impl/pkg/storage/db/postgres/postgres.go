package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/TBD54566975/did-dht-method/pkg/telemetry"
)

//go:embed migrations
var migrations embed.FS

type Postgres string

// NewPostgres creates a PostgresQL-based implementation of storage.Storage
func NewPostgres(uri string) (Postgres, error) {
	db := Postgres(uri)
	if err := db.migrate(); err != nil {
		return db, fmt.Errorf("error migrating postgres database: %v", err)
	}

	return db, nil
}

func (p Postgres) migrate() error {
	db, err := sql.Open("pgx/v5", string(p))
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations)
	if err = goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err = goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}

func (p Postgres) connect(ctx context.Context) (*Queries, *pgx.Conn, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.connect")
	defer span.End()

	conn, err := pgx.Connect(ctx, string(p))
	if err != nil {
		return nil, nil, err
	}

	return New(conn), conn, nil
}

func (p Postgres) WriteRecord(ctx context.Context, record pkarr.Record) error {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.WriteRecord")
	defer span.End()

	queries, db, err := p.connect(ctx)
	if err != nil {
		return err
	}
	defer db.Close(ctx)

	err = queries.WriteRecord(ctx, WriteRecordParams{
		Key:   record.Key[:],
		Value: record.Value[:],
		Sig:   record.Signature[:],
		Seq:   record.SequenceNumber,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p Postgres) ReadRecord(ctx context.Context, id []byte) (*pkarr.Record, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.ReadRecord")
	defer span.End()

	queries, db, err := p.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close(ctx)

	row, err := queries.ReadRecord(ctx, id)
	if err != nil {
		return nil, err
	}

	record, err := row.Record()
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (p Postgres) ListRecords(ctx context.Context, nextPageToken []byte, limit int) ([]pkarr.Record, []byte, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.ListRecords")
	defer span.End()

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
		record, err := pkarr.NewRecord(row.Key, row.Value, row.Sig, row.Seq)
		if err != nil {
			// TODO: do something useful if this happens
			logrus.WithContext(ctx).WithError(err).WithField("record_id", row.ID).Warn("error loading record from database, skipping")
			continue
		}

		records = append(records, *record)
	}

	if len(rows) == limit {
		nextPageToken = rows[len(rows)-1].Key
	} else {
		nextPageToken = nil
	}

	return records, nextPageToken, nil
}

func (p Postgres) Close() error {
	// no-op, postgres connection is closed after each request
	return nil
}

func (row PkarrRecord) Record() (*pkarr.Record, error) {
	return pkarr.NewRecord(row.Key, row.Value, row.Sig, row.Seq)
}

func (p Postgres) RecordCount(ctx context.Context) (int, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.RecordCount")
	defer span.End()

	queries, db, err := p.connect(ctx)
	if err != nil {
		return 0, err
	}
	defer db.Close(ctx)

	count, err := queries.RecordCount(ctx)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}
