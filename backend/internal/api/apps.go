package api

import (
	"context"
	"net/http"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func ListApps(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		appList, err := queries.GetApps(context.Background())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		var resList []apps.PackageListItem
		for _, app := range appList {
			resList = append(resList, apps.PackageListItem{
				Id:          app.ID,
				Name:        app.Schema.Name,
				Version:     app.Schema.Version,
				Author:      app.Schema.Author,
				Description: app.Schema.Description,
			})
		}

		return c.JSONPretty(200, resList, "  ")
	}
}

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
