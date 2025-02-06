package config

type Config struct {
	Host Host
	Ory  Ory
}

func LoadConfig() (*Config, error) {
	host, err := NewHost()
	if err != nil {
		return nil, err
	}

	ory, err := OryFromEnv()
	if err != nil {
		return nil, err
	}

	return &Config{
		Host: *host,
		Ory:  *ory,
	}, nil
}
