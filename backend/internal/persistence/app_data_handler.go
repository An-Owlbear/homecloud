package persistence

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"github.com/An-Owlbear/homecloud/backend/internal/config"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
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

func (h *AppDataHandler) RenderTemplates(
	ctx context.Context,
	queries *Queries,
	oryConfig config.Ory,
	hostConfig config.Host,
	appId string,
) error {
	dataPath := filepath.Join(h.storageConfig.DataPath, appId, "data")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return nil
	}

	err := filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".tmpl" {
			appInfo, err := queries.GetAppWithCreds(ctx, appId)

			templateFile, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			writer, err := os.Create(strings.TrimSuffix(path, ".tmpl"))
			if err != nil {
				return err
			}

			if err := ApplyAppTemplate(string(templateFile), writer, appInfo.Schema, appInfo.ClientID.String, appInfo.ClientSecret.String, oryConfig, hostConfig, h.storageConfig); err != nil {
				return err
			}
			if err := writer.Close(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
