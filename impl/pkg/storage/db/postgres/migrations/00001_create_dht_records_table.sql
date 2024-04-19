-- +goose Up
CREATE TABLE dht_records (
    id SERIAL PRIMARY KEY,
    key BYTEA UNIQUE NOT NULL,
    value BYTEA NOT NULL,
    sig BYTEA NOT NULL,
    seq BIGINT NOT NULL
);

CREATE TABLE failed_records (
    id BYTEA PRIMARY KEY,
    failure_count INTEGER NOT NULL
);

-- +goose Down
DROP TABLE failed_records;
DROP TABLE dht_records;