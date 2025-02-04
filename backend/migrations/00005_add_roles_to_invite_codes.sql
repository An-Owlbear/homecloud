-- +goose Up
-- +goose StatementBegin
ALTER TABLE invite_codes ADD COLUMN roles BLOB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE invite_codes DROP COLUMN roles;
-- +goose StatementEnd
