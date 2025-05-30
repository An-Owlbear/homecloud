package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	hydra "github.com/ory/hydra-client-go/v2"
	kratos "github.com/ory/kratos-client-go"

	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/templates"
)

type OryRequest struct {
	Flow           string `query:"flow"`
	Aal            string `query:"aal"`
	Refresh        string `query:"refresh"`
	ReturnTo       string `query:"return_to"`
	Organisation   string `query:"organisation"`
	Via            string `query:"via"`
	Code           string `query:"code"`
	LoginChallenge string `query:"login_challenge"`
}

func Login(kratosClient *kratos.APIClient, kratosAdmin kratos.IdentityAPI, oryConfig config.Ory) echo.HandlerFunc {
	return func(c echo.Context) error {
		// If there are no users in the system redirect to setup url
		users, err := auth.ListUsers(c.Request().Context(), kratosAdmin)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			return c.Redirect(http.StatusFound, "/auth/setup")
		}

		// Parses the query parameters in request
		var request OryRequest
		err = c.Bind(&request)
		if err != nil {
			slog.Error(err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid login request")
		}

		// Builds the redirect URL
		queryParams := url.Values{
			"aal":          {request.Aal},
			"refresh":      {request.Refresh},
			"return_to":    {request.ReturnTo},
			"organisation": {request.Organisation},
			"via":          {request.Via},
		}

		if request.LoginChallenge != "" {
			queryParams.Add("login_challenge", request.LoginChallenge)
		}

		redirectUri := oryConfig.Kratos.PublicAddress
		redirectUri.Path = "/self-service/login/browser"
		redirectUri.RawQuery = queryParams.Encode()
		redirectString := redirectUri.String()

		// Redirects if flow is not set
		if request.Flow == "" {
			return c.Redirect(http.StatusFound, redirectString)
		}

		// Retrieves login flow
		flow, resp, err := kratosClient.FrontendAPI.GetLoginFlow(c.Request().Context()).
			Id(request.Flow).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		// If login flow not found assume expired or missing and redirect
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return c.Redirect(http.StatusFound, redirectString)
			}

			return err
		}

		// If flow retrieved successfully render page
		return render(c, http.StatusOK, templates.Login(flow.Ui))
	}
}

func Registration(kratosClient *kratos.APIClient, oryConfig config.Ory) echo.HandlerFunc {
	return func(c echo.Context) error {
		inviteCode := c.QueryParam("code")
		flowId := c.QueryParam("flow")

		redirectUri := oryConfig.Kratos.PublicAddress
		redirectUri.Path = "/self-service/registration/browser"
		redirectUri.RawQuery = url.Values{"code": {inviteCode}}.Encode()

		// Redirects to Ory self-service if now flow is set
		if flowId == "" {
			return c.Redirect(http.StatusFound, redirectUri.String())
		}

		flow, resp, err := kratosClient.FrontendAPI.GetRegistrationFlow(c.Request().Context()).
			Id(flowId).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		// Redirects to Ory self-service to make a new flow if the given one is not found
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return c.Redirect(http.StatusFound, redirectUri.String())
			}

			return err
		}

		originalUrl, err := url.Parse(flow.RequestUrl)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid url received")
		}
		inviteCode = originalUrl.Query().Get("code")

		// Encodes the invite code to add to the registration form
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

func Settings(kratosClient *kratos.APIClient, oryConfig config.Ory) echo.HandlerFunc {
	return func(c echo.Context) error {
		flowId := c.QueryParam("flow")

		redirectUri := oryConfig.Kratos.PublicAddress
		redirectUri.Path = "/self-service/settings/browser"

		// Redirects to self-service url to create flow if none is given
		if flowId == "" {
			return c.Redirect(http.StatusMovedPermanently, redirectUri.String())
		}

		flow, resp, err := kratosClient.FrontendAPI.GetSettingsFlow(c.Request().Context()).
			Id(flowId).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		// Redirects to create flow if given one is not found
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return c.Redirect(http.StatusMovedPermanently, redirectUri.String())
			}

			slog.Error(err.Error())
			return err
		}

		// Returns form rendered from flow
		return render(c, http.StatusOK, templates.Settings(flow.Ui))
	}
}

func Recovery(kratosClient *kratos.APIClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		flowId := c.QueryParam("flow")
		if flowId == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid recovery link")
		}

		// Checks flow is valid and renders form if it is
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

func Logout(kratosClient *kratos.APIClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Creates and executes the logout flow
		flow, _, err := kratosClient.FrontendAPI.CreateBrowserLogoutFlow(c.Request().Context()).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Error during logout process")
		}

		_, err = kratosClient.FrontendAPI.UpdateLogoutFlow(c.Request().Context()).
			Cookie(c.Request().Header.Get("Cookie")).
			Token(flow.LogoutToken).
			Execute()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Error during logout process")
		}

		return c.Redirect(http.StatusFound, "/")
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

func InitialSetup(kratosAdmin kratos.IdentityAPI, queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		// If no users exist create an invite for the owner, otherwise return error
		users, err := auth.ListUsers(c.Request().Context(), kratosAdmin)
		if err != nil {
			slog.Error(err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users")
		}

		if len(users) != 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, "First user already created")
		}

		inviteCode, err := queries.CreateInviteCode(c.Request().Context(), persistence.CreateInviteCodeParams{
			Hours:     1,
			Rolesjson: "[\"admin\", \"user\"]",
		})
		if err != nil {
			slog.Error(err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create invite code")
		}

		return c.Redirect(http.StatusFound, "/auth/registration?code="+inviteCode.Code)
	}
}

type userOptionsResponse struct {
	persistence.UserOption
	UserRoles []string `json:"user_roles"`
}

func GetUserOptions(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := c.(*config.Context)
		options, err := queries.GetUserOptions(c.Request().Context(), cc.Session.Identity.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user options")
		}

		userMetadata, err := auth.ParseMetadataPublic(cc.Session.Identity.MetadataPublic)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user metadata")
		}

		return c.JSON(http.StatusOK, userOptionsResponse{
			UserOption: options,
			UserRoles:  userMetadata.Roles,
		})
	}
}

type updateOptionsRequest struct {
	CompletedWelcome *bool `json:"completed_welcome,omitempty"`
}

func UpdateUserOptions(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := c.(*config.Context)
		var request updateOptionsRequest
		if err := c.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid update request")
		}

		var completedWelcome sql.NullBool
		if request.CompletedWelcome != nil {
			completedWelcome = sql.NullBool{Bool: *request.CompletedWelcome, Valid: true}
		} else {
			completedWelcome = sql.NullBool{Valid: false}
		}

		if err := queries.UpdateUserOptions(c.Request().Context(), persistence.UpdateUserOptionsParams{
			UserID:           cc.Session.Identity.Id,
			CompletedWelcome: completedWelcome,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user options")
		}

		return c.NoContent(http.StatusNoContent)
	}
}
