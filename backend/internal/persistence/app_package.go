package persistence

type AppPackage struct {
	Schema      string             `json:"schema"`
	Version     string             `json:"version"`
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Author      string             `json:"author"`
	Description string             `json:"description"`
	Containers  []PackageContainer `json:"containers"`
}

type PackageContainer struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Environment map[string]string `json:"environment"`
	Ports       []string          `json:"ports"`
	Volumes     []string          `json:volumes`
	ProxyTarget bool              `json:"proxy_target"`
	ProxyPort   string            `json:"proxy_port"`
}
