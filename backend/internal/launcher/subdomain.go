package launcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var SubdomainError = errors.New("error assigning subdomain")

type SubdomainRequest struct {
	DeviceId  string `json:"device_id"`
	DeviceKey string `json:"device_key"`
	Subdomain string `json:"subdomain"`
	IPAddress string `json:"ip_address"`
}

func SetSubdomain(ctx context.Context, subdomainRequest SubdomainRequest) error {
	// Converts request body to JSON string
	body, err := json.Marshal(subdomainRequest)
	if err != nil {
		return err
	}

	// Creates request object
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://fcqvpr26frccoy4bwu4x2cp2gq0buqyv.lambda-url.eu-west-2.on.aws/",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	// Creates http client and sends request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	fmt.Println(string(responseBody))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return SubdomainError
	}

	return nil
}
