package config

import (
	"fmt"
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

func (h *Host) PublicUrl() string {
	scheme := "http"
	if h.HTTPS {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s", scheme, h.Host)
	if h.Port != 80 && h.Port != 443 {
		url = fmt.Sprintf("%s://%s:%d", scheme, h.Host, h.Port)
	}
	return url
}

func (h *Host) PublicSubdomain(app string) string {
	scheme := "http"
	if h.HTTPS {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s.%s", scheme, app, h.Host)
	if h.Port != 80 && h.Port != 443 {
		url = fmt.Sprintf("%s:%d", url, h.Port)
	}
	return url
}
