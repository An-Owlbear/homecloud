package config

import (
	"os"
	"strconv"
)

type Host struct {
	Host string
	Port int
}

func NewHost() (*Host, error) {
	port, err := strconv.Atoi(os.Getenv("HOMECLOUD_PORT"))
	if err != nil {
		return nil, err
	}

	return &Host{
		Host: os.Getenv("HOMECLOUD_HOST"),
		Port: port,
	}, nil
}
