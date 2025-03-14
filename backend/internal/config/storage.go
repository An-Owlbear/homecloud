package config

import (
	"os"
	"path/filepath"
)

type Storage struct {
	DataPath string
	AppDir   string
}

// GetAppDataMountPath retrieves the path that should be used when mounting the app data for use by docker. This is
// different from the path written to, since it's the path on the host
func (s Storage) GetAppDataMountPath(appId string) string {
	return filepath.Join(s.AppDir, s.DataPath, appId, "data")
}

// NewStorage create the configuration for the storage, if the user is in the host environment the path is found from
// the working directory, otherwise it is from the environment variable
func NewStorage(inHost bool) (*Storage, error) {
	var appDir string
	if inHost {
		var err error
		appDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	} else {
		appDir = os.Getenv("HOMECLOUD_APP_DIR")
	}

	return &Storage{
		DataPath: os.Getenv("DATA_PATH"),
		AppDir:   appDir,
	}, nil
}
