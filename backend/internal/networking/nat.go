package networking

import (
	"context"
)

func TryMapPort(ctx context.Context, externalPort uint16, internalPort uint16) error {
	client, err := PickRouterClient(ctx)
	if err != nil {
		return err
	}

	privateIP, err := GetPrivateIP()
	if err != nil {
		return err
	}

	err = client.AddPortMapping("", externalPort, "TCP", internalPort, privateIP.String(), true, "Homecloud", 3600)
	if err != nil {
		return err
	}

	return nil
}
