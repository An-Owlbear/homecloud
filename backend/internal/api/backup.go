package api

import (
	"net/http"

	"github.com/An-Owlbear/homecloud/backend/internal/storage"
	"github.com/labstack/echo/v4"
)

func ListExternalStorage() echo.HandlerFunc {
	return func(c echo.Context) error {
		devices, err := storage.ListExternalStorage()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSONPretty(http.StatusOK, devices, "  ")
	}
}
