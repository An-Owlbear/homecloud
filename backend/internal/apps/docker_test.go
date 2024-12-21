package apps

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/util"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
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
			},
		},
	}

	err = InstallApp(dockerClient, app)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
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
			Image: dindImage,
			ExposedPorts: map[nat.Port]struct{}{
				"2375/tcp": {},
			},
			Entrypoint: []string{"dockerd"},
			// tls=false must be specified to disable a 15 second sleep that occurs otherwise
			Cmd: []string{"--host=0.0.0.0:2375", "--tls=false"},
		},
		&container.HostConfig{
			Privileged: true,
			PortBindings: map[nat.Port][]nat.PortBinding{
				"2375/tcp": {
					{
						HostIP:   "0.0.0.0",
						HostPort: "2375",
					},
				},
			},
		},
		nil,
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
	dockerClient, err = client.NewClientWithOpts(client.WithHost("tcp://0.0.0.0:2375"), client.WithAPIVersionNegotiation())
	if err != nil {
		return
	}
	for i := 0; i < 100; i++ {
		_, err = dockerClient.Ping(context.Background())
		// If the connection is successful return
		if err == nil {
			fmt.Print(i)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	// If the dind container doesn't start in 20 seconds fail
	panic("Dind container didn't start in time")
}

// Cleans up the dind docker container
func CleanupDind() {
	errString := "Failure during cleanup of homecloud-test-container, you may need to remove it manually.\nCaused by error:\n"
	err := hostClient.ContainerStop(context.Background(), "homecloud-test-container", container.StopOptions{})
	if err != nil {
		panic(errString + err.Error())
	}

	err = hostClient.ContainerRemove(context.Background(), "homecloud-test-container", container.RemoveOptions{})
	if err != nil {
		panic(errString + err.Error())
	}
}
