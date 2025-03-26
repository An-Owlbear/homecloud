-- name: AddUser :exec
INSERT INTO user_options (user_id, completed_welcome)
VALUES (sqlc.arg(user_id), false);

-- name: GetUserOptions :one
SELECT * FROM user_options WHERE user_id = sqlc.arg(user_id);

-- name: UpdateUserOptions :exec
UPDATE user_options
SET completed_welcome = coalesce(sqlc.narg(completed_welcome), completed_welcome)
WHERE user_id = sqlc.arg(user_id);