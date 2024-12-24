-- name: CreateApp :exec
INSERT INTO apps (id, schema, date_added)
VALUES (sqlc.arg(id), jsonb(sqlc.arg(schema)), unixepoch());

-- name: RemoveApp :execresult
DELETE FROM apps where id = ?;
