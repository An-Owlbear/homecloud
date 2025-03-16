package networking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

var PortForwardError = errors.New("couldn't setup port forwarding properly")

func TryMapPort(ctx context.Context, externalPort uint16, internalPort uint16, deviceConfig config.DeviceConfig) error {
	client, err := PickRouterClient(ctx)
	if err != nil {
		return err
	}

	privateIP, err := GetPrivateIP()
	if err != nil {
		return err
	}

	_ = client.DeletePortMapping("", externalPort, "TCP")

	err = client.AddPortMapping("", externalPort, "TCP", internalPort, privateIP.String(), true, "Homecloud", 3600)
	if err != nil {
		return err
	}

	err = CheckPortForwarding(deviceConfig, int(externalPort))
	if err != nil {
		return err
	}

	return nil
}

type CheckPortForwardingRequest struct {
	DeviceId  string `json:"device_id"`
	DeviceKey string `json:"device_key"`
	Port      int    `json:"port"`
}

func CheckPortForwarding(deviceConfig config.DeviceConfig, forwardedPort int) error {
	requestBody := CheckPortForwardingRequest{
		DeviceId:  deviceConfig.DeviceId,
		DeviceKey: deviceConfig.DeviceKey,
		Port:      forwardedPort,
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"https://ysiu2v4par5wpvkzu6jgtm3pty0yyziy.lambda-url.eu-west-2.on.aws/",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(response.Body)
		fmt.Println(string(body))
		return PortForwardError
	}

	return nil
}
