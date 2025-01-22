package apps

import (
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/url"
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
	hosts.hosts[fmt.Sprintf("%s.%s:%d", hostAddress, hosts.config.Host, hosts.config.Port)] = proxyHost

	return nil
}

// RemoveProxy removes the proxy for the given host address
func (hosts *Hosts) RemoveProxy(hostAddress string) {
	delete(hosts.hosts, fmt.Sprintf("%s.%s:%d", hostAddress, hosts.config.Host, hosts.config.Port))
}
