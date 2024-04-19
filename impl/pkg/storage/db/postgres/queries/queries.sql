-- name: WriteRecord :exec
INSERT INTO dht_records(key, value, sig, seq) VALUES($1, $2, $3, $4);

-- name: ReadRecord :one
SELECT * FROM dht_records WHERE key = $1 LIMIT 1;

-- name: ListRecords :many
SELECT * FROM dht_records WHERE id > (SELECT id FROM dht_records WHERE dht_records.key = $1) ORDER BY id ASC LIMIT $2;

-- name: ListRecordsFirstPage :many
SELECT * FROM dht_records ORDER BY id ASC LIMIT $1;

-- name: RecordCount :one
SELECT count(*) AS exact_count FROM dht_records;
