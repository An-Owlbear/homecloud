package server

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"os"

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
	hosts[fmt.Sprintf("%s:%s", os.Getenv("HOMECLOUD_HOST"), os.Getenv("HOMECLOUD_PORT"))] = backendApi
	apps.AddProxy(hosts, "hydra", "hydra", "4444")
	apps.AddProxy(hosts, "login", "kratos-selfservice-ui-node", "4455")
	apps.AddProxy(hosts, "kratos", "kratos", "4433")

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogHost:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("REQUEST URI: %s%s, status: %v\n", v.Host, v.URI, v.Status)
			return nil
		},
	}))

	e.Any("/*", func(c echo.Context) (err error) {
		c.Logger().Error(fmt.Sprintf("Hosts: %+v", hosts))

		if host, ok := hosts[c.Request().Host]; ok {
			host.ServeHTTP(c.Response(), c.Request())
		} else {
			err = echo.ErrNotFound
		}

		return
	})

	return e
}
