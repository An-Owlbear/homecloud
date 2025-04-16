package apps

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

type HostsMap map[string]*echo.Echo

type Hosts struct {
	hosts        HostsMap
	tlsManager   *autocert.Manager
	config       config.Host
	checkedCerts bool
}

func NewHosts(hosts HostsMap, tlsManager *autocert.Manager, config config.Host) *Hosts {
	return &Hosts{
		hosts:        hosts,
		tlsManager:   tlsManager,
		config:       config,
		checkedCerts: false,
	}
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
	hosts.hosts[publicHost] = proxyHost

	// If HTTPS is enabled preload certificate, skips when nil, like during startup
	if hosts.config.HTTPS && hosts.tlsManager != nil {
		if _, err := hosts.tlsManager.GetCertificate(&tls.ClientHelloInfo{ServerName: publicHost}); err != nil {
			return err
		}
	}

	return nil
}

// RemoveProxy removes the proxy for the given host address
func (hosts *Hosts) RemoveProxy(hostAddress string) {
	delete(hosts.hosts, fmt.Sprintf("%s.%s:%d", hostAddress, hosts.config.Host, hosts.config.Port))
}

// SetAutoTLSManager sets the AutoTLSManager used for retrieving certificates. This is done to skip the process
// during startup
func (hosts *Hosts) SetAutoTLSManager(tlsManager *autocert.Manager) {
	hosts.tlsManager = tlsManager
}

// EnsureCertificates ensures TLS certificates have been retrieved for all domains
func (hosts *Hosts) EnsureCertificates() error {
	totalCerts := len(hosts.hosts)
	i := 0
	for host, _ := range hosts.hosts {
		slog.Info(fmt.Sprintf("Ensuring certificate for %s, %d remaining", host, totalCerts-i))
		if _, err := hosts.tlsManager.GetCertificate(&tls.ClientHelloInfo{ServerName: host}); err != nil {
			return err
		}
		i++
	}
	hosts.checkedCerts = true
	return nil
}

func (hosts *Hosts) CertsReady() bool {
	return hosts.checkedCerts
}
