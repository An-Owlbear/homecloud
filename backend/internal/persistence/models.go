// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package persistence

import (
	"database/sql"
	"time"
)

type App struct {
	ID           string
	Schema       []byte
	DateAdded    int64
	ClientID     sql.NullString
	ClientSecret sql.NullString
}

type InviteCode struct {
	Code       string
	ExpiryDate time.Time
}
