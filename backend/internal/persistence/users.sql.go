// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package persistence

import (
	"context"
	"database/sql"
)

const addUser = `-- name: AddUser :exec
INSERT INTO user_options (user_id, completed_welcome)
VALUES (?1, false)
`

func (q *Queries) AddUser(ctx context.Context, userID string) error {
	_, err := q.db.ExecContext(ctx, addUser, userID)
	return err
}

const getUserOptions = `-- name: GetUserOptions :one
SELECT user_id, completed_welcome FROM user_options WHERE user_id = ?1
`

func (q *Queries) GetUserOptions(ctx context.Context, userID string) (UserOption, error) {
	row := q.db.QueryRowContext(ctx, getUserOptions, userID)
	var i UserOption
	err := row.Scan(&i.UserID, &i.CompletedWelcome)
	return i, err
}

const updateUserOptions = `-- name: UpdateUserOptions :exec
UPDATE user_options
SET completed_welcome = coalesce(?1, completed_welcome)
WHERE user_id = ?2
`

type UpdateUserOptionsParams struct {
	CompletedWelcome sql.NullBool `json:"completed_welcome"`
	UserID           string       `json:"user_id"`
}

func (q *Queries) UpdateUserOptions(ctx context.Context, arg UpdateUserOptionsParams) error {
	_, err := q.db.ExecContext(ctx, updateUserOptions, arg.CompletedWelcome, arg.UserID)
	return err
}
