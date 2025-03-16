package docker_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/testutils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/google/go-cmp/cmp"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	fileReader, err := os.Open(filepath.Join(outputPath, volumeName+".tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		t.Fatal(err)
	}
	tarReader := tar.NewReader(gzipReader)

	expectedFiles := map[string]string{
		"test.txt": "testing\n",
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			filename := strings.TrimPrefix(header.Name, "./")
			fmt.Println(filename)
			fmt.Println(expectedFiles)
			expected, ok := expectedFiles[filename]
			if !ok {
				t.Fatalf("file %s in tar not expected", filename)
			}
			fileContents, err := io.ReadAll(tarReader)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(expected, string(fileContents)); diff != "" {
				t.Fatal(diff)
			}
			delete(expectedFiles, filename)
		}
	}

	if len(expectedFiles) > 0 {
		t.Fatalf("did not find expected files: %v", expectedFiles)
	}
}
