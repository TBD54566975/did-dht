-- name: WriteRecord :exec
INSERT INTO pkarr_records(key, value, sig, seq) VALUES($1, $2, $3, $4);

-- name: ReadRecord :one
SELECT * FROM pkarr_records WHERE key = $1 LIMIT 1;

-- name: ListRecords :many
SELECT * FROM pkarr_records;