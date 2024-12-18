package server

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/An-Owlbear/homecloud/backend/internal/api"
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func CreateServer() *echo.Echo {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer docker.Close()

	db, err := sql.Open("sqlite3", "file:tmp/data.db")
	if err != nil {
		panic(err)
	}

	queries := persistence.New(db)
	storeClient := apps.NewStoreClient(os.Getenv("STORE_URL"))

	e := echo.New()
	api.AddRoutes(e, docker, queries, storeClient)

	return e
}
