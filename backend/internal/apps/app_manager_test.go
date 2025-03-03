package apps

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/An-Owlbear/homecloud/backend/internal/docker"
	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
	"github.com/An-Owlbear/homecloud/backend/internal/testutils"
	"github.com/An-Owlbear/homecloud/backend/internal/util"
	"github.com/google/go-cmp/cmp"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

func TestUpdateApps(t *testing.T) {
	// Setup dependencies
	dockerClient, err := testutils.CreateDindClient()
	defer testutils.CleanupDind()
	if err != nil {
		t.Fatalf("Unexpected error setting up docker: %s", err.Error())
	}

	dbPath := filepath.Join(util.RootDir(), "tmp/test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Unexpected error setting up DB: %s", err.Error())
	}
	queries := persistence.New(db)
	defer db.Close()
	defer os.Remove(dbPath)

	goose.SetBaseFS(os.DirFS(util.RootDir()))
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("Unexpected error setting up goose: %s", err.Error())
	}

	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("Unexpected error apply DB migrations: %s", err.Error())
	}

	storeClient := NewStoreClient("https://raw.githubusercontent.com/An-Owlbear/homecloud/07ea723942127e2b04e01de5b5e3d3e5158be27c/apps/list.json")

	app := persistence.AppPackage{
		Schema:      "v1.0",
		Version:     "v1.5",
		Id:          "traefik.whoami",
		Name:        "whoami",
		Author:      "traefik",
		Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		Containers: []persistence.PackageContainer{
			{
				Name:        "whoami",
				Image:       "traefik/whoami:v1.10.3",
				ProxyTarget: false,
				ProxyPort:   "801",
				Ports:       []string{"8000:80"},
				Environment: map[string]string{
					"test_env": "value",
				},
			},
		},
	}

	// Install and update app
	err = docker.InstallApp(dockerClient, app, config.Host{})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	// err = UntilState(dockerClient, app.Id, ContainerRunning, time.Second*10, time.Millisecond*10)
	// if err != nil {
	// 	t.Fatalf("Error waiting for container to start: %s", err.Error())
	// }

	schemaJson, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Unexpected error encoding app schema: %s", err.Error())
	}
	err = queries.CreateApp(context.Background(), persistence.CreateAppParams{
		ID:     app.Id,
		Schema: schemaJson,
	})
	if err != nil {
		t.Fatalf("Unexpected error saving app to DB: %s", err.Error())
	}

	err = UpdateApps(dockerClient, storeClient, queries, config.Host{})
	if err != nil {
		t.Fatalf("Unexpected error whilst updating apps: %s", err.Error())
	}

	// Check updated app has correct values
	app.Version = "v1.6"
	app.Containers[0].Ports = []string{"8001:80"}
	app.Containers[0].ProxyPort = "80"
	app.Containers[0].ProxyTarget = true
	app.Containers[0].Environment = nil
	testutils.HelpTestAppPackage(dockerClient, app, t)

	// Check DB is updated properly
	dbApp, err := queries.GetApp(context.Background(), app.Id)
	if err != nil {
		t.Fatalf("Unexpected error querying DB: %s", err.Error())
	}

	if diff := cmp.Diff(app, dbApp.Schema); diff != "" {
		t.Fatalf("Incorrect app information, difference: %s", diff)
	}
}
