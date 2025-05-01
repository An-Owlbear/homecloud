package launcher

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
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

// SetupTemplates templates and copies the required files for Ory and Homecloud to their respective folders
func SetupTemplates(hostConfig config.Host, storageConfig config.Storage) error {
	scheme := "http"
	if hostConfig.HTTPS {
		scheme = "https"
	}
	host := hostConfig.Host
	if hostConfig.Port != 80 && hostConfig.Port != 443 {
		host = fmt.Sprintf("%s:%d", hostConfig.Host, hostConfig.Port)
	}

	hostUrl := url.URL{
		Scheme: scheme,
		Host:   host,
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
		return err
	}
	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}

	for _, file := range files {
		templatePath := path.Join(templatePath, file)
		templateFile, err := template.ParseFiles(templatePath)
		if err != nil {
			return err
		}

		writer, err := os.Create(path.Join(configDir, file))
		if err != nil {
			return err
		}
		err = templateFile.Execute(writer, templateParams)
		if err != nil {
			return err
		}
		err = writer.Close()
		if err != nil {
			return err
		}
	}

	homecloudFiles := []string{".env"}
	if config.GetEnvironment() == config.Development {
		homecloudFiles = append(homecloudFiles, ".dev.env")
	}

	// Copies required files to ory data folders
	for app, files := range map[string][]string{
		"ory.kratos":    {"ory_config/kratos.yml", "ory_config/identity.schema.json", "ory_config/invite_code.jsonnet"},
		"ory.hydra":     {"ory_config/hydra.yml"},
		"homecloud.app": homecloudFiles,
	} {
		dataPath := path.Join(storageConfig.DataPath, app, "data")
		if _, err := os.Stat(dataPath); err != nil {
			err = os.MkdirAll(dataPath, 0755)
			if err != nil {
				return err
			}
		}
		for _, file := range files {
			if err := os.MkdirAll(filepath.Dir(path.Join(dataPath, file)), 0755); err != nil {
				return err
			}
			writer, err := os.Create(path.Join(dataPath, file))
			if err != nil {
				return err
			}
			reader, err := os.Open(file)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, reader)
			if err != nil {
				return err
			}
			reader.Close()
			writer.Close()
		}
	}

	return nil
}
