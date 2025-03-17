package docker

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

func BackupVolume(
	ctx context.Context,
	dockerClient *client.Client,
	volumeName string,
	outputDir string,
) error {
	containerConfig := container.Config{
		Image: "busybox",
		Cmd:   []string{"tar", "-czf", fmt.Sprintf("/backup/%s.tar.gz", volumeName), "-C", "/target", "."},
	}

	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   outputDir,
				Target:   "/backup",
				ReadOnly: false,
			},
			{
				Type:     mount.TypeVolume,
				Source:   volumeName,
				Target:   "/target",
				ReadOnly: true,
			},
		},
		AutoRemove: true,
	}

	result, err := dockerClient.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, fmt.Sprintf("%s-backup", volumeName))
	if err != nil {
		return fmt.Errorf("failed creating volume backup container for %s: %w", volumeName, err)
	}

	err = dockerClient.ContainerStart(ctx, result.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed starting volume backup container for %s: %w", volumeName, err)
	}

	err = UntilRemoved(ctx, dockerClient, result.ID)

	return nil
}

func BackupFolder(storageConfig config.Storage, appId string, outputDir string) (string, error) {
	outputPath := filepath.Join(outputDir, "data.tar.gz")
	output, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed creating output file %s: %w", outputPath, err)
	}
	defer output.Close()
	gw := gzip.NewWriter(output)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	dirfs := os.DirFS(filepath.Join(storageConfig.DataPath, appId, "data"))
	if err := tw.AddFS(dirfs); err != nil {
		return "", fmt.Errorf("failed adding folder %s to tarball: %w", outputPath, err)
	}
	return outputPath, nil
}

func BackupAppData(
	ctx context.Context,
	dockerClient *client.Client,
	storageConfig config.Storage,
	appId string,
	outputDir string,
) error {
	if _, err := BackupFolder(storageConfig, appId, outputDir); err != nil {
		return fmt.Errorf("failed creating backup of non volume app data: %w", err)
	}

	appContainers, err := GetAppContainers(dockerClient, appId)
	if err != nil {
		return fmt.Errorf("failed getting app containers for %s during backup: %w", appId, err)
	}

	for _, appContainer := range appContainers {
		for _, appMount := range appContainer.Mounts {
			if appMount.Type == mount.TypeVolume {
				if err := BackupVolume(ctx, dockerClient, appMount.Name, outputDir); err != nil {
					return fmt.Errorf("failed backing up volume %s for %s: %w", appMount.Name, appId, err)
				}
			}
		}
	}

	return nil
}
