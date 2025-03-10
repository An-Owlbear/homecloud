package testutils

import (
	"database/sql"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/util"
	"github.com/pressly/goose/v3"
	"os"
	"path/filepath"
	"testing"
)

var db *sql.DB
var dbPath = filepath.Join(util.RootDir(), "tmp/test.db")

func SetupDB(t *testing.T) *persistence.Queries {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Unexpected error setting up DB: %s", err.Error())
	}

	goose.SetBaseFS(os.DirFS(util.RootDir()))
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("Unexpected error setting up goose: %s", err.Error())
	}

	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("Unexpected error apply DB migrations: %s", err.Error())
	}

	return persistence.New(db)
}

func CleanupDB() {
	db.Close()
	os.Remove(dbPath)
}
