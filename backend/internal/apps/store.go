package apps

import (
	"encoding/json"
	"fmt"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
)

type PackageListItem struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	ImageUrl    string   `json:"image_url"`
}

type StoreClient struct {
	config     config.Store
	Packages   []PackageListItem
	Categories []string
}

func NewStoreClient(config config.Store) *StoreClient {
	return &StoreClient{
		config: config,
	}
}

// UpdatePackageList updates the package list contained in the StoreClient struct
func (client *StoreClient) UpdatePackageList() error {
	// Retrieve package list over HTTP
	resp, err := http.Get(client.config.StoreUrl)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Parse package list
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &client.Packages)
	if err != nil {
		return err
	}

	//  Loops through the list setting the icon URL and adding the categories to the global categories list
	for i := range client.Packages {
		client.Packages[i].ImageUrl = strings.Trim(client.config.StoreUrl, "list.json") + "packages/" + client.Packages[i].Id + "/icon.png"

		for _, category := range client.Packages[i].Categories {
			// If the list doesn't already contain the category insert it at the correct position alphabetically
			if !slices.Contains(client.Categories, category) {
				insertIndex := sort.Search(len(client.Categories), func(i int) bool {
					return client.Categories[i] >= category
				})
				client.Categories = append(client.Categories, "")
				copy(client.Categories[insertIndex+1:], client.Categories[insertIndex:])
				client.Categories[insertIndex] = category
			}
		}
	}

	return nil
}

func (client *StoreClient) GetPackage(packageId string) (appPackage persistence.AppPackage, err error) {
	// Retrieve package file
	packagePath := strings.Trim(client.config.StoreUrl, "list.json") + "packages/" + packageId + "/schema.json"
	resp, err := http.Get(packagePath)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("invalid HTTP response %d", resp.StatusCode)
		return
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &appPackage)
	return
}

func (client *StoreClient) SearchPackages(search string) []PackageListItem {
	packages := make([]PackageListItem, 0)
	searchTerm := strings.ToLower(strings.TrimSpace(search))
	for _, appPackage := range client.Packages {
		if strings.Contains(strings.ToLower(appPackage.Name), searchTerm) {
			packages = append(packages, appPackage)
		}
	}
	return packages
}
