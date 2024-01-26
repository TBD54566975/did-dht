-- +goose Up
CREATE TABLE pkarr_records (
    key BYTEA UNIQUE NOT NULL,
    value BYTEA NOT NULL,
    sig BYTEA NOT NULL,
    seq BIGINT NOT NULL
);

-- +goose Down
DROP TABLE pkarr_records;