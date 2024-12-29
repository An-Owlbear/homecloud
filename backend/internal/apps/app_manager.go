package apps

import (
	"context"
	"encoding/json"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"golang.org/x/mod/semver"
)

type AppManager struct {
	dockerClient *client.Client
	storeClient  *StoreClient
	queries      *persistence.Queries
}

func NewAppManager(dockerClient *client.Client, storeClient *StoreClient, queries *persistence.Queries) *AppManager {
	return &AppManager{
		dockerClient: dockerClient,
		storeClient:  storeClient,
		queries:      queries,
	}
}

// UpdateApps updates the list of available apps and updates any outdated apps
func (am *AppManager) UpdateApps() error {
	err := am.storeClient.UpdatePackageList()
	if err != nil {
		return err
	}

	apps, err := am.queries.GetApps(context.Background())
	if err != nil {
		return err
	}

	// converts result to map
	appsMap := make(map[string]persistence.GetAppsRow)
	for _, app := range apps {
		appsMap[app.ID] = app
	}

	for _, listApp := range am.storeClient.Packages {
		// If the app is installed and the new version is greater update
		if app, ok := appsMap[listApp.Id]; ok && semver.Compare(listApp.Version, app.Schema.Version) == 1 {
			// Retrieve the full app package
			appPackage, err := am.storeClient.GetPackage(listApp.Id)
			if err != nil {
				return err
			}

			// Remove the app containers and reinstall in case of required changes
			err = RemoveContainers(am.dockerClient, app.ID)
			if err != nil {
				return err
			}

			err = InstallApp(am.dockerClient, appPackage)
			if err != nil {
				return err
			}

			schemaJson, err := json.Marshal(appPackage)
			if err != nil {
				return err
			}

			err = am.queries.UpdateApp(context.Background(), persistence.UpdateAppParams{
				ID:     app.ID,
				Schema: schemaJson,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
