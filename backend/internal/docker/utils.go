package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

func UntilRemoved(ctx context.Context, dockerClient *client.Client, containerId string) error {
	statusCh, errCh := dockerClient.ContainerWait(ctx, containerId, container.WaitConditionRemoved)
	select {
	case err := <-errCh:
		if err != nil && !errdefs.IsNotFound(err) {
			return fmt.Errorf("error waiting for container :%s: %v", containerId, err)
		}
	case <-statusCh:
		return nil
	}

	return nil
}
