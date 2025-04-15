package launcher

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/networking"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
)

func StartSystem(
	dockerClient *client.Client,
	storeClient *apps.StoreClient,
	hostConfig config.Host,
	oryConfig config.Ory,
	storageConfig config.Storage,
	launcherEnvConfig config.LauncherEnv,
	deviceConfig config.DeviceConfig,
) error {
	// Sets up port forwarding on local network
	if hostConfig.PortForward {
		// Stops containers to prevent attempting to map to port already in use
		if err := StopContainers(dockerClient); err != nil {
			return err
		}

		tempServer := echo.New()
		go func() {
			tempServer.Logger.Info(tempServer.Start(fmt.Sprintf(":%d", hostConfig.Port)))
		}()

		err := networking.TryMapPort(
			context.Background(),
			uint16(hostConfig.Port),
			uint16(hostConfig.Port),
			deviceConfig,
		)
		if err != nil {
			slog.Error("Error forwarding port: " + err.Error())
			//return err
		}

		err = networking.CheckPortForwarding(deviceConfig, hostConfig.Port)
		if err != nil {
			slog.Error("Error checking port forwarding: " + err.Error())
			//return err
		}

		if err := tempServer.Shutdown(context.Background()); err != nil {
			slog.Error("Error shutting down server: " + err.Error())
			return err
		}

		ticker := time.NewTicker(time.Hour)
		go func() {
			for {
				select {
				case <-ticker.C:
					err = networking.TryMapPort(
						context.Background(),
						uint16(hostConfig.Port),
						uint16(hostConfig.Port),
						deviceConfig,
					)
					if err != nil {
						slog.Error("Error forwarding port: " + err.Error())
						//return err
					}
				}
			}
		}()
	}

	if err := SetupTemplates(hostConfig, storageConfig); err != nil {
		return err
	}

	err := StartContainers(dockerClient, storeClient, oryConfig, hostConfig, storageConfig, launcherEnvConfig)
	if err != nil {
		return err
	}
	err = ConnectNetworks(dockerClient)
	if err != nil {
		return err
	}

	return nil
}
