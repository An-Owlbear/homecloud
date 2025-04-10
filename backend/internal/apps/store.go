package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
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

func NewPackageListItem(app persistence.GetAppsRow) PackageListItem {
	return PackageListItem{
		Id:          app.ID,
		Name:        app.Schema.Name,
		Version:     app.Schema.Version,
		Author:      app.Schema.Author,
		Description: app.Schema.Description,
		Categories:  app.Schema.Categories,
		ImageUrl:    "/assets/data/" + app.ID + "/icon.png",
	}
}

type StoreClient struct {
	config config.Store
}

func NewStoreClient(config config.Store) *StoreClient {
	return &StoreClient{
		config: config,
	}
}

// UpdatePackageList updates the package list contained in the StoreClient struct
func (client *StoreClient) UpdatePackageList(ctx context.Context, queries *persistence.Queries) error {
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

	var packages []persistence.FullPackageListItem
	err = json.Unmarshal(body, &packages)
	if err != nil {
		return err
	}

	for _, appPackage := range packages {
		appPackage.ImageUrl = strings.Trim(client.config.StoreUrl, "list.json") + "packages/" + appPackage.ID + "/icon.png"
		err = queries.InsertPackage(ctx, appPackage)
		if err != nil {
			return err
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
