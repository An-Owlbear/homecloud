-- name: CreateApp :exec
INSERT INTO apps (id, schema, date_added)
VALUES (sqlc.arg(id), jsonb(sqlc.arg(schema)), unixepoch());

-- name: RemoveApp :execresult
DELETE FROM apps where id = ?;

-- name: UpdateApp :exec
UPDATE apps SET schema = jsonb(sqlc.arg(schema))
WHERE id = sqlc.arg(id);
