package api

import (
	"errors"
	"net/http"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/storage"
	"github.com/docker/docker/client"
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

type backupRequest struct {
	TargetDevice string `json:"target_device"`
}

func BackupApp(dockerClient *client.Client, storageConfig config.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Retrieves request information
		appId := c.Param("appId")
		var request backupRequest
		if err := c.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Performs backup and returns response
		err := apps.BackupApp(c.Request().Context(), dockerClient, storageConfig, appId, request.TargetDevice)
		if err != nil {
			if errors.Is(err, storage.DriveInvalidError) {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusNoContent)
	}
}

type restoreRequest struct {
	TargetDevice string `json:"target_device"`
	Backup       string `json:"backup"`
}

func RestoreApp(
	dockerClient *client.Client,
	queries *persistence.Queries,
	hosts *apps.Hosts,
	appDataHandler *storage.AppDataHandler,
	hostConfig config.Host,
	storageConfig config.Storage,
	oryConfig config.Ory,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		appId := c.Param("appId")
		var request restoreRequest
		if err := c.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err := apps.RestoreApp(c.Request().Context(), dockerClient, queries, hosts, appDataHandler, hostConfig, storageConfig, oryConfig, appId, request.TargetDevice, request.Backup)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusNoContent)
	}
}
