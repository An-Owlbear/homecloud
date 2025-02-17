package persistence

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
	"os"
)

// SetupDB sets up the database for the given path, running migrations if required
func SetupDB(dbPath string, migrations embed.FS) (db *sql.DB, err error) {
	if _, err = os.Stat(dbPath); os.IsNotExist(err) {
		_, err = os.Create(dbPath)
		if err != nil {
			return
		}
	}

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return
	}

	goose.SetBaseFS(migrations)
	if err = goose.SetDialect("sqlite3"); err != nil {
		return
	}

	if err = goose.Up(db, "migrations"); err != nil {
		return
	}

	return
}
