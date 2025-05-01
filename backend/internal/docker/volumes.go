package docker

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// RemoveAppVolumes removes all docker volumes associated with the given application
func RemoveAppVolumes(ctx context.Context, dockerClient *client.Client, appId string) error {
	volumes, err := dockerClient.VolumeList(ctx, volume.ListOptions{Filters: filters.NewArgs(AppFilter(appId))})
	if err != nil {
		return err
	}

	for _, appVolume := range volumes.Volumes {
		if err := dockerClient.VolumeRemove(ctx, appVolume.Name, false); err != nil {
			return err
		}
	}

	return nil
}
