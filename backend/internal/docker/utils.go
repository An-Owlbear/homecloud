package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/util"
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
	err := util.WaitUntil(func() (bool, error) {
		info, err := dockerClient.ContainerInspect(context.Background(), containerId)
		if err != nil {
			return false, err
		}
		if info.State.Running {
			// For containers that don't report health data we can only assume they're ready
			if info.State.Health == nil {
				return true, nil
			}

			// If the container has health values to indicate not being ready wait, otherwise continue
			if info.State.Health.Status == "starting" || info.State.Health.Status == "unhealthy" {
				return false, nil
			} else {
				return true, nil
			}
		} else {
			return false, nil
		}
	}, time.Minute, time.Millisecond*25)
	if err != nil {
		return err
	}

	return nil
}
