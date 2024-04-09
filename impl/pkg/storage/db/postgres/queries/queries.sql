-- name: WriteRecord :exec
INSERT INTO pkarr_records(key, value, sig, seq) VALUES($1, $2, $3, $4);

-- name: ReadRecord :one
SELECT * FROM pkarr_records WHERE key = $1 LIMIT 1;

-- name: ListRecords :many
SELECT * FROM pkarr_records WHERE id > (SELECT id FROM pkarr_records WHERE pkarr_records.key = $1) ORDER BY id ASC LIMIT $2;

-- name: ListRecordsFirstPage :many
SELECT * FROM pkarr_records ORDER BY id ASC LIMIT $1;

-- name: RecordCount :one
SELECT count(*) AS exact_count FROM pkarr_records;
