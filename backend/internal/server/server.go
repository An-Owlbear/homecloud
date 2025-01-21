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
	hydra "github.com/ory/hydra-client-go/v2"
)

func CreateServer() *echo.Echo {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	hosts := apps.Hosts{}

	// Sets up docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()
	if err := docker.EnsureProxyNetwork(dockerClient); err != nil {
		panic(err)
	}

	// Sets of database connection
	db, err := sql.Open("sqlite3", "file:tmp/data.db")
	if err != nil {
		panic(err)
	}

	queries := persistence.New(db)
	storeClient := apps.NewStoreClient(os.Getenv("STORE_URL"))

	// Sets up ory hydra client
	hydraAdminConfig := hydra.NewConfiguration()
	hydraAdminConfig.Servers = []hydra.ServerConfiguration{
		{
			URL: "http://hydra:4445",
		},
	}
	hydraAdmin := hydra.NewAPIClient(hydraAdminConfig)

	// Setups of hard coded proxies
	backendApi := echo.New()
	api.AddRoutes(backendApi, dockerClient, queries, storeClient, hosts, hydraAdmin)
	hosts[fmt.Sprintf("%s:%s", os.Getenv("HOMECLOUD_HOST"), os.Getenv("HOMECLOUD_PORT"))] = backendApi
	apps.AddProxy(hosts, "hydra", "hydra", "4444")
	apps.AddProxy(hosts, "login", "kratos-selfservice-ui-node", "4455")
	apps.AddProxy(hosts, "kratos", "kratos", "4433")

	// Sets up global logging
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

	// Checks which HTTP server/proxy to send traffic to
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
