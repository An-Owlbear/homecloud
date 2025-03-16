package launcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/mod/semver"
	"io"
	"os"
	"strings"
)

var appPackages = []string{"ory.kratos", "ory.hydra", "homecloud.app"}

func StartContainers(
	dockerClient *client.Client,
	storeClient *apps.StoreClient,
	oryConfig config.Ory,
	hostConfig config.Host,
	storageConfig config.Storage,
	launcherConfig config.Launcher,
) error {
	// Installs ory hydra and kratos
	for _, packageName := range appPackages {
		// Retrieves package definition
		appPackage, err := storeClient.GetPackage(packageName)
		if err != nil {
			return err
		}

		// Checks if the app is already installed and continues if so
		appInstalled, err := docker.IsAppInstalled(dockerClient, packageName)
		if appInstalled {
			appVersion, err := docker.GetAppVersion(dockerClient, packageName)
			if err != nil {
				return err
			}

			// If app is outdated update it
			if launcherConfig.AlwaysUpdate || semver.Compare(appVersion, appPackage.Version) == -1 {
				fmt.Printf("Updating %s\n", packageName)
				err = docker.RemoveContainers(dockerClient, packageName)
				if err != nil {
					return err
				}
				err = TemplateInstall(dockerClient, appPackage, oryConfig, hostConfig, storageConfig)
				if err != nil {
					return err
				}
			} else {
				fmt.Printf("%s is already installed\n", packageName)
				if err := docker.StopApp(dockerClient, packageName); err != nil {
					return err
				}
			}
		} else {
			// Install app if it's not installed
			fmt.Printf("Installing %s\n", packageName)

			err = TemplateInstall(dockerClient, appPackage, oryConfig, hostConfig, storageConfig)
			if err != nil {
				return err
			}
		}

		err = docker.StartApp(dockerClient, packageName)
		if err != nil {
			return err
		}
	}

	return nil
}

func TemplateInstall(
	dockerClient *client.Client,
	appPackage persistence.AppPackage,
	oryConfig config.Ory,
	hostConfig config.Host,
	storageConfig config.Storage,
) error {
	appSchema, err := json.Marshal(appPackage)
	if err != nil {
		return err
	}

	var templatedApp bytes.Buffer
	err = persistence.ApplyAppTemplate(string(appSchema), &templatedApp, appPackage, "", "", oryConfig, hostConfig, storageConfig)
	if err != nil {
		return err
	}
	var parsedTemplate persistence.AppPackage
	if err := json.Unmarshal(templatedApp.Bytes(), &parsedTemplate); err != nil {
		return err
	}

	return docker.InstallApp(dockerClient, parsedTemplate, hostConfig, storageConfig)
}

func FollowLogs(dockerClient *client.Client) error {
	// Follows and prints logs for homecloud container
	containers, err := docker.GetAppContainers(dockerClient, "homecloud.app")
	if err != nil || len(containers) == 0 {
		panic(err)
	}

	containerInspect, err := dockerClient.ContainerInspect(context.Background(), containers[0].ID)
	if err != nil {
		panic(err)
	}

	logs, err := dockerClient.ContainerLogs(
		context.Background(), containers[0].ID, container.LogsOptions{
			Follow:     true,
			ShowStderr: true,
			ShowStdout: true,
		},
	)
	if err != nil {
		panic(err)
	}
	defer logs.Close()

	if containerInspect.Config.Tty {
		_, err = io.Copy(os.Stdout, logs)
	} else {
		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	}
	if err != nil {
		panic(err)
	}
	return nil
}

// ConnectNetworks connects the backend to the ory networks. This assumes the containers are using the expected names
func ConnectNetworks(dockerClient *client.Client) error {
	for _, networkName := range []string{"ory.kratos", "ory.hydra"} {
		err := dockerClient.NetworkConnect(
			context.Background(),
			networkName,
			"homecloud.app-homecloud",
			&network.EndpointSettings{},
		)
		if err != nil && !(errdefs.IsForbidden(err) && strings.Contains(err.Error(), "already exists")) {
			return err
		}
	}

	return nil
}
