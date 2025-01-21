package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	hydra "github.com/ory/hydra-client-go/v2"
	"net/http"
	"slices"
	"strings"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func AddPackage(storeClient *apps.StoreClient, queries *persistence.Queries, dockerClient *client.Client, hydraAdmin *hydra.APIClient) echo.HandlerFunc {
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

		// Checks if app is already installed before creating oauth2 client
		_, err = queries.GetApp(context.Background(), app.Id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		} else if err == nil {
			return echo.NewHTTPError(http.StatusConflict, "App already exists")
		}

		// Creates oauth2 client for the app if required
		var clientId string
		var clientSecret string
		if app.OidcEnabled {
			var redirectUris []string
			for _, appContainer := range app.Containers {
				redirectUris = slices.Concat(redirectUris, appContainer.OidcRedirectUris)
			}

			oidcClient, err := auth.SetupAppAuth(hydraAdmin, app.Name, strings.Join(app.OidcScopes[:], " "), redirectUris)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}

			clientId = oidcClient.GetClientId()
			clientSecret = oidcClient.GetClientSecret()
		}

		// Creates app in DB
		err = queries.CreateApp(context.Background(), persistence.CreateAppParams{
			ID:           app.Id,
			Schema:       string(json),
			ClientID:     sql.NullString{String: clientId, Valid: clientId != ""},
			ClientSecret: sql.NullString{String: clientSecret, Valid: clientSecret != ""},
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Install and sets up app containers
		err = docker.InstallApp(dockerClient, app)
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

func UpdateApps(dockerClient *client.Client, storeClient *apps.StoreClient, queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := apps.UpdateApps(dockerClient, storeClient, queries)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "Apps updated")
	}
}
