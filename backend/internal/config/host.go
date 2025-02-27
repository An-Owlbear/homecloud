package config

import (
	"os"
	"strconv"
)

type Host struct {
	Host  string
	Port  int
	HTTPS bool
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

	return &Host{
		Host:  os.Getenv("HOMECLOUD_HOST"),
		Port:  port,
		HTTPS: https,
	}, nil
}
