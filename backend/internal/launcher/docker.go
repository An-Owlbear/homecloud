package launcher

import (
	"context"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/apps"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"os"
)

var appPackages = []string{"ory.kratos", "ory.hydra", "homecloud.app"}

func StartContainers(dockerClient *client.Client, storeClient *apps.StoreClient) error {
	// Installs ory hydra and kratos
	for _, packageName := range appPackages {
		// Checks if the app is already installed and continues if so
		appInstalled, err := docker.IsAppInstalled(dockerClient, packageName)
		if appInstalled {
			fmt.Printf("%s is already installed\n", packageName)
		} else {
			fmt.Printf("Installing %s\n", packageName)
			kratosPackage, err := storeClient.GetPackage(packageName)
			if err != nil {
				return err
			}

			err = docker.InstallApp(dockerClient, kratosPackage)
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

func StopContainers(dockerClient *client.Client) error {
	for _, packageName := range appPackages {
		err := docker.StopApp(dockerClient, packageName)
		if err != nil {
			return err
		}
	}

	return nil
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

	logs, err := dockerClient.ContainerLogs(context.Background(), containers[0].ID, container.LogsOptions{
		Follow:     true,
		ShowStderr: true,
		ShowStdout: true,
	})
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
