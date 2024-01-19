-- +goose Up
CREATE TABLE pkarr_records (
    id SERIAL PRIMARY KEY,
    key VARCHAR(32) NOT NULL,
    value VARCHAR(1000) NOT NULL,
    sig VARCHAR(64) NOT NULL
);

-- +goose Down
DROP TABLE pkarr_records;