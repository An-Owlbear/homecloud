-- +goose Up
-- +goose StatementBegin
CREATE TABLE temp_apps (
	id TEXT PRIMARY KEY,
	schema BLOB NOT NULL,
	date_added INTEGER NOT NULL
);

INSERT INTO temp_apps (id, schema, date_added)
SELECT json_extract(schema, '$.id') as id, schema, date_added
FROM apps;

DROP TABLE apps;
ALTER TABLE temp_apps RENAME TO apps;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE temp_apps (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	schema BLOB NOT NULL,
	date_added INTEGER NOT NULL
);

INSERT INTO temp_apps (schema, date_added)
SELECT schema, date_added from apps;

DROP TABLE apps;
ALTER TABLE temp_apps RENAME TO apps;
-- +goose StatementEnd
