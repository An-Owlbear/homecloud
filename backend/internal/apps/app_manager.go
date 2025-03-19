package apps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/client"
	"golang.org/x/mod/semver"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/storage"
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
		return fmt.Errorf("UpdateApps: failed to update packages list: %w", err)
	}

	apps, err := queries.GetAppsWithCreds(context.Background())
	if err != nil {
		return fmt.Errorf("UpdateApps: failed to get application details: %w", err)
	}

	// converts result to map
	appsMap := make(map[string]persistence.AppWithCreds)
	for _, app := range apps {
		appsMap[app.ID] = app
	}

	packages, err := queries.GetPackages(context.Background())
	if err != nil {
		return fmt.Errorf("UpdateApps: failed to get packages list: %w", err)
	}

	for _, listApp := range packages {
		// If the app is installed and the new version is greater update
		if app, ok := appsMap[listApp.ID]; ok && semver.Compare(listApp.Version, app.Schema.Version) == 1 {
			// Retrieve the full app package
			appPackage, err := storeClient.GetPackage(listApp.ID)
			if err != nil {
				return fmt.Errorf("UpdateApps: failed to get full package details: %w", err)
			}

			// Remove the app containers and reinstall in case of required changes
			err = docker.RemoveContainers(dockerClient, app.ID)
			if err != nil {
				return fmt.Errorf("UpdateApps: failed to remove containers: %w", err)
			}

			err = docker.InstallApp(dockerClient, appPackage, hostConfig, storageConfig)
			if err != nil {
				return fmt.Errorf("UpdateApps: failed to reinstall newer version fo app: %w", err)
			}

			schemaJson, err := json.Marshal(appPackage)
			if err != nil {
				return fmt.Errorf("UpdateApps: failed to marshal app package json: %w", err)
			}
			var templatedString bytes.Buffer
			err = storage.ApplyAppTemplate(
				string(schemaJson),
				&templatedString,
				appPackage,
				app.ClientID.String,
				app.ClientSecret.String,
				oryConfig,
				hostConfig,
				storageConfig,
			)
			if err != nil {
				return fmt.Errorf("UpdateApps: failed to apply app template to data template files: %w", err)
			}

			err = queries.UpdateApp(
				context.Background(), persistence.UpdateAppParams{
					ID:     app.ID,
					Schema: templatedString.String(),
				},
			)
			if err != nil {
				return fmt.Errorf("UpdateApps: failed to update app entry in DB: %w", err)
			}
		}
	}

	return nil
}

func StartApp(
	dockerClient *client.Client,
	queries *persistence.Queries,
	hosts *Hosts,
	appDataHandler *storage.AppDataHandler,
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
	appDataHandler *storage.AppDataHandler,
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

func BackupApp(
	ctx context.Context,
	dockerClient *client.Client,
	storageConfig config.Storage,
	appId string,
	targetDevice string,
) error {
	details, err := storage.GetExternalPartition(targetDevice)
	if err != nil {
		return fmt.Errorf("error checking drive is external: %w", err)
	}

	mountPath, err := storage.MountPartition(details)
	if err != nil {
		return fmt.Errorf("error mounting partition: %w", err)
	}
	defer storage.UnmountPartition(details)

	outputPath := filepath.Join(mountPath, "backup", appId, time.Now().Format("20060102150405"))
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("error creating backup directory: %w", err)
	}

	err = docker.BackupAppData(ctx, dockerClient, storageConfig, appId, outputPath)
	if err != nil {
		return fmt.Errorf("error backup app data: %w", err)
	}

	return nil
}

func RestoreApp(
	ctx context.Context,
	dockerClient *client.Client,
	queries *persistence.Queries,
	hosts *Hosts,
	appDataHandler *storage.AppDataHandler,
	hostConfig config.Host,
	storageConfig config.Storage,
	oryConfig config.Ory,
	appId string,
	targetDevice string,
	targetBackup string,
) error {
	// Checks and mounts drive
	details, err := storage.GetExternalPartition(targetDevice)
	if err != nil {
		return fmt.Errorf("error checking drive is external: %w", err)
	}

	mountPath, err := storage.MountPartition(details)
	if err != nil {
		return fmt.Errorf("error mounting partition: %w", err)
	}
	defer storage.UnmountPartition(details)

	// Checks the specified backup exists on the drive
	backupPath := filepath.Join(mountPath, "backup", appId, targetBackup)
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("specified backup %s does not exist on drive %s: %w", targetBackup, targetDevice, err)
	}

	// Stops app
	if err := StopApp(dockerClient, queries, appId); err != nil {
		return fmt.Errorf("error stopping app: %w", err)
	}

	// Removes the app containers and volumes
	if err := docker.UninstallApp(dockerClient, appId); err != nil {
		return fmt.Errorf("error removing containers for %s: %w", appId, err)
	}
	if err := docker.RemoveAppVolumes(ctx, dockerClient, appId); err != nil {
		return fmt.Errorf("error removing volumes for %s: %w", appId, err)
	}

	// Clears app data folder
	if err := os.RemoveAll(filepath.Join(storageConfig.DataPath, appId, "data")); err != nil {
		return fmt.Errorf("error removing data directory: %w", err)
	}

	if err := docker.RestoreAppData(ctx, dockerClient, storageConfig, appId, backupPath); err != nil {
		return fmt.Errorf("error restoring app data: %w", err)
	}

	// Recreates app containers
	app, err := queries.GetApp(ctx, appId)
	if err != nil {
		return fmt.Errorf("error retrieving app information: %w", err)
	}

	if err := docker.InstallApp(dockerClient, app.Schema, hostConfig, storageConfig); err != nil {
		return fmt.Errorf("error recreating app containers: %w", err)
	}

	if err := StartApp(dockerClient, queries, hosts, appDataHandler, hostConfig, oryConfig, appId); err != nil {
		return fmt.Errorf("error starting app: %w", err)
	}

	return nil
}
