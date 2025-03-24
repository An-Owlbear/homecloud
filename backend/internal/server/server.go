package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"regexp"

	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"
	"golang.org/x/crypto/acme/autocert"

	"github.com/An-Owlbear/homecloud/backend"
	"github.com/An-Owlbear/homecloud/backend/internal/api"
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/storage"
)

func CreateServer() {
	// Loads configuration
	if os.Getenv("ENVIRONMENT") == "DEV" {
		if err := godotenv.Load(".dev.env"); err != nil {
			panic(err)
		}
	}
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

	// Sets of database connection
	db, err := persistence.SetupDB("db/data.db", backend.Migrations)
	if err != nil {
		panic(err)
	}

	queries := persistence.New(db)
	storeClient := apps.NewStoreClient(serverConfig.Store)
	err = storeClient.UpdatePackageList(context.Background(), queries)
	if err != nil {
		panic(err)
	}

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

	// Sets up data storage handling
	appDataHandler := storage.NewAppDataHandler(serverConfig.Storage, serverConfig.Store)

	// Sets up hosts config
	hostsMap := apps.HostsMap{}
	hosts := apps.NewHosts(hostsMap, serverConfig.Host)

	// Sets up proxies for installed apps
	err = apps.SetupProxies(dockerClient, queries, hosts, appDataHandler, serverConfig.Host, serverConfig.Ory)
	if err != nil {
		panic(err)
	}

	// Sets up proxy for launcher on host
	launcherUrl, err := url.Parse(serverConfig.Launcher.Url)
	if err != nil {
		panic(err)
	}
	launcherTargets := []*middleware.ProxyTarget{{URL: launcherUrl}}
	launcherProxy := middleware.Proxy(middleware.NewRoundRobinBalancer(launcherTargets))

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
		appDataHandler,
		*serverConfig,
		launcherProxy,
	)
	hostname := serverConfig.Host.Host
	if serverConfig.Host.Port != 80 && serverConfig.Host.Port != 443 {
		hostname = fmt.Sprintf("%s:%d", serverConfig.Host.Host, serverConfig.Host.Port)
	}
	hostsMap[hostname] = backendApi

	// Adds reverse proxies for ory services
	hosts.AddProxy("hydra", "hydra", "4444")
	hosts.AddProxy("login", "kratos-selfservice-ui-node", "4455")
	hosts.AddProxy("kratos", "kratos", "4433")

	// Sets up global logging
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogHost:     true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("REQUEST URI: %s%s, status: %v\n", v.Host, v.URI, v.Status)
			if v.Error != nil {
				fmt.Printf("Error: %s\n", v.Error.Error())
			}
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

	if serverConfig.Host.HTTPS {
		slog.Info("Starting server with HTTPS")
		allowedHosts := regexp.MustCompile(fmt.Sprintf("^(?:.*\\.)?%s$", regexp.QuoteMeta(serverConfig.Host.Host)))
		e.AutoTLSManager.HostPolicy = func(ctx context.Context, host string) error {
			if matches := allowedHosts.MatchString(host); !matches {
				return fmt.Errorf("invalid host: %s", host)
			}
			return nil
		}
		e.AutoTLSManager.Cache = autocert.DirCache("data/.cache")
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		slog.Info("Starting server without HTTPS - THIS IS NOT SAFE FOR PRODUCTION ENVIRONMENTS")
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", serverConfig.Host.Port)))
	}
}
