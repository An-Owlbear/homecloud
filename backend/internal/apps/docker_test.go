package apps

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

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

func TestInstallApp(t *testing.T) {
	dockerClient, err := CreateDindClient()
	defer CleanupDind()
	if err != nil {
		t.Fatalf("Unexpected error setting up docker: %s", err.Error())
	}

	app := persistence.AppPackage{
		Schema:      "v1.0",
		Version:     "v1.5",
		Id:          "traefik.whoami",
		Name:        "whoami",
		Author:      "traefik",
		Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		Containers: []persistence.PackageContainer{
			{
				Name:        "whoami",
				Image:       "traefik/whoami:v1.10.3",
				ProxyTarget: true,
				ProxyPort:   "80",
				Ports:       []string{"8000:80"},
				Environment: map[string]string{
					"test_env": "value",
				},
				Volumes: []string{
					"test_vol:/opt/bind1",
					"test_vol2:/opt/bind2",
					"./docker_test.go:/opt/random_file.go",
				},
			},
		},
	}

	err = InstallApp(dockerClient, app)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	HelpTestAppPackage(dockerClient, app, t)
}

func HelpTestAppPackage(dockerClient *client.Client, app persistence.AppPackage, t *testing.T) {
	// Retrieves the network created for the app to check it and for later use
	networks, err := dockerClient.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", APP_ID_LABEL, app.Id))),
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
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", APP_ID_LABEL, app.Id))),
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
		t.Fatalf("Incorrect container environment variables\nExpected values: %+v\nActual values:%+v", envVars, expectedEnv)
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
		Filters: filters.NewArgs(appFilter(app.Id)),
	})
	if len(volumes.Volumes) != expectedVolumeNum {
		t.Fatalf("Invalid number of volumes\nExpected: %d\nActual: %d", expectedVolumeNum, len(volumes.Volumes))
	}

	// Creates k,v map of volumes for testing
	checkVols := make(map[string]string)
	for _, vol := range app.Containers[0].Volumes {
		volParts := strings.Split(vol, ":")
		if strings.HasPrefix(volParts[0], "./") {
			execPath, err := os.Executable()
			if err != nil {
				t.Fatalf("Unexpected error gettign exec path: %s", err.Error())
			}
			volParts[0] = filepath.Join(execPath, volParts[0][2:])
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

func TestStopApp(t *testing.T) {
	dockerClient, err := CreateDindClient()
	defer CleanupDind()
	if err != nil {
		t.Fatalf("Unexpected error setting up docker: %s", err.Error())
	}

	app := persistence.AppPackage{
		Schema:      "v1.0",
		Version:     "v1.5",
		Id:          "traefik.whoami",
		Name:        "whoami",
		Author:      "traefik",
		Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		Containers: []persistence.PackageContainer{
			{
				Name:        "whoami",
				Image:       "traefik/whoami:v1.10.3",
				ProxyTarget: true,
				ProxyPort:   "80",
			},
		},
	}

	err = InstallApp(dockerClient, app)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	results, err := GetAppContainers(dockerClient, app.Id)
	for _, result := range results {
		err = UntilState(dockerClient, result.ID, ContainerRunning, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error after starting containers: %s", err.Error())
		}
	}

	err = StopApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error stopping app: %s", err.Error())
	}

	for _, result := range results {
		err = UntilState(dockerClient, result.ID, ContainerExited, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error after stopping containers: %s", err.Error())
		}
	}

	results, err = GetAppContainers(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error retriving stopped containers: %s", err.Error())
	}

	for _, result := range results {
		if result.State != "exited" {
			t.Fatalf("Container %s is not stopped, its status is %s", result.Names[0], result.State)
		}
	}
}

func TestStartApp(t *testing.T) {
	dockerClient, err := CreateDindClient()
	defer CleanupDind()
	if err != nil {
		t.Fatalf("Unexpected error setting up docker: %s", err.Error())
	}

	app := persistence.AppPackage{
		Schema:      "v1.0",
		Version:     "v1.5",
		Id:          "traefik.whoami",
		Name:        "whoami",
		Author:      "traefik",
		Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		Containers: []persistence.PackageContainer{
			{
				Name:        "whoami",
				Image:       "traefik/whoami:v1.10.3",
				ProxyTarget: true,
				ProxyPort:   "80",
			},
		},
	}

	err = InstallApp(dockerClient, app)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	results, err := GetAppContainers(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error after installing app: %s", err.Error())
	}

	for _, result := range results {
		err = UntilState(dockerClient, result.ID, ContainerRunning, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error waiting for containers to start: %s", err.Error())
		}
	}

	err = StopApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error stopping app: %s", err.Error())
	}

	for _, result := range results {
		err = UntilState(dockerClient, result.ID, ContainerExited, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error waiting for containers to start: %s", err.Error())
		}
	}

	err = StartApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error starting app: %s", err.Error())
	}

	for _, result := range results {
		err = UntilState(dockerClient, result.ID, ContainerRunning, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error waiting for containers to start: %s", err.Error())
		}
	}

	results, err = GetAppContainers(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error retrieving containers: %s", err.Error())
	}

	for _, result := range results {
		if result.State != string(ContainerRunning) {
			t.Fatalf("Container %s is not started, its status is %s", result.Names[0], result.State)
		}
	}
}

func TestUninstallApp(t *testing.T) {
	dockerClient, err := CreateDindClient()
	defer CleanupDind()
	if err != nil {
		t.Fatalf("Unexpected error setting up container: %s", err.Error())
	}

	app := persistence.AppPackage{
		Schema:      "v1.0",
		Version:     "v1.5",
		Id:          "traefik.whoami",
		Name:        "whoami",
		Author:      "traefik",
		Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		Containers: []persistence.PackageContainer{
			{
				Name:        "whoami",
				Image:       "traefik/whoami:v1.10.3",
				ProxyTarget: true,
				ProxyPort:   "80",
			},
		},
	}

	err = InstallApp(dockerClient, app)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	err = UninstallApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Error uninstall app: %s", err.Error())
	}

	results, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(appFilter(app.Id)),
	})
	if err != nil {
		t.Fatalf("Unexpected error after uninstall app: %s", err.Error())
	}

	if len(results) != 0 {
		t.Fatalf("Containers found when they're supposed to be deleted")
	}

	networks, err := dockerClient.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(appFilter(app.Id)),
	})
	if err != nil {
		t.Fatalf("Error occured when finding networks: %s", err.Error())
	}

	if len(networks) != 0 {
		t.Fatalf("Networks belonging to app found when they're supposed to be removed")
	}
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
	isDownloaded, err := isImageDownloaded(hostClient, dindImage)
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
		err = errors.New("Dind container didn't start in time")
		return
	}

	err = EnsureProxyNetwork(dockerClient)
	if err != nil {
		return
	}

	return
}

// Cleans up the dind docker container
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
