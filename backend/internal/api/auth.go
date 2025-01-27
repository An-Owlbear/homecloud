package api

import (
	"encoding/json"
	"github.com/An-Owlbear/homecloud/backend/internal/templates"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"log/slog"
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

		slog.Info(flow.RequestUrl)
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
