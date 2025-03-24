package config

import (
	"os"
	"strconv"
)

type Host struct {
	Host        string
	Port        int
	HTTPS       bool
	PortForward bool
}

func NewHost() (*Host, error) {
	port, err := strconv.Atoi(os.Getenv("HOMECLOUD_PORT"))
	if err != nil {
		return nil, err
	}

	https, err := strconv.ParseBool(Getenv("USE_HTTPS", "true"))
	if err != nil {
		return nil, err
	}

	portForward, err := strconv.ParseBool(Getenv("PORT_FORWARD", "true"))
	if err != nil {
		return nil, err
	}

	host := os.Getenv("HOMECLOUD_HOST")

	return &Host{
		Host:        host,
		Port:        port,
		HTTPS:       https,
		PortForward: portForward,
	}, nil
}
