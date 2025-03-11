package launcher

import (
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/networking"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"net/http"
	"os/exec"
)

func AddRoutes(
	e *echo.Echo,
	dockerClient *client.Client,
	storeClient *apps.StoreClient,
	deviceConfig config.DeviceConfig,
) {
	e.GET("/api/v1/update", CheckUpdateHandler())
	e.POST("/api/v1/update", ApplyUpdatesHandler(dockerClient))
	e.POST("/api/v1/set_subdomain", SetSubdomainHandler(deviceConfig))
}

type checkUpdateResponse struct {
	UpdateRequired bool `json:"update_required"`
}

func CheckUpdateHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		updates, err := CheckUpdates()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, checkUpdateResponse{UpdateRequired: updates})
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

type SubdomainAPIRequest struct {
	Subdomain string `json:"subdomain"`
}

func SetSubdomainHandler(deviceConfig config.DeviceConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req SubdomainAPIRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		publicIP, err := networking.GetPublicIP()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to setup networking")
		}

		err = SetSubdomain(c.Request().Context(), SubdomainRequest{
			DeviceId:  deviceConfig.DeviceId,
			DeviceKey: deviceConfig.DeviceKey,
			Subdomain: req.Subdomain,
			IPAddress: publicIP,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error setting subdomain")
		}

		return c.NoContent(http.StatusNoContent)
	}
}
