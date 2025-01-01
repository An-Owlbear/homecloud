package apps

import (
	"fmt"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Hosts map[string]*echo.Echo

// AddProxy creates and adds a new reverse proxy at the given address
func AddProxy(hosts Hosts, hostAddress string, proxyAddress string, proxyPort string) error {
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
	hosts[fmt.Sprintf("%s.home.cloud:1323", hostAddress)] = proxyHost

	return nil
}

// RemoveProxy removes the proxy for the given host address
func RemoveProxy(hosts Hosts, hostAddress string) {
	delete(hosts, fmt.Sprintf("%s.home.cloud:1323", hostAddress))
}
