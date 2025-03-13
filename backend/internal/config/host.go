package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Host struct {
	Host        string
	Port        int
	HTTPS       bool
	PortForward bool
	Url         url.URL
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
	scheme := "https"
	if !https {
		scheme = "http"
	}

	hostUrl := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", host, port),
	}

	return &Host{
		Host:        host,
		Port:        port,
		HTTPS:       https,
		PortForward: portForward,
		Url:         hostUrl,
	}, nil
}
