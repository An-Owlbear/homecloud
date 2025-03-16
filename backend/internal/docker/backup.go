package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"path/filepath"
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
				Source:   filepath.Join(outputDir),
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
