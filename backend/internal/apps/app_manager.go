package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"golang.org/x/mod/semver"
)

// UpdateApps updates the list of available apps and updates any outdated apps
func UpdateApps(dockerClient *client.Client, storeClient *StoreClient, queries *persistence.Queries) error {
	err := storeClient.UpdatePackageList()
	if err != nil {
		return err
	}

	apps, err := queries.GetApps(context.Background())
	if err != nil {
		return err
	}

	// converts result to map
	appsMap := make(map[string]persistence.GetAppsRow)
	for _, app := range apps {
		appsMap[app.ID] = app
	}

	for _, listApp := range storeClient.Packages {
		// If the app is installed and the new version is greater update
		if app, ok := appsMap[listApp.Id]; ok && semver.Compare(listApp.Version, app.Schema.Version) == 1 {
			// Retrieve the full app package
			appPackage, err := storeClient.GetPackage(listApp.Id)
			if err != nil {
				return err
			}

			// Remove the app containers and reinstall in case of required changes
			err = docker.RemoveContainers(dockerClient, app.ID)
			if err != nil {
				return err
			}

			err = docker.InstallApp(dockerClient, appPackage)
			if err != nil {
				return err
			}

			schemaJson, err := json.Marshal(appPackage)
			if err != nil {
				return err
			}

			err = queries.UpdateApp(context.Background(), persistence.UpdateAppParams{
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

func StartApp(dockerClient *client.Client, queries *persistence.Queries, hosts Hosts, appId string) error {
	app, err := queries.GetApp(context.Background(), appId)
	if err != nil {
		return err
	}

	// Start app containers
	if err := docker.StartApp(dockerClient, app.ID); err != nil {
		return err
	}

	// Retrieve containers and wait for them to finish starting
	containers, err := docker.GetAppContainers(dockerClient, app.ID)
	if err != nil {
		return err
	}

	for _, appContainer := range containers {
		err = docker.UntilState(dockerClient, appContainer.ID, docker.ContainerRunning, time.Second*20, time.Millisecond*10)
		if err != nil {
			return err
		}
	}

	// Check if any containers need proxying and proxy if needed
	for _, packageContainer := range app.Schema.Containers {
		if packageContainer.ProxyTarget {
			err = AddProxy(hosts, app.Schema.Name, fmt.Sprintf("%s-%s", app.Schema.Id, packageContainer.Name), packageContainer.ProxyPort)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
