package main

import (
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"net/url"
	"os"
	"path"
	"text/template"
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
	err := godotenv.Load()
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

	// Creates docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	storeClient := apps.NewStoreClient(os.Getenv("SYSTEM_STORE_URL"))

	// Installs ory hydra and kratos
	for _, packageName := range []string{"ory.kratos", "ory.hydra"} {
		// Checks if the app is already installed and continues if so
		appInstalled, err := docker.IsAppInstalled(dockerClient, packageName)
		if appInstalled {
			fmt.Printf("%s is already installed\n", packageName)
			continue
		}

		fmt.Printf("Installing %s\n", packageName)
		kratosPackage, err := storeClient.GetPackage(packageName)
		if err != nil {
			panic(err)
		}

		err = docker.InstallApp(dockerClient, kratosPackage)
		if err != nil {
			panic(err)
		}

		err = docker.StartApp(dockerClient, packageName)
		if err != nil {
			panic(err)
		}
	}
}
