package apps

import (
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"reflect"
	"testing"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
)

const storeUrl = "https://raw.githubusercontent.com/An-Owlbear/homecloud/07ea723942127e2b04e01de5b5e3d3e5158be27c/apps"

// Tests whether the retrieved list of packages is correct
func TestUpdatePackageList(t *testing.T) {
	client := prepareClient()
	err := client.UpdatePackageList()

	if err != nil {
		t.Fatalf("Unexpected error occured: %s", err.Error())
	}

	expectedResults := []PackageListItem{
		{
			Id:          "immich-app.immich",
			Name:        "immich",
			Version:     "v1.124.3",
			Author:      "immich-app",
			Description: "High-performance self-hosted photo and video management solution",
			ImageUrl:    storeUrl + "/packages/immich-app.immich/icon.png",
		},
		{
			Id:          "paperless-ngx.paperless-ngx",
			Name:        "paperless-ngx",
			Version:     "v1.0",
			Author:      "paperless-ngx",
			Description: " A community-supported supercharged version of paperless: scan, index and archive all your physical documents",
			ImageUrl:    storeUrl + "/packages/paperless-ngx.paperless-ngx/icon.png",
		},
		{
			Id:          "traefik.whoami",
			Name:        "whoami",
			Version:     "v1.6",
			Author:      "traefik",
			Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
			ImageUrl:    storeUrl + "/packages/traefik.whoami/icon.png",
		},
	}

	if !reflect.DeepEqual(expectedResults, client.Packages) {
		t.Fatalf("Packages are not equal\nExpected packages: %+v\nActual data: %+v", expectedResults, client.Packages)
	}
}

// Tests whether the retrieved individual package is correct
func TestGetPackage(t *testing.T) {
	client := prepareClient()

	app, err := client.GetPackage("traefik.whoami")
	if err != nil {
		t.Fatalf("Unexpected error occured : %s", err.Error())
	}

	expectedPackage := persistence.AppPackage{
		Schema:      "v1.0",
		Version:     "v1.6",
		Id:          "traefik.whoami",
		Name:        "whoami",
		Author:      "traefik",
		Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		Containers: []persistence.PackageContainer{
			{
				Name:        "whoami",
				Image:       "traefik/whoami:v1.10.3",
				ProxyTarget: true,
				ProxyPort:   "80",
				Ports:       []string{"8001:80"},
			},
		},
	}

	if !reflect.DeepEqual(expectedPackage, app) {
		t.Fatalf("Package not expected value\nExpected value: %+v\nActual value: %+v", expectedPackage, app)
	}
}

// Tests retrieving an unknown package is correctly handled
func TestGetUnknownPackage(t *testing.T) {
	client := prepareClient()
	_, err := client.GetPackage("fake.package")
	expectedErr := "invalid HTTP response 404"

	if err.Error() != expectedErr {
		t.Fatalf("Unexpected error response\nExpected: %s\nActual: %s", expectedErr, err.Error())
	}
}

// Prepares the client
// update to use local web server
func prepareClient() *StoreClient {
	// URL to repository at a specific commit so new packages don't affect the tests
	return NewStoreClient(config.Store{StoreUrl: storeUrl + "/list.json"})
}
