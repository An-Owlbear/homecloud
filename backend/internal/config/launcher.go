package config

import (
	"os"
	"strconv"
)

type Launcher struct {
	Url          string
	AlwaysUpdate bool
}

func NewLauncher() (*Launcher, error) {
	alwaysUpdate, err := strconv.ParseBool(Getenv("HOMECLOUD_LAUNCHER_ALWAYS_UPDATE", "false"))
	if err != nil {
		return nil, err
	}

	return &Launcher{
		Url:          os.Getenv("HOMECLOUD_LAUNCHER_URL"),
		AlwaysUpdate: alwaysUpdate,
	}, nil
}
