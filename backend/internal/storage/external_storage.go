package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ListExternalStorage() ([]string, error) {
	// Retrieves list of devices. This only works on linux systems, but since that is the main target it will be no
	// problem
	devices, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("error reading devices from /sys/block: %w", err)
	}

	externalDevices := make([]string, 0)
	for _, device := range devices {
		reader, err := os.Open(filepath.Join("/sys/block", device.Name(), "removable"))
		if err != nil {
			return nil, fmt.Errorf("error opening device %s: %w", device.Name(), err)
		}
		contents, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("error reading devices from /sys/block: %w", err)
		}

		stringContents := string(contents)
		if stringContents == "1\n" || stringContents == "1" {
			externalDevices = append(externalDevices, device.Name())
		}
	}

	return externalDevices, nil
}
