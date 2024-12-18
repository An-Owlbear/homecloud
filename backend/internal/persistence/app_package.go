package persistence

type AppPackage struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Containers  []struct {
		Name        string            `json:"name"`
		Image       string            `json:"image"`
		Env         map[string]string `json:"env"`
		Ports       []string          `json:"ports"`
		ProxyTarget bool              `json:"proxy_target"`
		ProxyPort   string            `json:"proxy_port"`
	} `json:"containers"`
}
