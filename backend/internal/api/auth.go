package api

import (
	"encoding/json"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/templates"
	"github.com/labstack/echo/v4"
	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"
	"net/http"
	"net/url"
	"path"
)

func Login(kratosClient *kratos.APIClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		flowId := c.QueryParam("flow")
		if flowId == "" {
			return c.Redirect(http.StatusMovedPermanently, "http://kratos.hc.anowlbear.com:1323/self-service/login/browser")
		}

		flow, resp, err := kratosClient.FrontendAPI.GetLoginFlow(c.Request().Context()).
			Id(flowId).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return c.Redirect(http.StatusMovedPermanently, path.Join(kratosClient.GetConfig().Host, "/self-service/login/browser"))
			}

			return err
		}

		return render(c, http.StatusOK, templates.Login(flow.Ui))
	}
}

func Registration(kratosClient *kratos.APIClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		inviteCode := c.QueryParam("code")
		flowId := c.QueryParam("flow")
		if flowId == "" {
			return c.Redirect(http.StatusFound, "http://kratos.hc.anowlbear.com:1323/self-service/registration/browser?code="+inviteCode)
		}

		flow, resp, err := kratosClient.FrontendAPI.GetRegistrationFlow(c.Request().Context()).
			Id(flowId).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return c.Redirect(http.StatusFound, "http://kratos.hc.anowlbear.com:1323/self-service/registration/browser?code="+inviteCode)
			}

			return err
		}

		originalUrl, err := url.Parse(flow.RequestUrl)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid url received")
		}
		inviteCode = originalUrl.Query().Get("code")

		if inviteCode == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot signup without an invite code")
		}
		inviteCodeRequest := invitationCodeRequest{
			InvitationCode: inviteCode,
		}
		inviteRequestString, err := json.Marshal(inviteCodeRequest)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid invite code")
		}

		return render(c, http.StatusOK, templates.Registration(flow.Ui, string(inviteRequestString)))
	}
}

func Settings(kratosClient *kratos.APIClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		flowId := c.QueryParam("flow")
		if flowId == "" {
			return c.Redirect(http.StatusMovedPermanently, "http://kratos.hc.anowlbear.com:1323/self-service/settings/browser")
		}

		flow, resp, err := kratosClient.FrontendAPI.GetSettingsFlow(c.Request().Context()).
			Id(flowId).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return c.Redirect(http.StatusMovedPermanently, path.Join(kratosClient.GetConfig().Host, "/self-service/settings/browser"))
			}

			return err
		}

		return render(c, http.StatusOK, templates.Settings(flow.Ui))
	}
}

func Recovery(kratosClient *kratos.APIClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		flowId := c.QueryParam("flow")
		if flowId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid recovery link")
		}

		flow, _, err := kratosClient.FrontendAPI.GetRecoveryFlow(c.Request().Context()).
			Id(flowId).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid recovery link")
		}

		return render(c, http.StatusOK, templates.Recovery(flow.Ui))
	}
}

func ListUsers(kratosAdminClient kratos.IdentityAPI) echo.HandlerFunc {
	return func(c echo.Context) error {
		users, err := auth.ListUsers(c.Request().Context(), kratosAdminClient)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users")
		}

		return c.JSON(http.StatusOK, users)
	}
}

func DeleteUser(kratosIdentity kratos.IdentityAPI) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId := c.Param("id")
		err := auth.DeleteUser(c.Request().Context(), kratosIdentity, userId)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user")
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func ResetPassword(kratosAdmin kratos.IdentityAPI) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Creates the password reset code
		userId := c.Param("id")
		code, resp, err := kratosAdmin.
			CreateRecoveryCodeForIdentity(c.Request().Context()).
			CreateRecoveryCodeForIdentityBody(*kratos.NewCreateRecoveryCodeForIdentityBody(userId)).
			Execute()

		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "User not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
		}

		// Returns response with code
		return c.JSON(http.StatusOK, code)
	}
}

// OidcConsent handles the oidc login to applications. Assumes the user is authenticated
func OidcConsent(hydraClient *hydra.APIClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parses parameter and current session
		challenge := c.QueryParam("consent_challenge")
		if challenge == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid consent challenge")
		}

		cc, ok := c.(*config.Context)
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "Error during request")
		}

		var traits auth.Traits
		err := auth.ParseSessionData(cc.Session.Identity.Traits, &traits)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user session")
		}

		// Retrieves consent request
		consent, _, err := hydraClient.OAuth2API.
			GetOAuth2ConsentRequest(c.Request().Context()).
			ConsentChallenge(challenge).
			Execute()

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get consent challenge")
		}

		// Creates the accept request
		acceptRequest := hydra.NewAcceptOAuth2ConsentRequest()
		acceptRequest.GrantScope = consent.RequestedScope
		acceptRequest.GrantAccessTokenAudience = consent.RequestedAccessTokenAudience

		acceptSession := hydra.NewAcceptOAuth2ConsentRequestSessionWithDefaults()
		acceptSession.IdToken = map[string]interface{}{
			"email": traits.Email,
			"name":  traits.Name,
		}

		acceptRequest.Session = acceptSession

		// IMPORTANT - currently automatically accepts oauth2 request, as currently only profile and email scopes
		// are functional, and it will only be used with self-hosted apps. If more functionality is added this
		// should be changed
		accept, _, err := hydraClient.OAuth2API.
			AcceptOAuth2ConsentRequest(c.Request().Context()).
			ConsentChallenge(challenge).
			AcceptOAuth2ConsentRequest(*acceptRequest).
			Execute()

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to accept consent request")
		}

		return c.Redirect(http.StatusFound, accept.RedirectTo)
	}
}
