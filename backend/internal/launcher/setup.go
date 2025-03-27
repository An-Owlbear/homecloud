package launcher

import (
	"context"
	"log/slog"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/networking"
	"github.com/docker/docker/client"
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

	// Sets up port forwarding on local network
	if hostConfig.PortForward {
		err = networking.TryMapPort(
			context.Background(),
			uint16(hostConfig.Port),
			uint16(hostConfig.Port),
			deviceConfig,
		)
		if err != nil {
			slog.Error("Error forwarding port: ", err)
			//return err
		}

		err = networking.CheckPortForwarding(deviceConfig, hostConfig.Port)
		if err != nil {
			slog.Error("Error checking port forwarding: ", err)
			//return err
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
						slog.Error("Error forwarding port: ", err)
						//return err
					}
				}
			}
		}()
	}

	return nil
}
