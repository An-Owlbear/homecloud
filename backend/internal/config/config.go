package config

type Host struct {
	Host string
	Port int
}

func NewHost(host string, port int) Host {
	return Host{
		Host: host,
		Port: port,
	}
}
