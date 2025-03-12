package testutils

import "github.com/An-Owlbear/homecloud/backend/internal/config"

func SetupTempStorage() config.Storage {
	// TODO: clear storage directory
	return config.Storage{
		DataPath: "/tmp/homecloud_test/storage",
	}
}
