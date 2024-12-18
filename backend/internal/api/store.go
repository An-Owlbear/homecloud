package api

import (
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/labstack/echo/v4"
)

func AddPackage(storeClient *apps.StoreClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.QueryParams().Get("id")
		if id == "" {
			return c.String(400, "Must provide id query parameter")
		}

		app, err := storeClient.GetPackage(id)
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
