package testutils

import (
	"context"
	"errors"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"io"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/util"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/joho/godotenv"
)

var hostClient *client.Client

var BasicHostConfig = config.Host{
	Host:        "example.com",
	Port:        80,
	HTTPS:       true,
	PortForward: true,
}

// CreateDindClient creates a containerised docker environment for testing. This environment should be
// removed using CleanupDocker at the end
// This method is much slower and therefore may be worse for normal development. Investigate
// options other using local docker environment depending on a test option
func CreateDindClient() (dockerClient *client.Client, err error) {
	// Loads from environment variables to reach the host docker environment
	godotenv.Load(filepath.Join(util.RootDir(), ".env"))
	hostClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return
	}

	// creates the network and attaches the current container to it
	networkRes, err := hostClient.NetworkCreate(context.Background(), "homecloud-tester", network.CreateOptions{})
	if err != nil {
		return
	}

	err = hostClient.NetworkConnect(context.Background(), networkRes.ID, os.Getenv("TEST_CONTAINER_NAME"), &network.EndpointSettings{})
	if err != nil {
		return
	}

	// Downloads the dind image if needed
	dindImage := "docker:27.4.1-dind"
	isDownloaded, err := docker.IsImageDownloaded(hostClient, dindImage)
	if err != nil {
		panic(err.Error())
	}
	if !isDownloaded {
		reader, err := hostClient.ImagePull(context.Background(), dindImage, image.PullOptions{})
		if err != nil {
			panic(err.Error())
		}
		io.ReadAll(reader)
	}

	// Creates the dind container
	dindContainer, err := hostClient.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:      dindImage,
			Entrypoint: []string{"dockerd"},
			// tls=false must be specified to disable a 15 second sleep that occurs otherwise
			Cmd: []string{"--host=0.0.0.0:2375", "--tls=false"},
		},
		&container.HostConfig{
			Privileged:  true,
			NetworkMode: "bridge",
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkRes.ID: {NetworkID: networkRes.ID},
			},
		},
		nil,
		"homecloud-test-container",
	)
	if err != nil {
		return
	}

	// Starts the dind container
	err = hostClient.ContainerStart(context.Background(), dindContainer.ID, container.StartOptions{})
	if err != nil {
		return
	}

	// Waits until the docker runtime inside dind has started before returning
	dockerClient, err = client.NewClientWithOpts(client.WithHost("tcp://homecloud-test-container:2375"), client.WithAPIVersionNegotiation())
	if err != nil {
		return
	}

	for i := 0; i < 200; i++ {
		_, err = dockerClient.Ping(context.Background())
		// If the connection is successful return
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		// If the dind container doesn't start in 20 seconds fail
		err = errors.New("dind container didn't start in time")
		return
	}

	// Creates network on dind client
	_, err = dockerClient.NetworkCreate(context.Background(), "homecloud.app", network.CreateOptions{})
	if err != nil {
		err = fmt.Errorf("couldn't created network in dind container: %w", err)
		return
	}

	return
}

// CleanupDind Cleans up the dind docker container
func CleanupDind() {
	errString := "Failure during cleanup of homecloud-test-container, you may need to remove it manually.\nCaused by error:\n"
	err := hostClient.ContainerStop(context.Background(), "homecloud-test-container", container.StopOptions{})
	if err != nil && !client.IsErrNotFound(err) {
		panic(errString + err.Error())
	}

	err = hostClient.ContainerRemove(context.Background(), "homecloud-test-container", container.RemoveOptions{})
	if err != nil && !client.IsErrNotFound(err) {
		panic(errString + err.Error())
	}

	err = hostClient.NetworkDisconnect(context.Background(), "homecloud-tester", os.Getenv("TEST_CONTAINER_NAME"), true)
	if err != nil && !client.IsErrNotFound(err) {
		panic(errString + err.Error())
	}

	err = hostClient.NetworkRemove(context.Background(), "homecloud-tester")
	if err != nil && !client.IsErrNotFound(err) {
		panic(errString + err.Error())
	}
}

func HelpTestAppPackage(dockerClient *client.Client, app persistence.AppPackage, storageConfig config.Storage, t *testing.T) {
	// Retrieves the network created for the app to check it and for later use
	networks, err := dockerClient.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", docker.APP_ID_LABEL, app.Id))),
	})
	if err != nil {
		t.Fatalf("Unexpected error search networking: %s", err.Error())
	}
	if len(networks) != 1 {
		t.Fatalf("Unexpected number of tagged networks found: %d", len(networks))
	}
	appNetwork := networks[0]

	// Checks there's the correct number of applications and that their information is as expected
	results, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", docker.APP_ID_LABEL, app.Id))),
	})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if len(results) != 1 {
		t.Fatalf("Unexpected number of tagged containers %d, should be %d", len(results), 1)
	}

	result, err := dockerClient.ContainerInspect(context.Background(), results[0].ID)
	if err != nil {
		t.Fatalf("Unexpected error inspecting cotainer: %s", err.Error())
	}

	// Checks name is correct
	expectedName := app.Id + "-" + app.Containers[0].Name
	if result.Name != "/"+expectedName {
		t.Fatalf("Unexpected container Names %s, should be %s", result.Name, expectedName)
	}

	// Checks the network configuration is correct
	// TODO: update network testing to be more generalised
	if networkIds := slices.Collect(maps.Keys(result.NetworkSettings.Networks)); networkIds[0] == appNetwork.ID {
		t.Fatalf("Incorrect network ID set in container\nExpected value: %s\nActual value: %s", appNetwork.ID, networkIds[0])
	}

	// Tests environment variables are correct
	envVars := result.Config.Env
	// Creates a list of env vars with the path variable removed
	for i, envVar := range envVars {
		if strings.HasPrefix(envVar, "PATH=") {
			envVars = append(envVars[:i], envVars[i+1:]...)
		}
	}

	expectedEnv := []string{}
	for k, v := range app.Containers[0].Environment {
		expectedEnv = append(expectedEnv, fmt.Sprintf("%s=%s", k, v))
	}
	if !reflect.DeepEqual(envVars, expectedEnv) {
		t.Fatalf("Incorrect container environment variables\nExpected values: %+v\nActual values:%+v", expectedEnv, envVars)
	}

	// Tests the correct port mappings are done correctly
	expectedPortMap := nat.PortMap{}
	for _, portString := range app.Containers[0].Ports {
		portParts := strings.Split(portString, ":")
		containerPort := portParts[1]
		if !strings.HasSuffix(containerPort, "/tcp") || !strings.HasSuffix(containerPort, "/udp") {
			containerPort += "/tcp"
		}
		expectedPortMap[nat.Port(containerPort)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: portParts[0],
			},
		}
	}

	if !reflect.DeepEqual(result.NetworkSettings.Ports, expectedPortMap) {
		t.Fatalf("Incorrect port mappings\nExpected: %+v\nActual: %+v", expectedPortMap, result.NetworkSettings.Ports)
	}

	// Tests volumes are properly created and bound
	expectedVolumeNum := 0
	for _, vol := range app.Containers[0].Volumes {
		if !strings.HasPrefix(vol, "./") && !strings.HasPrefix(vol, ".") {
			expectedVolumeNum++
		}
	}

	volumes, err := dockerClient.VolumeList(context.Background(), volume.ListOptions{
		Filters: filters.NewArgs(docker.AppFilter(app.Id)),
	})
	if len(volumes.Volumes) != expectedVolumeNum {
		t.Fatalf("Invalid number of volumes\nExpected: %d\nActual: %d", expectedVolumeNum, len(volumes.Volumes))
	}

	// Creates k,v map of volumes for testing
	checkVols := make(map[string]string)
	for _, vol := range app.Containers[0].Volumes {
		volParts := strings.Split(vol, ":")
		if strings.HasPrefix(volParts[0], "./") {
			execPath, err := os.Getwd()
			if err != nil {
				t.Fatalf("Unexpected error getting exec path: %s", err.Error())
			}
			volParts[0] = filepath.Join(execPath, storageConfig.DataPath, app.Id, "data", volParts[0][2:])
		}
		checkVols[volParts[0]] = volParts[1]
	}

	// Tests all created volumes with the app label are valid
	volListCheck := make(map[string]string)
	for k, v := range checkVols {
		volListCheck[k] = v
	}
	for _, volResult := range volumes.Volumes {
		if _, ok := volListCheck[volResult.Name]; !ok {
			t.Fatalf("Volume not one of the expected volumes: %s", volResult.Name)
		}
		delete(volListCheck, volResult.Name)
	}

	// Checks the volume bindings on the container are as expected
	containerVolCheck := make(map[string]string)
	for k, v := range checkVols {
		containerVolCheck[k] = v
	}
	for _, containerVol := range result.Mounts {
		var volName string
		if containerVol.Type == mount.TypeVolume {
			volName = containerVol.Name
		} else if containerVol.Type == mount.TypeBind {
			volName = containerVol.Source
		} else {
			t.Fatalf("Unexpected type of volume: %s", containerVol.Type)
		}

		if mountLoc, ok := containerVolCheck[volName]; !ok {
			t.Fatalf("Volume not one of the expected volumes: %s", volName)
		} else if containerVol.Destination != mountLoc {
			t.Fatalf("Incorrect mount location for volume: %s\nExpected: %s\nActual: %s", volName, mountLoc, containerVol.Destination)
		}
		delete(containerVolCheck, containerVol.Name)
	}
}
