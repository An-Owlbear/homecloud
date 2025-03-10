package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/auth"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	hydra "github.com/ory/hydra-client-go/v2"
	"net/http"
	"strings"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func ListPackages(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		packages, err := queries.GetPackages(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving packages")
		}
		return c.JSONPretty(200, packages, "  ")
	}
}

func GetPackage(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		pkg, err := queries.GetPackage(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return echo.NewHTTPError(http.StatusNotFound, "package not found")
			}
			return err
		}
		return c.JSONPretty(200, pkg, "  ")
	}
}

type SearchParams struct {
	SearchTerm string `query:"q"`
	Category   string `query:"category"`
	Developer  string `query:"developer"`
}

func SearchPackages(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		var params SearchParams
		if err := c.Bind(&params); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid search parameters")
		}

		packages, err := queries.SearchPackages(c.Request().Context(), params.SearchTerm, params.Category, params.Developer)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving packages")
		}
		return c.JSONPretty(200, packages, "  ")
	}
}

func ListCategories(queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		categories, err := queries.GetCategories(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving categories")
		}
		return c.JSONPretty(200, categories, "  ")
	}
}

func AddPackage(
	storeClient *apps.StoreClient,
	queries *persistence.Queries,
	dockerClient *client.Client,
	hydraAdmin *hydra.APIClient,
	hostConfig config.Host,
	appDataHandler *persistence.AppDataHandler,
) echo.HandlerFunc {
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
				// If the redirect uri starts with a slash append the actual host
				for _, redirectUri := range appContainer.OidcRedirectUris {
					if strings.HasPrefix(redirectUri, "/") {
						redirectUri = fmt.Sprintf("http://%s.%s:%d%s", app.Name, hostConfig.Host, hostConfig.Port, redirectUri)
					}
					redirectUris = append(redirectUris, redirectUri)
				}
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

		// Downloads the app package and stores required files
		err = appDataHandler.SavePackage(app.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Install and sets up app containers
		err = docker.InstallApp(dockerClient, app, hostConfig)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.JSONPretty(200, app, "  ")
	}
}

func CheckUpdates(storeClient *apps.StoreClient, queries *persistence.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := storeClient.UpdatePackageList(c.Request().Context(), queries)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.String(200, "List updated")
	}
}

func UpdateApps(dockerClient *client.Client, storeClient *apps.StoreClient, queries *persistence.Queries, hostConfig config.Host) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := apps.UpdateApps(dockerClient, storeClient, queries, hostConfig)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "Apps updated")
	}
}
