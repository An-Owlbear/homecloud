package server

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/An-Owlbear/homecloud/backend/internal/api"
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func CreateServer() *echo.Echo {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	hosts := apps.Hosts{}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()
	if err := docker.EnsureProxyNetwork(dockerClient); err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", "file:tmp/data.db")
	if err != nil {
		panic(err)
	}

	queries := persistence.New(db)
	storeClient := apps.NewStoreClient(os.Getenv("STORE_URL"))

	backendApi := echo.New()
	api.AddRoutes(backendApi, dockerClient, queries, storeClient, hosts)
	hosts["home.cloud:1323"] = backendApi

	e := echo.New()
	e.Any("/*", func(c echo.Context) (err error) {
		if host, ok := hosts[c.Request().Host]; ok {
			host.ServeHTTP(c.Response(), c.Request())
		} else {
			err = echo.ErrNotFound
		}

		return
	})

	return e
}
