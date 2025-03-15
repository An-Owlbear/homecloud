package apps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"time"

	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/docker/docker/client"
	"golang.org/x/mod/semver"
)

// UpdateApps updates the list of available apps and updates any outdated apps
func UpdateApps(
	dockerClient *client.Client,
	storeClient *StoreClient,
	queries *persistence.Queries,
	oryConfig config.Ory,
	hostConfig config.Host,
	storageConfig config.Storage,
) error {
	err := storeClient.UpdatePackageList(context.Background(), queries)
	if err != nil {
		return err
	}

	apps, err := queries.GetAppsWithCreds(context.Background())
	if err != nil {
		return err
	}

	// converts result to map
	appsMap := make(map[string]persistence.AppWithCreds)
	for _, app := range apps {
		appsMap[app.ID] = app
	}

	packages, err := queries.GetPackages(context.Background())
	if err != nil {
		return err
	}

	for _, listApp := range packages {
		// If the app is installed and the new version is greater update
		if app, ok := appsMap[listApp.ID]; ok && semver.Compare(listApp.Version, app.Schema.Version) == 1 {
			// Retrieve the full app package
			appPackage, err := storeClient.GetPackage(listApp.ID)
			if err != nil {
				return err
			}

			// Remove the app containers and reinstall in case of required changes
			err = docker.RemoveContainers(dockerClient, app.ID)
			if err != nil {
				return err
			}

			err = docker.InstallApp(dockerClient, appPackage, hostConfig, storageConfig)
			if err != nil {
				return err
			}

			schemaJson, err := json.Marshal(appPackage)
			if err != nil {
				return err
			}
			var templatedString bytes.Buffer
			err = persistence.ApplyAppTemplate(
				string(schemaJson),
				&templatedString,
				app.ClientID.String,
				app.ClientSecret.String,
				oryConfig,
				hostConfig,
				storageConfig,
			)
			if err != nil {
				return err
			}

			err = queries.UpdateApp(
				context.Background(), persistence.UpdateAppParams{
					ID:     app.ID,
					Schema: templatedString.String(),
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func StartApp(
	dockerClient *client.Client,
	queries *persistence.Queries,
	hosts *Hosts,
	appDataHandler *persistence.AppDataHandler,
	hostConfig config.Host,
	oryConfig config.Ory,
	appId string,
) error {
	app, err := queries.GetApp(context.Background(), appId)
	if err != nil {
		return err
	}

	// Renders the templates in the config files
	err = appDataHandler.RenderTemplates(context.Background(), queries, oryConfig, hostConfig, appId)
	if err != nil {
		return err
	}

	// Start app containers
	if err := docker.StartApp(dockerClient, appId); err != nil {
		return err
	}

	// Retrieve containers and wait for them to finish starting
	containers, err := docker.GetAppContainers(dockerClient, appId)
	if err != nil {
		return err
	}

	for _, appContainer := range containers {
		err = docker.UntilState(
			dockerClient,
			appContainer.ID,
			docker.ContainerRunning,
			time.Second*20,
			time.Millisecond*10,
		)
		if err != nil {
			return err
		}
	}

	// Check if any containers need proxying and proxy if needed
	for _, packageContainer := range app.Schema.Containers {
		if packageContainer.ProxyTarget {
			err = hosts.AddProxy(
				app.Schema.Name,
				fmt.Sprintf("%s-%s", app.Schema.Id, packageContainer.Name),
				packageContainer.ProxyPort,
			)
			if err != nil {
				return err
			}
		}
	}

	// Sets the status in the database
	err = queries.SetStatus(
		context.Background(), persistence.SetStatusParams{
			ID:     appId,
			Status: string(docker.ContainerRunning),
		},
	)

	return nil
}

// StopApp - stops the specified app and sets the status in the database
func StopApp(dockerClient *client.Client, queries *persistence.Queries, appId string) error {
	_, err := queries.GetApp(context.Background(), appId)
	if err != nil {
		return err
	}

	// Stops the containers of the app
	err = docker.StopApp(dockerClient, appId)
	if err != nil {
		return err
	}

	// Sets the status in the database
	err = queries.SetStatus(
		context.Background(), persistence.SetStatusParams{
			ID:     appId,
			Status: string(docker.ContainerExited),
		},
	)
	if err != nil {
		return err
	}

	// TODO: instead of proxying app proxy static page instead
	return nil
}

func SetupProxies(
	dockerClient *client.Client,
	queries *persistence.Queries,
	hosts *Hosts,
	appDataHandler *persistence.AppDataHandler,
	hostConfig config.Host,
	oryConfig config.Ory,
) error {
	apps, err := queries.GetApps(context.Background())
	if err != nil {
		return err
	}

	// Ensures apps are properly started with proxies
	for _, app := range apps {
		if app.Status == string(docker.ContainerRunning) {
			err = StartApp(dockerClient, queries, hosts, appDataHandler, hostConfig, oryConfig, app.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
