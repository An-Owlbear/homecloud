package persistence

import (
	"archive/tar"
	"compress/gzip"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type AppDataHandler struct {
	storageConfig config.Storage
	storeConfig   config.Store
	http          *http.Client
}

func NewAppDataHandler(storageConfig config.Storage, storeConfig config.Store) *AppDataHandler {
	return &AppDataHandler{
		storageConfig: storageConfig,
		storeConfig:   storeConfig,
		http:          &http.Client{},
	}
}

func (h *AppDataHandler) SavePackage(appId string) error {
	packageUrl, err := url.JoinPath(strings.Trim(h.storeConfig.StoreUrl, "list.json"), "packages", appId, "package.tar.gz")
	if err != nil {
		return err
	}

	resp, err := h.http.Get(packageUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	uncompressed, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer uncompressed.Close()
	tarReader := tar.NewReader(uncompressed)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path.Join(h.storageConfig.DataPath, appId, header.Name), 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(path.Join(h.storageConfig.DataPath, appId, header.Name))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
