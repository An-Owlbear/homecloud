-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_options(
    user_id TEXT PRIMARY KEY,
    completed_welcome BOOLEAN NOT NULL DEFAULT false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_options;
-- +goose StatementEnd
