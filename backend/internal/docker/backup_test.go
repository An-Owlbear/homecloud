package docker_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/testutils"
)

func TestBackup(t *testing.T) {
	dockerClient, err := testutils.CreateDindClient()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupDind()
	defer testutils.CleanupSharedStorage()

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
	if err != nil {
		t.Fatal(err)
	}

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
		"test2.txt":          "more test data",
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

func TestRestoreVolume(t *testing.T) {
	// Sets up docker environment and downloads busybox image
	dockerClient, err := testutils.CreateDindClient()
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.CleanupDind()
	defer testutils.CleanupSharedStorage()

	download, err := dockerClient.ImagePull(context.Background(), "busybox:latest", image.PullOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(io.Discard, download)
	if err != nil {
		t.Fatal(err)
	}

	// Creates archive of data to simulate restoration
	testData := map[string]string{
		"test.txt":           "testing_data",
		"test2.txt":          "more test data",
		"subfolder/test.txt": "subfolder_test",
	}

	testPath := filepath.Join(testutils.SharedDirectory, "test_restore_volume")
	testFilesPath := filepath.Join(testPath, "test_files")
	for filename, content := range testData {
		if err := os.MkdirAll(filepath.Join(testFilesPath, filepath.Dir(filename)), 0755); err != nil {
			t.Fatal(err)
		}
		writer, err := os.Create(filepath.Join(testFilesPath, filename))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.WriteString(content); err != nil {
			t.Fatal(err)
		}
		writer.Close()
	}

	archivePath := filepath.Join(testPath, "test.tar.gz")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gzipWriter := gzip.NewWriter(archiveFile)
	tarWriter := tar.NewWriter(gzipWriter)

	dirfs := os.DirFS(testFilesPath)
	if err := tarWriter.AddFS(dirfs); err != nil {
		t.Fatal(err)
	}
	tarWriter.Close()
	gzipWriter.Close()
	archiveFile.Close()

	// Tests restoring the volume
	err = docker.RestoreVolume(context.Background(), dockerClient, "restore-test", archivePath)
	if err != nil {
		t.Fatal(err)
	}

	// Creates container to run commands from
	cmdContainer, err := dockerClient.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:     "busybox",
			Cmd:       []string{"sh"},
			Tty:       true,
			OpenStdin: true,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: "restore-test",
					Target: "/volume",
				},
			},
			AutoRemove: true,
		},
		nil,
		nil,
		"test-restore-container",
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Starting test container")
	err = dockerClient.ContainerStart(context.Background(), cmdContainer.ID, container.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// For each file attempt to read them from the volume and fail test if the contents don't match
	for filename, expectedContent := range testData {
		fmt.Println("exec test container for file " + filename)
		exec, err := dockerClient.ContainerExecCreate(context.Background(), cmdContainer.ID, container.ExecOptions{
			Tty:          true,
			AttachStdout: true,
			AttachStderr: true,
			Cmd:          []string{"cat", filepath.Join("/volume", filename)},
		})
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("attach to exec test container for file " + filename)
		result, err := dockerClient.ContainerExecAttach(context.Background(), exec.ID, container.ExecAttachOptions{
			Tty: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		content, err := io.ReadAll(result.Reader)
		if err != nil {
			t.Fatal(err)
		}

		trimmedContent := strings.TrimSpace(string(content))
		if trimmedContent != expectedContent {
			t.Fatalf("Expected `%s` for %s but got `%q`", expectedContent, filename, trimmedContent)
		}
	}
}

func TestRestoreFolder(t *testing.T) {
	storageConfig := testutils.SetupTempStorage()

	appData := map[string]string{
		"test.txt":                   "testing_data",
		"subfolder/test.txt":         "subfolder_test",
		"some_other/data/again.json": "{\"key\":\"value\"}",
		"root_again":                 "test",
	}

	archiveDir := filepath.Join(testutils.TempDirectory, "test_restore_app_data")
	archiveFile, err := testutils.CreateTarArchive(archiveDir, appData)
	if err != nil {
		t.Fatal(err)
	}

	appId := "test.app"
	err = docker.RestoreFolder(storageConfig, appId, archiveFile)
	if err != nil {
		t.Fatal(err)
	}

	for filename, expectedContent := range appData {
		reader, err := os.Open(filepath.Join(storageConfig.DataPath, appId, "data", filename))
		if err != nil {
			t.Fatal(err)
		}
		fileContent, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		reader.Close()

		if string(fileContent) != expectedContent {
			t.Fatalf("File content `%s` does not match expected `%s`", string(fileContent), expectedContent)
		}
	}
}
