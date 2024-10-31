package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"github.com/tv42/zbase32"

	"github.com/TBD54566975/did-dht/pkg/dht"
	"github.com/TBD54566975/did-dht/pkg/telemetry"
)

//go:embed migrations
var migrations embed.FS

type SQLite string

// NewSQLite creates a SQLite-based implementation of storage.Storage
func NewSQLite(uri string) (SQLite, error) {
	db := SQLite(uri)
	if err := db.migrate(); err != nil {
		return db, fmt.Errorf("error migrating sqlite database: %v", err)
	}

	return db, nil
}

func (s SQLite) migrate() error {
	db, err := sql.Open("sqlite3", string(s))
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations)
	if err = goose.SetDialect("sqlite"); err != nil {
		return err
	}

	if err = goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}

func (s SQLite) connect(ctx context.Context) (*Queries, *sql.DB, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "sqlite.connect")
	defer span.End()

	db, err := sql.Open("sqlite3", string(s))
	if err != nil {
		return nil, nil, err
	}

	return New(db), db, nil
}

func (s SQLite) WriteRecord(ctx context.Context, record dht.BEP44Record) error {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.WriteRecord")
	defer span.End()

	queries, db, err := s.connect(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

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

func (s SQLite) ReadRecord(ctx context.Context, id string) (*dht.BEP44Record, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.ReadRecord")
	defer span.End()

	queries, db, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	decodedID, err := zbase32.DecodeString(id)
	if err != nil {
		return nil, err
	}
	row, err := queries.ReadRecord(ctx, decodedID)
	if err != nil {
		return nil, err
	}

	record, err := row.Record()
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s SQLite) ListRecords(ctx context.Context, nextPageToken []byte, limit int) ([]dht.BEP44Record, []byte, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.ListRecords")
	defer span.End()

	queries, db, err := s.connect(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	var rows []DhtRecord
	if nextPageToken == nil {
		rows, err = queries.ListRecordsFirstPage(ctx, int64(limit))
	} else {
		rows, err = queries.ListRecords(ctx, ListRecordsParams{
			Key:   nextPageToken,
			Limit: int64(limit),
		})
	}
	if err != nil {
		return nil, nil, err
	}

	var records []dht.BEP44Record
	for _, row := range rows {
		record, err := dht.NewBEP44Record(row.Key, row.Value, row.Sig, row.Seq)
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

func (row DhtRecord) Record() (*dht.BEP44Record, error) {
	return dht.NewBEP44Record(row.Key, row.Value, row.Sig, row.Seq)
}

func (s SQLite) RecordCount(ctx context.Context) (int, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.RecordCount")
	defer span.End()

	queries, db, err := s.connect(ctx)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	count, err := queries.RecordCount(ctx)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (s SQLite) WriteFailedRecord(ctx context.Context, id string) error {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.WriteFailedRecord")
	defer span.End()

	queries, db, err := s.connect(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	err = queries.WriteFailedRecord(ctx, WriteFailedRecordParams{
		ID:           []byte(id),
		FailureCount: 1,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s SQLite) ListFailedRecords(ctx context.Context) ([]dht.FailedRecord, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.ListFailedRecords")
	defer span.End()

	queries, db, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := queries.ListFailedRecords(ctx)
	if err != nil {
		return nil, err
	}

	var failedRecords []dht.FailedRecord
	for _, row := range rows {
		failedRecords = append(failedRecords, dht.FailedRecord{
			ID:    string(row.ID),
			Count: int(row.FailureCount),
		})
	}

	return failedRecords, nil
}

func (s SQLite) FailedRecordCount(ctx context.Context) (int, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "postgres.FailedRecordCount")
	defer span.End()

	queries, db, err := s.connect(ctx)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	count, err := queries.FailedRecordCount(ctx)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (s SQLite) Close() error {
	// no-op, sqlite connection is closed after each request
	return nil
}
