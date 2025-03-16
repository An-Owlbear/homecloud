package apps

import (
	"fmt"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

type HostsMap map[string]*echo.Echo

type Hosts struct {
	hosts  HostsMap
	config config.Host
}

func NewHosts(hosts HostsMap, config config.Host) *Hosts {
	return &Hosts{hosts: hosts, config: config}
}

// AddProxy creates and adds a new reverse proxy at the given address
func (hosts *Hosts) AddProxy(hostAddress string, proxyAddress string, proxyPort string) error {
	// creates echo host for the application
	proxyHost := echo.New()

	appUrl, err := url.Parse(fmt.Sprintf("http://%s:%s", proxyAddress, proxyPort))
	if err != nil {
		return err
	}

	targets := []*middleware.ProxyTarget{
		{
			URL: appUrl,
		},
	}

	proxyHost.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(targets)))

	publicHost := fmt.Sprintf("%s.%s", hostAddress, hosts.config.Host)
	if hosts.config.Port != 80 && hosts.config.Port != 443 {
		publicHost = fmt.Sprintf("%s:%d", publicHost, hosts.config.Port)
	}
	hosts.hosts[fmt.Sprintf("%s.%s:%d", hostAddress, hosts.config.Host, hosts.config.Port)] = proxyHost

	return nil
}

// RemoveProxy removes the proxy for the given host address
func (hosts *Hosts) RemoveProxy(hostAddress string) {
	delete(hosts.hosts, fmt.Sprintf("%s.%s:%d", hostAddress, hosts.config.Host, hosts.config.Port))
}
