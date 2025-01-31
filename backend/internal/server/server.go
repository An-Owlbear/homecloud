package server

import (
	"database/sql"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	kratos "github.com/ory/kratos-client-go"
	"os"
	"strconv"

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

	// Loads the configuration
	hostPort, err := strconv.Atoi(os.Getenv("HOMECLOUD_PORT"))
	if err != nil {
		panic(err)
	}
	hostConfig := config.NewHost(os.Getenv("HOMECLOUD_HOST"), hostPort)

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

	// Sets up ory kratos client
	kratosConfig := kratos.NewConfiguration()
	kratosConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: "http://kratos:4433",
		},
	}
	kratosClient := kratos.NewAPIClient(kratosConfig)

	// Sets up ory kratos admin client
	kratosAdminConfig := kratos.NewConfiguration()
	kratosAdminConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: "http://kratos:4434",
		},
	}
	kratosAdmin := kratos.NewAPIClient(kratosAdminConfig)

	// Setups of hard coded proxies
	hostsMap := apps.HostsMap{}
	hosts := apps.NewHosts(hostsMap, hostConfig)
	backendApi := echo.New()
	api.AddRoutes(
		backendApi,
		dockerClient,
		queries,
		storeClient,
		hosts,
		hydraAdmin,
		kratosClient,
		kratosAdmin.IdentityAPI,
		hostConfig,
	)
	hostsMap[fmt.Sprintf("%s:%d", hostConfig.Host, hostConfig.Port)] = backendApi
	hosts.AddProxy("hydra", "hydra", "4444")
	hosts.AddProxy("login", "kratos-selfservice-ui-node", "4455")
	hosts.AddProxy("kratos", "kratos", "4433")

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
		if host, ok := hostsMap[c.Request().Host]; ok {
			host.ServeHTTP(c.Response(), c.Request())
		} else {
			err = echo.ErrNotFound
		}

		return
	})

	return e
}
