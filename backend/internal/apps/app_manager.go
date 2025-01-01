package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"golang.org/x/mod/semver"
)

type AppManager struct {
	dockerClient *client.Client
	storeClient  *StoreClient
	queries      *persistence.Queries
	hosts        Hosts
}

func NewAppManager(
	dockerClient *client.Client,
	storeClient *StoreClient,
	queries *persistence.Queries,
	hosts Hosts,
) *AppManager {
	return &AppManager{
		dockerClient: dockerClient,
		storeClient:  storeClient,
		queries:      queries,
		hosts:        hosts,
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

func (am *AppManager) StartApp(appId string) error {
	app, err := am.queries.GetApp(context.Background(), appId)
	if err != nil {
		return err
	}

	// Start app containers
	if err := StartApp(am.dockerClient, app.ID); err != nil {
		return err
	}

	// Retrieve containers and wait for them to finish starting
	containers, err := GetAppContainers(am.dockerClient, app.ID)
	if err != nil {
		return err
	}

	for _, appContainer := range containers {
		err = UntilState(am.dockerClient, appContainer.ID, ContainerRunning, time.Second*20, time.Millisecond*10)
		if err != nil {
			return err
		}
	}

	// Check if any containers need proxying and proxy if needed
	for _, packageContainer := range app.Schema.Containers {
		if packageContainer.ProxyTarget {
			err = AddProxy(am.hosts, app.Schema.Name, fmt.Sprintf("%s-%s", app.Schema.Id, packageContainer.Name), packageContainer.ProxyPort)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
