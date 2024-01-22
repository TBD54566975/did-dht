-- +goose Up
CREATE TABLE pkarr_records (
    key VARCHAR(43) PRIMARY KEY NOT NULL, -- VARCHAR(43) holds 32 bytes base64-encoded
    value VARCHAR(1334) NOT NULL, -- VARCHAR(1334) holds 1000 bytes base64-encoded
    sig VARCHAR(86) NOT NULL, -- VARCHAR(86) holds 64 bytes base64-encoded
    seq BIGINT NOT NULL
);

-- +goose Down
DROP TABLE pkarr_records;