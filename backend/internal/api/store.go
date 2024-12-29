package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
)

func AddPackage(storeClient *apps.StoreClient, queries *persistence.Queries, dockerClient *client.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("appId")
		if id == "" {
			return c.String(400, "Must provide id query parameter")
		}

		// Retrieves app from store
		app, err := storeClient.GetPackage(id)
		if err != nil {
			return c.String(500, err.Error())
		}

		// Converts to json string for storing in DB
		json, err := json.Marshal(app)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Creates app in DB
		err = queries.CreateApp(context.Background(), persistence.CreateAppParams{
			ID:     app.Id,
			Schema: string(json),
		})
		if err != nil {
			// If the app is already in the database return an error
			if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return echo.NewHTTPError(http.StatusBadRequest, "App already installed")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Install and sets up app containers
		err = apps.InstallApp(dockerClient, app)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.JSONPretty(200, app, "  ")
	}
}

func CheckUpdates(storeClient *apps.StoreClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := storeClient.UpdatePackageList()
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.String(200, "List updated")
	}
}

func UpdateApps(appManager *apps.AppManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := appManager.UpdateApps()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "Apps updated")
	}
}
