package api

import (
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func StartApp(dockerClient *client.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		appId := c.Param("appId")
		if appId == "" {
			return c.String(400, "App ID must be set")
		}

		err := apps.StartApp(dockerClient, appId)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.String(200, "App started!")
	}
}

func StopApp(dockerClient *client.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		appId := c.Param("appId")
		if appId == "" {
			return c.String(400, "App ID must be set")
		}

		err := apps.StopApp(dockerClient, appId)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.String(200, "App stopped!")
	}
}
