package persistence

type AppPackage struct {
	Schema      string             `json:"schema"`
	Version     string             `json:"version"`
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Author      string             `json:"author"`
	Description string             `json:"description"`
	Categories  []string           `json:"categories"`
	OidcEnabled bool               `json:"oidc_enabled"`
	OidcScopes  []string           `json:"oidc_scopes"`
	Containers  []PackageContainer `json:"containers"`
}

type PackageContainer struct {
	Name             string            `json:"name"`
	Image            string            `json:"image"`
	Command          string            `json:"command"`
	Restart          string            `json:"restart"`
	Environment      map[string]string `json:"environment"`
	Ports            []string          `json:"ports"`
	Volumes          []string          `json:"volumes"`
	ExtraHosts       []string          `json:"extra_hosts"`
	Privileged       bool              `json:"privileged"`
	ProxyTarget      bool              `json:"proxy_target"`
	ProxyPort        string            `json:"proxy_port"`
	OidcRedirectUris []string          `json:"oidc_redirect_uris"`
}
