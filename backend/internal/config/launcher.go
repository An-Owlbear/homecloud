package config

import (
	"os"
	"strconv"
)

type LauncherEnv struct {
	Url            string
	AlwaysUpdate   bool
	ConfigFilename string
}

func NewLauncher() (*LauncherEnv, error) {
	alwaysUpdate, err := strconv.ParseBool(Getenv("HOMECLOUD_LAUNCHER_ALWAYS_UPDATE", "false"))
	if err != nil {
		return nil, err
	}

	return &LauncherEnv{
		Url:            os.Getenv("HOMECLOUD_LAUNCHER_URL"),
		AlwaysUpdate:   alwaysUpdate,
		ConfigFilename: Getenv("HOMECLOUD_LAUNCHER_CONFIG_FILENAME", "config.json"),
	}, nil
}
