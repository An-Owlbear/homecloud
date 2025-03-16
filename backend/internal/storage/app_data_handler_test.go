package storage

import (
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"os"
	"path"
	"testing"
)

func TestSavePackage(t *testing.T) {
	dataHandler := NewAppDataHandler(
		config.Storage{DataPath: "/tmp/homecloud_testing"},
		config.Store{StoreUrl: "https://raw.githubusercontent.com/An-Owlbear/homecloud/07ea723942127e2b04e01de5b5e3d3e5158be27c/apps/list.json"},
	)

	appId := "immich-app.immich"
	err := dataHandler.SavePackage(appId)
	if err != nil {
		t.Fatalf("Failed saving package: %s", err.Error())
	}

	expectedFiles := []string{"schema.json", "icon.png"}
	for _, file := range expectedFiles {
		if _, err := os.Stat(path.Join(dataHandler.storageConfig.DataPath, appId, file)); os.IsNotExist(err) {
			t.Fatalf("Expected file %s to exist", file)
		}
	}
}
