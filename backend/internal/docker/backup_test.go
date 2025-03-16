package docker_test

import (
	"context"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/testutils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestBackup(t *testing.T) {
	dockerClient, err := testutils.CreateDindClient()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupDind()

	volumeName := "testing-volume"
	_, err = dockerClient.VolumeCreate(context.Background(), volume.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		t.Fatal(err)
	}

	download, err := dockerClient.ImagePull(context.Background(), "busybox:latest", image.PullOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(io.Discard, download)

	containerResult, err := dockerClient.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: "busybox",
			Cmd:   []string{"/bin/sh", "-c", "echo testing >> /volume/test.txt"},
		},
		&container.HostConfig{
			Binds:      []string{fmt.Sprintf("%s:/volume", volumeName)},
			AutoRemove: true,
		},
		nil,
		nil,
		fmt.Sprintf("%s-writer", volumeName),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = dockerClient.ContainerStart(context.Background(), containerResult.ID, container.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = docker.UntilRemoved(context.Background(), dockerClient, containerResult.ID)
	if err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(testutils.SharedDirectory, volumeName)
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = docker.BackupVolume(context.Background(), dockerClient, volumeName, outputPath)
	if err != nil {
		t.Fatal(err)
	}

	expectedFiles := map[string]string{
		"test.txt": "testing\n",
	}

	testutils.CheckTarArchive(t, filepath.Join(outputPath, volumeName+".tar.gz"), expectedFiles)
}

func TestBackupFolder(t *testing.T) {
	folder := "/tmp/backup_test"
	appFolder := filepath.Join(folder, "test.app", "data")

	testData := map[string]string{
		"test.txt":           "testing_data",
		"test2.txt":          "more test dat",
		"subfolder/test.txt": "subfolder_test",
	}

	for filename, content := range testData {
		if err := os.MkdirAll(filepath.Join(appFolder, filepath.Dir(filename)), 0755); err != nil {
			t.Fatal(err)
		}
		writer, err := os.Create(filepath.Join(appFolder, filename))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.WriteString(content); err != nil {
			t.Fatal(err)
		}
		writer.Close()
	}

	backupFolder := "/tmp/backup_test/output"
	if err := os.MkdirAll(backupFolder, 0755); err != nil {
		t.Fatal(err)
	}

	archive, err := docker.BackupFolder(config.Storage{DataPath: folder}, "test.app", backupFolder)
	if err != nil {
		t.Fatal(err)
	}

	testutils.CheckTarArchive(t, archive, testData)
}
