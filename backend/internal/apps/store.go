package apps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/An-Owlbear/homecloud/backend/internal/persistence"
)

type PackageListItem struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

type StoreClient struct {
	repoUrl  string
	packages []PackageListItem
}

func NewStoreClient(repoUrl string) *StoreClient {
	return &StoreClient{
		repoUrl: repoUrl,
	}
}

// UpdatePackageList updates the package list contained in the StoreClient struct
func (client StoreClient) UpdatePackageList() error {
	// Retrieve package list over HTTP
	resp, err := http.Get(client.repoUrl)
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

	err = json.Unmarshal(body, &client.packages)
	if err != nil {
		return err
	}

	return nil
}

func (client StoreClient) GetPackage(packageId string) (appPackage persistence.AppPackage, err error) {
	// Retrieve package file
	packagePath := strings.Trim(client.repoUrl, "list.json") + "packages/" + packageId + ".json"
	resp, err := http.Get(packagePath)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Invalid HTTP response %d", resp.StatusCode))
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
