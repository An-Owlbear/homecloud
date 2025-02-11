package launcher

import (
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"net/http"
)

func AddRoutes(
	e *echo.Echo,
	dockerClient *client.Client,
	storeClient *apps.StoreClient,
) {
	e.GET("/api/v1/update", CheckUpdateHandler())
	e.POST("/api/v1/update", ApplyUpdatesHandler(dockerClient))
}

func CheckUpdateHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		updates, err := CheckUpdates()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, updates)
	}
}

func ApplyUpdatesHandler(dockerClient *client.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		//err := ApplyUpdates()
		//if err != nil {
		//	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		//}

		// After returning function stop containers and restart system
		c.Response().After(func() {
			println("test")
			StopContainers(dockerClient)
		})

		return c.String(http.StatusOK, "")
	}
}
