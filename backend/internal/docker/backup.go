package docker

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

// BackupVolume backs up the contents of the specified docker volume to the specified local host directory
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

	// Configures mounts for the volume and specified local directory
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

	// Creates and starts a minimal container to back up the volume
	result, err := dockerClient.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, fmt.Sprintf("%s-backup", volumeName))
	if err != nil {
		return fmt.Errorf("failed creating volume backup container for %s: %w", volumeName, err)
	}

	err = dockerClient.ContainerStart(ctx, result.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed starting volume backup container for %s: %w", volumeName, err)
	}

	// Waits until the containers are removed to ensure the backups are complete
	err = UntilRemoved(ctx, dockerClient, result.ID)

	return nil
}

// BackupFolder backs up the data directory of the specified app to the specified local host directory
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

// BackupAppData backs up all volumes and the data directory for the given app
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

// RestoreVolume restores data from the specified backup to the named docker volume
func RestoreVolume(
	ctx context.Context,
	dockerClient *client.Client,
	volumeName string,
	backupPath string,
) error {
	containerConfig := container.Config{
		Image: "busybox",
		Cmd:   []string{"tar", "-xzf", "/backup/volume.tar.gz", "-C", "/target"},
	}

	// Mounts backup location and target volume
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: backupPath,
				Target: "/backup/volume.tar.gz",
			},
			{
				Type:   mount.TypeVolume,
				Source: volumeName,
				Target: "/target",
			},
		},
		AutoRemove: true,
	}

	// Creates and starts minimal containers for restoration
	result, err := dockerClient.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, fmt.Sprintf("%s-restore", volumeName))
	if err != nil {
		return fmt.Errorf("failed creating volume restore container for %s: %w", volumeName, err)
	}

	err = dockerClient.ContainerStart(ctx, result.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed starting volume restore container for %s: %w", volumeName, err)
	}

	// Waits until containers finish restoring data and are removed
	err = UntilRemoved(ctx, dockerClient, result.ID)
	if err != nil {
		return fmt.Errorf("failed removing volume restore container for %s: %w", volumeName, err)
	}

	return nil
}

// RestoreFolder restores data from the specified backup to an app's data folder
func RestoreFolder(storageConfig config.Storage, appId string, backupPath string) error {
	// Ensures path to restore to exists
	restorePath := filepath.Join(storageConfig.DataPath, appId, "data")
	if err := os.MkdirAll(restorePath, 0755); err != nil {
		return fmt.Errorf("error creating app folder to restore backup: %w", err)
	}

	// Opens backup archive
	fileReader, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("error opening backup file %s: %w", backupPath, err)
	}
	defer fileReader.Close()
	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	// Restore archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("error reading tar entry: %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filepath.Join(restorePath, header.Name), 0755); err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Join(restorePath, filepath.Dir(header.Name)), 0755); err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
			outFile, err := os.Create(filepath.Join(restorePath, header.Name))
			if err != nil {
				return fmt.Errorf("error creating file: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("error writing file: %w", err)
			}
			outFile.Close()
		}
	}

	return nil
}

// RestoreAppData restores the application folder data and volumes from a backup. Assume the containers are already
// removed and that the app data folder has been cleared
func RestoreAppData(
	ctx context.Context,
	dockerClient *client.Client,
	storageConfig config.Storage,
	appId string,
	backupPath string,
) error {
	err := RestoreFolder(storageConfig, appId, filepath.Join(backupPath, "data.tar.gz"))
	if err != nil {
		return fmt.Errorf("failed restoring app data: %w", err)
	}

	files, err := os.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("failed restoring app data: %w", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), appId) {
			err = RestoreVolume(ctx, dockerClient, strings.TrimSuffix(file.Name(), ".tar.gz"), filepath.Join(backupPath, file.Name()))
			if err != nil {
				return fmt.Errorf("failed restoring app data: %w", err)
			}
		}
	}

	return nil
}
