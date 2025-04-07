package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
)

type State string

const (
	ContainerCreated    State = "created"
	ContainerRestarting State = "restarting"
	ContainerRunning    State = "running"
	ContainerPaused     State = "paused"
	ContainerExited     State = "exited"
	ContainerDead       State = "dead"
)

const APP_ID_LABEL = "AppID"
const AppVersionLabel = "AppVersion"

var NotFoundError = errors.New("no containers for app found")
var InvalidContainerError = errors.New("container has invalid configuration")

func InstallApp(
	dockerClient *client.Client,
	app persistence.AppPackage,
	serverHostConfig config.Host,
	storageConfig config.Storage,
) error {
	// Creates the network if it doesn't already exist
	networkId, err := GetOrCreateNetwork(context.Background(), dockerClient, app.Id, map[string]string{
		APP_ID_LABEL:    app.Id,
		AppVersionLabel: app.Version,
	})
	if err != nil {
		return err
	}

	for _, containerDef := range app.Containers {
		// Checks if the image is already downloaded and downloads it if not
		alreadyDownloaded, err := IsImageDownloaded(dockerClient, containerDef.Image)
		if err != nil {
			return err
		}

		// If not downloaded the image
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

		// sets up the environment variables for the container
		var env []string
		for key, value := range containerDef.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}

		containerConfig := &container.Config{
			Image:    containerDef.Image,
			Hostname: containerDef.Name,
			Env:      env,
			Labels: map[string]string{
				APP_ID_LABEL:    app.Id,
				AppVersionLabel: app.Version,
			},
		}

		// Sets the command if given
		if containerDef.Command != "" {
			containerConfig.Cmd = strings.Split(containerDef.Command, " ")
		}

		// creates port mappings
		portMap := nat.PortMap{}
		for _, ports := range containerDef.Ports {
			portParts := strings.Split(ports, ":")
			var containerPort = portParts[1]
			if !strings.HasSuffix(containerPort, "/tcp") && !strings.HasSuffix(containerPort, "/udp") {
				containerPort += "/tcp"
			}
			portMap[nat.Port(containerPort)] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: portParts[0],
				},
			}
		}

		// creates required volumes
		formattedVolumes := []string{}
		for _, vol := range containerDef.Volumes {
			volumeParts := strings.Split(vol, ":")

			// mounts to app specific data folder if mount is local file
			if strings.HasPrefix(volumeParts[0], "./") {
				// local files are converted to map to the data folder - e.g. ./config.json becomes data_dir/app_id/data/config.json
				volumeParts[0] = filepath.Join(storageConfig.GetAppDataMountPath(app.Id), volumeParts[0][2:])
			} else if strings.HasPrefix(volumeParts[0], "!AppDir") {
				appDir, err := os.Getwd()
				if err != nil {
					return err
				}
				volumeParts[0] = filepath.Join(appDir, strings.TrimPrefix(volumeParts[0], "!AppDir"))
			} else if !strings.HasPrefix(volumeParts[0], "/") {
				// Checks if the volume exists before creating
				volumeParts[0] = fmt.Sprintf("%s-%s", app.Id, volumeParts[0])
				if _, err = dockerClient.VolumeInspect(context.Background(), volumeParts[0]); err != nil {
					_, err = dockerClient.VolumeCreate(context.Background(), volume.CreateOptions{
						Name:   volumeParts[0],
						Labels: map[string]string{APP_ID_LABEL: app.Id},
					})
					if err != nil {
						return err
					}
				}
			}

			formattedVolumes = append(formattedVolumes, strings.Join(volumeParts, ":"))
		}

		// Gets restart policy, defaults to always
		restart := container.RestartPolicyAlways
		if containerDef.Restart != "" {
			restart = container.RestartPolicyMode(containerDef.Restart)
		}

		hostConfig := &container.HostConfig{
			NetworkMode: "bridge",
			RestartPolicy: container.RestartPolicy{
				Name: restart,
			},
			PortBindings: portMap,
			Binds:        formattedVolumes,
			ExtraHosts: append(containerDef.ExtraHosts,
				fmt.Sprintf("hydra.%s:host-gateway", serverHostConfig.Host),
				fmt.Sprintf("%s:host-gateway", serverHostConfig.Host),
			),
			Privileged: containerDef.Privileged,
		}

		containerName := app.Id + "-" + containerDef.Name

		// sets up the network config for the container
		networkingConfig := &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{networkId: {NetworkID: networkId}},
		}
		if containerDef.ProxyTarget {
			// adds the network shared with homecloud api to the network config
			proxyNetworkId, err := GetOrCreateNetwork(context.Background(), dockerClient, app.Id+"-proxy", map[string]string{
				APP_ID_LABEL:    app.Id,
				AppVersionLabel: app.Version,
			})
			if err != nil {
				return err
			}

			// Ensures main container is connected to proxy network
			if err := dockerClient.NetworkConnect(context.Background(), proxyNetworkId, "homecloud.app-homecloud", &network.EndpointSettings{}); err != nil {
				return err
			}
			networkingConfig.EndpointsConfig[proxyNetworkId] = &network.EndpointSettings{NetworkID: proxyNetworkId}
		}

		_, err = dockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, networkingConfig, nil, containerName)
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

	// Starts each container
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
	// Retrieves list of containers filtered by app ID
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

func RemoveContainers(dockerClient *client.Client, appId string) error {
	// Stop the app so containers can be deleted
	err := StopApp(dockerClient, appId)
	if err != nil {
		return err
	}

	// Retrieve a list of containers to delete
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(AppFilter(appId)),
	})
	if err != nil {
		return err
	}

	// Ensure each container is stopped before deleting them, along with their volumes
	for _, containerResult := range containers {
		err = UntilState(dockerClient, containerResult.ID, ContainerExited, time.Second*10, time.Millisecond*100)
		if err != nil {
			return err
		}

		err = dockerClient.ContainerRemove(context.Background(), containerResult.ID, container.RemoveOptions{RemoveVolumes: true})
		if err != nil {
			return err
		}
	}

	return nil
}

func UninstallApp(dockerClient *client.Client, appId string) error {
	// removes containers
	err := RemoveContainers(dockerClient, appId)
	if err != nil {
		return err
	}

	networks, err := dockerClient.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(AppFilter(appId)),
	})
	if err != nil {
		return err
	}

	for _, networkResult := range networks {
		// Retrieves info from network inspect, this is necessary because for some reason networks structs from
		// NetworkList don't have information of attached containers despite having the field
		networkInspect, err := dockerClient.NetworkInspect(context.Background(), networkResult.ID, network.InspectOptions{})
		if err != nil {
			return err
		}

		// Disconnects all containers still attached, currently this will only be Homecloud attached to the proxy network
		for containerID := range networkInspect.Containers {
			err = dockerClient.NetworkDisconnect(context.Background(), networkResult.ID, containerID, true)
			if err != nil {
				return fmt.Errorf("failed to disconnect container %s: %w", containerID, err)
			}
		}

		// Wait until containers have been removed
		err = util.WaitUntil(func() (bool, error) {
			networkInfo, err := dockerClient.NetworkInspect(context.Background(), networkResult.ID, network.InspectOptions{})
			if err != nil {
				return false, err
			}
			return len(networkInfo.Containers) == 0, nil
		}, time.Second*10, time.Millisecond*50)
		if err != nil {
			networkInspect, _ := dockerClient.NetworkInspect(context.Background(), networkResult.ID, network.InspectOptions{})
			fmt.Println(networkInspect.Containers)
			return err
		}

		// Remove network
		err = dockerClient.NetworkRemove(context.Background(), networkResult.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAppContainers retrieves a list of all containers belonging to the specified app
func GetAppContainers(dockerClient *client.Client, appId string) (containers []types.Container, err error) {
	return dockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(AppFilter(appId)),
	})
}

// UntilState waits until the given container is started
func UntilState(
	dockerClient *client.Client,
	containerId string,
	state State,
	timeout time.Duration,
	interval time.Duration,
) error {
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

	return fmt.Errorf("%w: container didn't reach state %s in  time", util.TimeoutError, state)
}

// AppFilter simple function for creating
func AppFilter(appId string) filters.KeyValuePair {
	return filters.Arg("label", fmt.Sprintf("%s=%s", APP_ID_LABEL, appId))
}

// IsImageDownloaded Checks if the image is downloaded already
func IsImageDownloaded(dockerClient *client.Client, imageName string) (alreadyDownloaded bool, err error) {
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

// IsAppInstalled checks if an app of the given ID is installed
func IsAppInstalled(dockerClient *client.Client, appId string) (installed bool, err error) {
	containers, err := GetAppContainers(dockerClient, appId)
	if err != nil {
		return
	}

	installed = len(containers) > 0
	return
}

// IsAppRunning checks if an app is running
func IsAppRunning(dockerClient *client.Client, appId string) (running bool, err error) {
	containers, err := GetAppContainers(dockerClient, appId)
	if err != nil {
		return
	}

	if len(containers) == 0 {
		return false, errors.New("no containers found")
	}

	for _, container := range containers {
		if container.State != string(ContainerRunning) {
			return false, nil
		}
	}

	return true, nil
}

// GetAppVersion retrieves the version of the app based off the docker container metadata
func GetAppVersion(dockerClient *client.Client, appId string) (version string, err error) {
	containers, err := GetAppContainers(dockerClient, appId)
	if err != nil {
		return
	}

	if len(containers) == 0 {
		return "", NotFoundError
	}

	appVersion, ok := containers[0].Labels[AppVersionLabel]
	if !ok {
		return "", InvalidContainerError
	}

	return appVersion, nil
}
