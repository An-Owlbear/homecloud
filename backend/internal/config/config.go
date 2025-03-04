package config

type Config struct {
	Host    Host
	Ory     Ory
	Store   Store
	Storage Storage
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

	store := NewStore()

	storage := NewStorage()

	return &Config{
		Host:    *host,
		Ory:     *ory,
		Store:   *store,
		Storage: *storage,
	}, nil
}
