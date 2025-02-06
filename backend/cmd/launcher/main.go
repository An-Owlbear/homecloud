package main

import (
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
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
}
