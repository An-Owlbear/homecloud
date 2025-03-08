-- name: CreateApp :exec
INSERT INTO apps (id, schema, date_added, client_id, client_secret, status)
VALUES (sqlc.arg(id), jsonb(sqlc.arg(schema)), unixepoch(), sqlc.arg(client_id), sqlc.arg(client_secret), 'running');

-- name: RemoveApp :execresult
DELETE FROM apps where id = ?;

-- name: UpdateApp :exec
UPDATE apps SET schema = jsonb(sqlc.arg(schema))
WHERE id = sqlc.arg(id);

-- name: getAppUnparsed :one
SELECT id, json(schema) as schema, date_added, status FROM apps
WHERE id = sqlc.arg(id);

-- name: getAppsUnparsed :many
SELECT id, json(schema) as schema, date_added, status FROM apps;
