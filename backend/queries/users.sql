-- name: AddUser :exec
INSERT INTO user_options (user_id, completed_welcome)
VALUES (sqlc.arg(user_id), false);

-- name: GetUserOptions :one
SELECT * FROM user_options WHERE user_id = sqlc.arg(user_id);