package api

import (
	"github.com/An-Owlbear/homecloud/backend/internal/templates"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"log/slog"
	"net/http"
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
		slog.Info(kratosClient.GetConfig().Host)

		flowId := c.QueryParam("flow")
		if flowId == "" {
			return c.Redirect(http.StatusMovedPermanently, "http://kratos.hc.anowlbear.com:1323/self-service/registration/browser")
		}

		flow, resp, err := kratosClient.FrontendAPI.GetRegistrationFlow(c.Request().Context()).
			Id(flowId).
			Cookie(c.Request().Header.Get("Cookie")).
			Execute()

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return c.Redirect(http.StatusMovedPermanently, "http://kratos.hc.anowlbear.com:1323/self-service/registration/browser")
			}

			return err
		}

		return render(c, http.StatusOK, templates.Registration(flow.Ui))
	}
}
