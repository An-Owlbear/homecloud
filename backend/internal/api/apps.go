package api

import (
	"context"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
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

func UninstallApp(queries *persistence.Queries, dockerClient *client.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		appId := c.Param("appId")

		// Removes the app entry from the database
		result, err := queries.RemoveApp(context.Background(), appId)
		if err != nil {
			return c.String(500, err.Error())
		}

		// Returns 404 if no rows are deleted - e.g. no app is found
		rows, err := result.RowsAffected()
		if err != nil && rows == 0 {
			c.String(404, appId+" not found")
		}

		// Uninstalls the app deleting the docker resources
		if err := apps.UninstallApp(dockerClient, appId); err != nil {
			return c.String(500, err.Error())
		}

		return c.String(200, "App uninstalled!")
	}
}
