package server

import (
	"database/sql"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/api"
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"
	"os"
)

func CreateServer() *echo.Echo {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	// Loads the configuration
	serverConfig, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

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
			URL: serverConfig.Ory.Hydra.PrivateAddress.String(),
		},
	}
	hydraAdmin := hydra.NewAPIClient(hydraAdminConfig)

	// Sets up ory kratos client
	kratosConfig := kratos.NewConfiguration()
	kratosConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: serverConfig.Ory.Kratos.PrivateAddress.String(),
		},
	}
	kratosClient := kratos.NewAPIClient(kratosConfig)

	// Sets up ory kratos admin client
	kratosAdminConfig := kratos.NewConfiguration()
	kratosAdminConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: serverConfig.Ory.KratosAdmin.PrivateAddress.String(),
		},
	}
	kratosAdmin := kratos.NewAPIClient(kratosAdminConfig)

	// Sets up hosts config
	hostsMap := apps.HostsMap{}
	hosts := apps.NewHosts(hostsMap, serverConfig.Host)

	// Sets up proxies for installed apps
	err = apps.SetupProxies(dockerClient, queries, hosts)
	if err != nil {
		panic(err)
	}

	// Sets up backend API
	backendApi := echo.New()
	backendApi.Use(config.ContextMiddleware)
	backendApi.Use(auth.Middleware(kratosClient.FrontendAPI))
	api.AddRoutes(
		backendApi,
		dockerClient,
		queries,
		storeClient,
		hosts,
		hydraAdmin,
		kratosClient,
		kratosAdmin.IdentityAPI,
		*serverConfig,
	)
	hostsMap[fmt.Sprintf("%s:%d", serverConfig.Host.Host, serverConfig.Host.Port)] = backendApi

	// Adds reverse proxies for ory services
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
