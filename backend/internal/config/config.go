package config

type Config struct {
	Host     Host
	Ory      Ory
	Store    Store
	Storage  Storage
	Launcher LauncherEnv
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

	storage, err := NewStorage(false)
	if err != nil {
		return nil, err
	}

	launcher, err := NewLauncher()
	if err != nil {
		return nil, err
	}

	return &Config{
		Host:     *host,
		Ory:      *ory,
		Store:    *store,
		Storage:  *storage,
		Launcher: *launcher,
	}, nil
}
