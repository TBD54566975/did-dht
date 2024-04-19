-- +goose Up
CREATE TABLE dht_records (
    id SERIAL PRIMARY KEY,
    key BYTEA UNIQUE NOT NULL,
    value BYTEA NOT NULL,
    sig BYTEA NOT NULL,
    seq BIGINT NOT NULL
);

-- +goose Down
DROP TABLE dht_records;