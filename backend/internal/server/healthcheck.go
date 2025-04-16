package server

import (
	"net/http"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/labstack/echo/v4"
)

// HealthCheckServer creates an echo server for Docker's health check
func HealthCheckServer(hosts *apps.Hosts) *echo.Echo {
	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		if !hosts.CertsReady() {
			return c.String(http.StatusServiceUnavailable, "starting")
		}
		return c.String(http.StatusOK, "ready")
	})

	return e
}
