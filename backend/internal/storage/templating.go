package storage

import (
	"fmt"
	"io"
	"text/template"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
)

type PackageTemplateParams struct {
	OAuthClientID     string
	OAuthClientSecret string
	OAuthIssuerUrl    string
	HostUrl           string
	HomecloudAppDir   string
	AppUrl            string
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

	appUrl := hostConfig.Url
	appUrl.Host = fmt.Sprintf("%s.%s", app.Name, hostConfig.Url.Host)

	parameters := PackageTemplateParams{
		OAuthClientID:     oauthClientID,
		OAuthClientSecret: oauthClientSecret,
		OAuthIssuerUrl:    oryConfig.Hydra.PublicAddress.String(),
		HostUrl:           hostConfig.Url.String(),
		HomecloudAppDir:   storageConfig.AppDir,
		AppUrl:            appUrl.String(),
	}

	return appTemplate.Execute(output, parameters)
}
