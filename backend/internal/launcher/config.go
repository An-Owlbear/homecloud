package launcher

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// This file is used for the configuration file managed by the launcher
// Unlike config/launcher.go which is for the environment variables for the launcher

type Config struct {
	filename  string
	Subdomain string `json:"address"`
}

func SetupConfig(filename string) (*Config, error) {
	var config *Config

	// If config doesn't exist create a file for it and write the default values, otherwise read from the existing file
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return nil, err
		}
		file, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("error creating config file: %w", err)
		}
		file.Close()

		config = &Config{Subdomain: ""}
		config.filename = filename
		if err := config.Save(); err != nil {
			return nil, fmt.Errorf("error saving new config file: %w", err)
		}
	} else {
		configReader, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("error opening config file: %w", err)
		}
		defer configReader.Close()

		fileContents, err := io.ReadAll(configReader)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}

		err = json.Unmarshal(fileContents, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
	}

	config.filename = filename
	return config, nil
}

func (c *Config) Save() error {
	writer, err := os.Create(c.filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer writer.Close()

	configString, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	_, err = writer.Write(configString)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}
