-- name: CreateApp :exec
INSERT INTO apps (id, schema, date_added)
VALUES (?, jsonb(?), unixepoch());

-- name: RemoveApp :exec
DELETE FROM apps where id = ?;
