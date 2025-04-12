package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"text/template"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
)

type PackageTemplateParams struct {
	OAuthClientID     string
	OAuthClientSecret string
	OAuthIssuerUrl    string
	HostUrl           string
	FullHostUrl       string
	HostPort          int
	HomecloudAppDir   string
	AppUrl            string
	UrlScheme         string
	Environment       string
}

// ApplyAppTemplate applies templated values to the given input. In the case of templating an app package the app
// package itself must also be passed to ensure values like name and url are properly set
func ApplyAppTemplate(
	input string,
	output io.Writer,
	app persistence.AppPackage,
	oauthClientID string,
	oauthClientSecret string,
	oryConfig config.Ory,
	hostConfig config.Host,
	storageConfig config.Storage,
) error {
	appTemplate, err := template.New("appTemplate").Parse(input)
	if err != nil {
		return err
	}

	scheme := "http"
	if hostConfig.HTTPS {
		scheme = "https"
	}
	hostUrl := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", hostConfig.Host, hostConfig.Port),
	}

	parameters := PackageTemplateParams{
		OAuthClientID:     oauthClientID,
		OAuthClientSecret: oauthClientSecret,
		OAuthIssuerUrl:    oryConfig.Hydra.PublicAddress.String(),
		HostUrl:           hostConfig.Host,
		FullHostUrl:       hostUrl.String(),
		HostPort:          hostConfig.Port,
		HomecloudAppDir:   storageConfig.AppDir,
		AppUrl:            hostConfig.PublicSubdomain(app.Name),
		UrlScheme:         hostUrl.Scheme,
		Environment:       string(config.GetEnvironment()),
	}

	return appTemplate.Execute(output, parameters)
}

func TemplateAppPackage(
	input persistence.AppPackage,
	oauthClientID string,
	oauthClientSecret string,
	oryConfig config.Ory,
	hostConfig config.Host,
	storageConfig config.Storage,
) (persistence.AppPackage, error) {
	packageBytes, err := json.Marshal(input)
	if err != nil {
		return persistence.AppPackage{}, err
	}

	var templateOutput bytes.Buffer
	err = ApplyAppTemplate(
		string(packageBytes),
		&templateOutput,
		input,
		oauthClientID,
		oauthClientSecret,
		oryConfig,
		hostConfig,
		storageConfig,
	)
	if err != nil {
		return persistence.AppPackage{}, err
	}

	var output persistence.AppPackage
	err = json.Unmarshal(templateOutput.Bytes(), &output)
	if err != nil {
		return persistence.AppPackage{}, err
	}

	return output, nil
}
