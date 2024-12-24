-- +goose Up
-- +goose StatementBegin
CREATE TABLE apps (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	schema BLOB NOT NULL,
	date_added INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apps;
-- +goose StatementEnd
