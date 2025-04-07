package auth

import (
	"context"

	hydra "github.com/ory/hydra-client-go/v2"
)

// SetupAppAuth setups auth clients for a single app
func SetupAppAuth(
	hydraAdmin *hydra.APIClient,
	name string,
	scope string,
	redirectUris []string,
	oauthMethod string,
) (createdClient *hydra.OAuth2Client, err error) {
	oauth2Client := *hydra.NewOAuth2Client()
	oauth2Client.SetScope(scope)
	oauth2Client.SetClientName(name)
	oauth2Client.SetRedirectUris(redirectUris)
	oauth2Client.SetGrantTypes([]string{"authorization_code"})
	oauth2Client.SetResponseTypes([]string{"code", "id_token"})
	oauth2Client.SetTokenEndpointAuthMethod(oauthMethod)

	createdClient, _, err = hydraAdmin.OAuth2API.
		CreateOAuth2Client(context.Background()).
		OAuth2Client(oauth2Client).
		Execute()

	return
}
