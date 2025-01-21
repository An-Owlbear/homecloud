-- +goose Up
-- +goose StatementBegin
ALTER TABLE apps ADD COLUMN client_id TEXT;
ALTER TABLE apps ADD COLUMN client_secret TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE apps DROP COLUMN client_id;
ALTER TABLE apps DROP client_secret;
-- +goose StatementEnd
