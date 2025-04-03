package docker

import (
	"context"
	"fmt"
	"time"

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

func UntilHealthy(ctx context.Context, dockerClient *client.Client, containerId string) error {
	err := UntilState(dockerClient, containerId, ContainerRunning, time.Second*20, time.Millisecond*10)
	if err != nil {
		return err
	}
	return nil
}
