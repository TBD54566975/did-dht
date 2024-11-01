-- +goose Up
CREATE TABLE dht_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key BLOB UNIQUE NOT NULL,
    value BLOB NOT NULL,
    sig BLOB NOT NULL,
    seq INTEGER NOT NULL
);

CREATE TABLE failed_records (
    id BLOB PRIMARY KEY,
    failure_count INTEGER NOT NULL
);

-- +goose Down
DROP TABLE failed_records;
DROP TABLE dht_records;