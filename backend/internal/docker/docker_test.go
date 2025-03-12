package docker_test

import (
	"context"
	"testing"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/testutils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

func TestInstallApp(t *testing.T) {
	dockerClient, err := testutils.CreateDindClient()
	defer testutils.CleanupDind()
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

	storageConfig := testutils.SetupTempStorage()
	err = docker.InstallApp(dockerClient, app, testutils.BasicHostConfig, storageConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	testutils.HelpTestAppPackage(dockerClient, app, storageConfig, t)
}

func TestStopApp(t *testing.T) {
	dockerClient, err := testutils.CreateDindClient()
	defer testutils.CleanupDind()
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

	storageConfig := testutils.SetupTempStorage()
	err = docker.InstallApp(dockerClient, app, testutils.BasicHostConfig, storageConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	results, err := docker.GetAppContainers(dockerClient, app.Id)
	for _, result := range results {
		err = docker.UntilState(dockerClient, result.ID, docker.ContainerRunning, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error after starting containers: %s", err.Error())
		}
	}

	err = docker.StopApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error stopping app: %s", err.Error())
	}

	for _, result := range results {
		err = docker.UntilState(dockerClient, result.ID, docker.ContainerExited, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error after stopping containers: %s", err.Error())
		}
	}

	results, err = docker.GetAppContainers(dockerClient, app.Id)
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
	dockerClient, err := testutils.CreateDindClient()
	defer testutils.CleanupDind()
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

	storageConfig := testutils.SetupTempStorage()
	err = docker.InstallApp(dockerClient, app, testutils.BasicHostConfig, storageConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	results, err := docker.GetAppContainers(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error after installing app: %s", err.Error())
	}

	for _, result := range results {
		err = docker.UntilState(dockerClient, result.ID, docker.ContainerRunning, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error waiting for containers to start: %s", err.Error())
		}
	}

	err = docker.StopApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error stopping app: %s", err.Error())
	}

	for _, result := range results {
		err = docker.UntilState(dockerClient, result.ID, docker.ContainerExited, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error waiting for containers to start: %s", err.Error())
		}
	}

	err = docker.StartApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error starting app: %s", err.Error())
	}

	for _, result := range results {
		err = docker.UntilState(dockerClient, result.ID, docker.ContainerRunning, time.Second*10, time.Millisecond*100)
		if err != nil {
			t.Fatalf("Unexpected error waiting for containers to start: %s", err.Error())
		}
	}

	results, err = docker.GetAppContainers(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Unexpected error retrieving containers: %s", err.Error())
	}

	for _, result := range results {
		if result.State != string(docker.ContainerRunning) {
			t.Fatalf("Container %s is not started, its status is %s", result.Names[0], result.State)
		}
	}
}

func TestUninstallApp(t *testing.T) {
	dockerClient, err := testutils.CreateDindClient()
	defer testutils.CleanupDind()
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

	storageConfig := testutils.SetupTempStorage()
	err = docker.InstallApp(dockerClient, app, testutils.BasicHostConfig, storageConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	err = docker.UninstallApp(dockerClient, app.Id)
	if err != nil {
		t.Fatalf("Error uninstall app: %s", err.Error())
	}

	results, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(docker.AppFilter(app.Id)),
	})
	if err != nil {
		t.Fatalf("Unexpected error after uninstall app: %s", err.Error())
	}

	if len(results) != 0 {
		t.Fatalf("Containers found when they're supposed to be deleted")
	}

	networks, err := dockerClient.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(docker.AppFilter(app.Id)),
	})
	if err != nil {
		t.Fatalf("Error occured when finding networks: %s", err.Error())
	}

	if len(networks) != 0 {
		t.Fatalf("Networks belonging to app found when they're supposed to be removed")
	}
}
