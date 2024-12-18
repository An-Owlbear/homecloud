package apps

import (
	"reflect"
	"testing"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
)

// Tests whether the retrieved list of packages is correct
func TestUpdatePackageList(t *testing.T) {
	client := prepareClient()
	err := client.UpdatePackageList()

	if err != nil {
		t.Fatalf("Unexpected error occured: %s", err.Error())
	}

	expectedResults := []PackageListItem{
		{
			Id:          "traefik.whoami",
			Name:        "whoami",
			Version:     "v1.5",
			Author:      "traefik",
			Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		},
	}

	if !reflect.DeepEqual(expectedResults, client.packages) {
		t.Fatalf("Packages are not equal\nExpected packages: %+v\nActual data: %+v", expectedResults, client.packages)
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
		Schema: "v1.0",
		Version: "v1.5",
		Id: "traefik.whoami",
		Name: "whoami",
		Author: "traefik",
		Description: "Tiny Go webserver that prints OS information and HTTP request to output.",
		Containers: []persistence.PackageContainer {
			{
				Name: "whoami",
				Image: "traefik/whoami:v1.10.3",
				ProxyTarget: true,
				ProxyPort: "80",
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
	expectedErr := "Invalid HTTP response 404"

	if err.Error() != expectedErr {
		t.Fatalf("Unexpected error response\nExpected: %s\nActual: %s", expectedErr, err.Error())
	}
}

// Prepares the client
// update to use local web server
func prepareClient() *StoreClient {
	// URL to repository at a specific commit so new packages don't affect the tests
	return NewStoreClient("https://raw.githubusercontent.com/An-Owlbear/homecloud/e1ddd9a14e400a53b943cd0934bbd60c897968f0/apps/list.json")
}
