package api

import (
	"context"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"log/slog"
	"net/http"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

type InstalledApp struct {
	apps.PackageListItem
	Status string `json:"status"`
}

func ListApps(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Retrieves list of apps
		appList, err := queries.GetApps(context.Background())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// returns them in a more compact format fit for lists
		resList := make([]InstalledApp, 0)
		for _, app := range appList {
			resList = append(resList, InstalledApp{
				PackageListItem: apps.PackageListItem{
					Id:          app.ID,
					Name:        app.Schema.Name,
					Version:     app.Schema.Version,
					Author:      app.Schema.Author,
					Description: app.Schema.Description,
					Categories:  app.Schema.Categories,
					ImageUrl:    "/assets/data/" + app.ID + "/icon.png",
				},
				Status: app.Status,
			})
		}

		return c.JSONPretty(200, resList, "  ")
	}
}

func StartApp(
	dockerClient *client.Client,
	queries *persistence.Queries,
	hosts *apps.Hosts,
	appDataHandler *persistence.AppDataHandler,
	hostConfig config.Host,
	oryConfig config.Ory,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		appId := c.Param("appId")
		err := apps.StartApp(dockerClient, queries, hosts, appDataHandler, hostConfig, oryConfig, appId)
		if err != nil {
			slog.Error("Error starting app:" + err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start app")
		}

		return c.String(200, "App started!")
	}
}

func StopApp(dockerClient *client.Client, queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		appId := c.Param("appId")
		if appId == "" {
			return c.String(400, "App ID must be set")
		}

		err := apps.StopApp(dockerClient, queries, appId)
		if err != nil {
			return c.String(500, "Failed to stop app")
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
			return c.String(404, appId+" not found")
		}

		// Uninstalls the app deleting the docker resources
		if err := docker.UninstallApp(dockerClient, appId); err != nil {
			return c.String(500, err.Error())
		}

		return c.String(200, "App uninstalled!")
	}
}
