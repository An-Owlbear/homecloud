package apps

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type State string
const (
	ContainerCreated State = "created"
	ContainerRestarting State = "restarting"
	ContainerRunning State = "running"
	ContainerPaused State = "paused"
	ContainerExited State = "exited"
	ContainerDead State = "dead"
)

const APP_ID_LABEL = "AppID"

var TimeoutError = errors.New("Container didn't start in time")

func InstallApp(dockerClient *client.Client, app persistence.AppPackage) error {
	networkVar, err := dockerClient.NetworkCreate(context.Background(), app.Id, network.CreateOptions{
		Labels: map[string]string{APP_ID_LABEL: app.Id},
	})
	if err != nil {
		return err
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{networkVar.ID: {NetworkID: networkVar.ID}},
	}

	for _, containerDef := range app.Containers {
		// Checks if the image is already downloaded and downloads it if not
		alreadyDownloaded, err := isImageDownloaded(dockerClient, containerDef.Image)
		if err != nil {
			return err
		}

		if !alreadyDownloaded {
			reader, err := dockerClient.ImagePull(context.Background(), containerDef.Image, image.PullOptions{})
			if err != nil {
				return err
			}

			_, err = io.ReadAll(reader)
			if err != nil {
				return err
			}
		}

		containerConfig := &container.Config{
			Image:    containerDef.Image,
			Hostname: containerDef.Name,
			Labels:   map[string]string{APP_ID_LABEL: app.Id},
		}

		hostConfig := &container.HostConfig{
			NetworkMode: "bridge",
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
		}

		containerName := app.Id + "-" + containerDef.Name

		result, err := dockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, networkingConfig, nil, containerName)
		if err != nil {
			return err
		}

		err = dockerClient.ContainerStart(context.Background(), result.ID, container.StartOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// StartApp starts all containers related to the app
func StartApp(dockerClient *client.Client, appID string) error {
	// Retrieves filtered list of containers filtered by app ID
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", APP_ID_LABEL, appID))),
	})
	if err != nil {
		return err
	}

	// Starts each contianer
	for _, containerResult := range containers {
		err = dockerClient.ContainerStart(context.Background(), containerResult.ID, container.StartOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// StopApp stops all containers related to the app
func StopApp(dockerClient *client.Client, appID string) error {
	// Retrieves list of containers filterted by app ID
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", APP_ID_LABEL, appID))),
	})
	if err != nil {
		return err
	}

	// Stops each container
	for _, containerResult := range containers {
		err = dockerClient.ContainerStop(context.Background(), containerResult.ID, container.StopOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func UninstallApp(dockerClient *client.Client, appId string) error {
	// Stop the app so containers can be deleted
	err := StopApp(dockerClient, appId)
	if err != nil {
		return err
	}

	// Retrieve a list of containers to delete
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(appFilter(appId)),
	})
	if err != nil {
		return err
	}

	// Ensure each container is stopped before deleting them, along with their volumes
	for _, containerResult := range containers {
		err = UntilState(dockerClient, containerResult.ID, ContainerExited, time.Second * 10, time.Millisecond * 100)
		if err != nil {
			return err
		}

		err = dockerClient.ContainerRemove(context.Background(), containerResult.ID, container.RemoveOptions{RemoveVolumes: true})
		if err != nil {
			return err
		}
	}

	networks, err := dockerClient.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(appFilter(appId)),
	})
	if err != nil {
		return err
	}

	for _, networkResult := range networks {
		err := dockerClient.NetworkRemove(context.Background(), networkResult.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAppContainers retrieves a list of all containers belonging to the specified app
func GetAppContainers(dockerClient *client.Client, appId string) (containers []types.Container, err error) {
	return dockerClient.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(appFilter(appId)),
	})
}

// UntilState waits until the given container is started
func UntilState(dockerClient *client.Client, containerId string, state State, timeout time.Duration, interval time.Duration) error {
	currentTimeout := time.Duration(0)
	for currentTimeout < timeout {
		info, err := dockerClient.ContainerInspect(context.Background(), containerId)
		if err != nil {
			return err
		}

		if info.State.Status == string(state) {
			return nil
		}

		time.Sleep(interval)
		currentTimeout += interval
	}

	return TimeoutError
}

// simple function for creating 
func appFilter(appId string) filters.KeyValuePair {
	return filters.Arg("label", fmt.Sprintf("%s=%s", APP_ID_LABEL, appId))
}

// Checks if the image is downloaded already
func isImageDownloaded(dockerClient *client.Client, imageName string) (alreadyDownloaded bool, err error) {
	images, err := dockerClient.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				alreadyDownloaded = true
				return
			}
		}
	}

	return
}
