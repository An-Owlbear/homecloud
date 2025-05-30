package launcher

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/networking"
)

func AddRoutes(
	e *echo.Echo,
	dockerClient *client.Client,
	storeClient *apps.StoreClient,
	hostConfig config.Host,
	oryConfig config.Ory,
	storageConfig config.Storage,
	launcherEnvConfig config.LauncherEnv,
	deviceConfig config.DeviceConfig,
	launcherConfig *Config,
) {
	e.GET("/api/v1/update", CheckUpdateHandler())
	e.POST("/api/v1/update", ApplyUpdatesHandler(dockerClient))
	e.POST("/api/v1/set_subdomain", SetSubdomainHandler(dockerClient, storeClient, hostConfig, oryConfig, storageConfig, launcherEnvConfig, deviceConfig, launcherConfig))
	e.POST("/api/v1/check_subdomain", CheckSubdomain())
	e.GET("/api/v1/current_subdomain", GetRegisteredDomain(launcherConfig))

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "launcher_spa",
		HTML5: true,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/api")
		},
	}))
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

		// Manually writes headers and response to finish HTTP request before restarting
		c.Response().WriteHeader(http.StatusOK)
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
		c.Response().Write([]byte("Updated and restarting"))
		c.Response().Flush()

		// Restarts the server
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

func SetSubdomainHandler(
	dockerClient *client.Client,
	storeClient *apps.StoreClient,
	hostConfig config.Host,
	oryConfig config.Ory,
	storageConfig config.Storage,
	launcherEnvConfig config.LauncherEnv,
	deviceConfig config.DeviceConfig,
	launcherConfig *Config,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req SubdomainAPIRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		// Retrieves the public IP address for subdomain assignment
		publicIP, err := networking.GetPublicIP()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to setup networking")
		}

		// Requests to set the devices registered subdomain
		err = SetSubdomain(c.Request().Context(), SubdomainRequest{
			DeviceId:  deviceConfig.DeviceId,
			DeviceKey: deviceConfig.DeviceKey,
			Subdomain: req.Subdomain,
			IPAddress: publicIP,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error setting subdomain")
		}

		// Updates the launcher configuration to remember the assignment
		launcherConfig.Subdomain = req.Subdomain
		if err := launcherConfig.Save(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save config")
		}

		// Starts Docker system now the subdomain is configured
		hostConfig.Host = fmt.Sprintf("%s.homecloudapp.com", req.Subdomain)
		err = StartSystem(dockerClient, storeClient, hostConfig, oryConfig, storageConfig, launcherEnvConfig, deviceConfig)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusNoContent)
	}
}

type checkSubdomainRequest struct {
	Address string `json:"address"`
}

type checkSubdomainResponse struct {
	Taken bool `json:"taken"`
}

func CheckSubdomain() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req checkSubdomainRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		// Looks up IP address, if it isn't found return a normal response
		ip, err := net.LookupIP(req.Address)
		if err != nil {
			var dnsErr *net.DNSError
			if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
				response := checkSubdomainResponse{Taken: false}
				return c.JSON(http.StatusOK, response)
			}
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid address")
		}

		// Retrieves the public IP address
		publicIP, err := networking.GetPublicIP()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get public IP")
		}

		// Return response with IP address
		response := checkSubdomainResponse{Taken: ip[0].String() != publicIP}
		return c.JSON(http.StatusOK, response)
	}
}

type getRegisteredDomainResponse struct {
	Subdomain string `json:"subdomain"`
}

func GetRegisteredDomain(launcherConfig *Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, getRegisteredDomainResponse{Subdomain: launcherConfig.Subdomain})
	}
}
