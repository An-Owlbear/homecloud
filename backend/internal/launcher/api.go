package launcher

import (
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"net/http"
	"os/exec"
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
		updatesAvailable, err := CheckUpdates()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if !updatesAvailable {
			return c.String(http.StatusOK, "No updates available")
		}

		err = ApplyUpdates()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// After returning function stop containers and restart system
		c.Response().After(func() {
		})

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
		c.Response().WriteHeader(http.StatusOK)
		c.Response().Write([]byte("Updated and restarting"))
		c.Response().Flush()

		go func() {
			if err := exec.Command("reboot").Run(); err != nil {
				c.Logger().Error("Failed to reboot: ", err.Error())
			}
		}()

		return nil
	}
}
