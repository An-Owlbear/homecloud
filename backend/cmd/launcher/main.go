package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/launcher"
)

func main() {
	// Loads development environment variables if specified
	if os.Getenv("ENVIRONMENT") == "DEV" {
		if err := godotenv.Load(".dev.env"); err != nil {
			panic(err)
		}
	}
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	deviceConfig := config.NewDeviceConfig()
	launcherEnvConfig, err := config.NewLauncher()
	if err != nil {
		panic(err)
	}

	launcherConfig, err := launcher.SetupConfig(launcherEnvConfig.ConfigFilename)
	if err != nil {
		panic(err)
	}

	storageConfig, err := config.NewStorage(true)
	if err != nil {
		panic(err)
	}

	// Setups up port configuration
	hostConfig, err := config.NewHost()
	if err != nil {
		panic(err)
	}

	// Creates docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	storeClient := apps.NewStoreClient(config.Store{StoreUrl: os.Getenv("SYSTEM_STORE_URL")})

	// Starts containers and sets up networks
	oryConfig, err := config.OryFromEnv()
	if err != nil {
		panic(err)
	}

	// Unless the host is hard coded by an environment variable set it from the config file
	if hostConfig.Host == "" && launcherConfig.Subdomain != "" {
		hostConfig.Host = launcherConfig.Subdomain + ".homecloudapp.com"
	}

	// If the config is already set start the system
	if hostConfig.Host != "" {
		if err := launcher.StartSystem(dockerClient, storeClient, *hostConfig, *oryConfig, *storageConfig, *launcherEnvConfig, deviceConfig); err != nil {
			panic(err)
		}
	}

	// Prints message after interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\n\nExiting launcher, homecloud container still running")
		os.Exit(0)
	}()

	e := echo.New()
	e.Use(
		middleware.RequestLoggerWithConfig(
			middleware.RequestLoggerConfig{
				LogStatus: true,
				LogURI:    true,
				LogHost:   true,
				LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
					fmt.Printf("REQUEST URI: %s%s, status: %v\n", v.Host, v.URI, v.Status)
					return nil
				},
			},
		),
	)
	launcher.AddRoutes(e, dockerClient, storeClient, *hostConfig, *oryConfig, *storageConfig, *launcherEnvConfig, deviceConfig, launcherConfig)
	e.Logger.Fatal(e.Start(":1324"))
}
