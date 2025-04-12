package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

func GetOrCreateNetwork(
	ctx context.Context,
	dockerClient *client.Client,
	networkName string,
	labels map[string]string,
) (string, error) {
	var networkId string
	networkInspect, err := dockerClient.NetworkInspect(ctx, networkName, network.InspectOptions{})
	if err != nil {
		networkVar, err := dockerClient.NetworkCreate(ctx, networkName, network.CreateOptions{
			Labels: labels,
		})
		if err != nil {
			return "", err
		}
		networkId = networkVar.ID
	} else {
		networkId = networkInspect.ID
	}

	return networkId, nil
}

func ConnectProxyNetworks(
	ctx context.Context,
	dockerClient *client.Client,
	dockerConfig config.Docker,
) error {
	networks, err := dockerClient.NetworkList(ctx, network.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "label",
				Value: APP_ID_LABEL,
			}, filters.KeyValuePair{
				Key:   "name",
				Value: "-proxy",
			},
		),
	})
	if err != nil {
		return fmt.Errorf("error listing proxy networks: %w", err)
	}

	for _, appNetwork := range networks {
		err = dockerClient.NetworkConnect(ctx, appNetwork.ID, dockerConfig.ContainerName, &network.EndpointSettings{})
		if err != nil && !IsNetworkAlreadyConnectErr(err) {
			return fmt.Errorf("error connecting proxy network %s: %w", appNetwork.Name, err)
		}
	}

	return nil
}

func IsNetworkAlreadyConnectErr(err error) bool {
	return err != nil && errdefs.IsForbidden(err) && strings.Contains(err.Error(), "already exists")
}
