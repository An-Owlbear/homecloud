-- +goose Up
-- +goose StatementBegin
CREATE TABLE invite_codes (
    code TEXT PRIMARY KEY,
    expiry_date DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE invite_codes;
-- +goose StatementEnd
