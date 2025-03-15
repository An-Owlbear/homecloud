package persistence

import (
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"io"
	"text/template"
)

type PackageTemplateParams struct {
	OAuthClientID     string
	OAuthClientSecret string
	OAuthIssuerUrl    string
	HostUrl           string
	HomecloudAppDir   string
}

func ApplyAppTemplate(
	input string,
	output io.Writer,
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

	parameters := PackageTemplateParams{
		OAuthClientID:     oauthClientID,
		OAuthClientSecret: oauthClientSecret,
		OAuthIssuerUrl:    oryConfig.Hydra.PublicAddress.String(),
		HostUrl:           hostConfig.Url.String(),
		HomecloudAppDir:   storageConfig.AppDir,
	}

	return appTemplate.Execute(output, parameters)
}
