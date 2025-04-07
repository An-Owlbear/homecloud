package config

import "os"

type Docker struct {
	ContainerName string
}

func NewDocker() *Docker {
	return &Docker{
		ContainerName: os.Getenv("HOMECLOUD_CONTAINER_NAME"),
	}
}
