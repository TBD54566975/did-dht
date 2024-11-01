-- name: WriteRecord :exec
INSERT INTO dht_records(key, value, sig, seq) VALUES(?, ?, ?, ?);

-- name: ReadRecord :one
SELECT * FROM dht_records WHERE key = ? LIMIT 1;

-- name: ListRecords :many
SELECT * FROM dht_records WHERE id > (SELECT id FROM dht_records WHERE dht_records.key = ?) ORDER BY id ASC LIMIT ?;

-- name: ListRecordsFirstPage :many
SELECT * FROM dht_records ORDER BY id ASC LIMIT ?;

-- name: RecordCount :one
SELECT count(*) AS exact_count FROM dht_records;

-- name: WriteFailedRecord :exec
INSERT INTO failed_records(id, failure_count)
VALUES(?, ?)
ON CONFLICT (id) DO UPDATE SET failure_count = failed_records.failure_count + 1;

-- name: ListFailedRecords :many
SELECT * FROM failed_records;

-- name: FailedRecordCount :one
SELECT count(*) AS exact_count FROM failed_records;