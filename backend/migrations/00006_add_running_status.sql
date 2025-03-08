-- +goose Up
-- +goose StatementBegin
ALTER TABLE apps ADD COLUMN status TEXT NOT NULL default 'running';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE apps DROP COLUMN status;
-- +goose StatementEnd
