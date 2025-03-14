package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"text/template"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/launcher"
	"github.com/An-Owlbear/homecloud/backend/internal/networking"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

const templateDir = "templates"
const configDir = "ory_config"

type OryTemplateParams struct {
	HostUrl        string
	KratosUrl      string
	KratosAdminUrl string
	HydraUrl       string
	RootHost       string
}

func main() {
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
	launcherConfig, err := config.NewLauncher()
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

	hostUrl := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", hostConfig.Host, hostConfig.Port),
	}
	kratosUrl := hostUrl
	kratosUrl.Host = fmt.Sprintf("%s.%s", "kratos", hostUrl.Host)
	hydraUrl := hostUrl
	hydraUrl.Host = fmt.Sprintf("%s.%s", "hydra", hostUrl.Host)

	kratosAdminUrl := url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:4434",
	}

	templateParams := OryTemplateParams{
		HostUrl:        hostUrl.String(),
		KratosUrl:      kratosUrl.String(),
		KratosAdminUrl: kratosAdminUrl.String(),
		HydraUrl:       hydraUrl.String(),
		RootHost:       hostConfig.Host,
	}

	// Parses and produces templates
	templatePath := path.Join(configDir, templateDir)
	dir, err := os.Open(templatePath)
	if err != nil {
		panic(err)
	}
	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		templatePath := path.Join(templatePath, file)
		templateFile, err := template.ParseFiles(templatePath)
		if err != nil {
			panic(err)
		}

		writer, err := os.Create(path.Join(configDir, file))
		if err != nil {
			panic(err)
		}
		err = templateFile.Execute(writer, templateParams)
		if err != nil {
			panic(err)
		}
		err = writer.Close()
		if err != nil {
			panic(err)
		}
	}

	// Copies required files to ory data folders
	for app, files := range map[string][]string{
		"ory.kratos":    {"ory_config/kratos.yml", "ory_config/identity.schema.json", "ory_config/invite_code.jsonnet"},
		"ory.hydra":     {"ory_config/hydra.yml"},
		"homecloud.app": {".env", ".dev.env"},
	} {
		dataPath := path.Join(storageConfig.DataPath, app, "data")
		if _, err := os.Stat(dataPath); err != nil {
			err = os.MkdirAll(dataPath, 0755)
			if err != nil {
				panic(err)
			}
		}
		for _, file := range files {
			if err := os.MkdirAll(filepath.Dir(path.Join(dataPath, file)), 0755); err != nil {
				panic(err)
			}
			writer, err := os.Create(path.Join(dataPath, file))
			if err != nil {
				panic(err)
			}
			reader, err := os.Open(file)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(writer, reader)
			if err != nil {
				panic(err)
			}
			reader.Close()
			writer.Close()
		}
	}

	// Creates docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	storeClient := apps.NewStoreClient(config.Store{StoreUrl: os.Getenv("SYSTEM_STORE_URL")})

	// Starts containers and sets up networks
	err = launcher.StartContainers(dockerClient, storeClient, *hostConfig, *storageConfig, *launcherConfig)
	if err != nil {
		panic(err)
	}
	err = launcher.ConnectNetworks(dockerClient)
	if err != nil {
		panic(err)
	}

	// Sets up port forwarding on local network
	if hostConfig.PortForward {
		err = networking.TryMapPort(context.Background(), uint16(hostConfig.Port), uint16(hostConfig.Port), deviceConfig)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Printing logs from homecloud container")

	// Prints message after interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\n\nExiting launcher, homecloud container still running")
		os.Exit(0)
	}()

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
	launcher.AddRoutes(e, dockerClient, storeClient, deviceConfig)
	e.Logger.Fatal(e.Start(":1324"))
}
