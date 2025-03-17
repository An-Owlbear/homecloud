package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
)

type DriveInfo struct {
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	Available uint64 `json:"available"`
}

type lsblkJson struct {
	BlockDevices []lsblkDevice `json:"blockdevices"`
}

type lsblkJsonSingle struct {
	BlockDevices []lsblkDetails `json:"blockdevices"`
}

type lsblkDevice struct {
	lsblkDetails
	Children []lsblkDetails `json:"children"`
}

type lsblkDetails struct {
	Name      string `json:"name"`
	RM        bool   `json:"rm"`
	RO        bool   `json:"ro"`
	PartLabel string `json:"partlabel"`
	Size      uint64 `json:"size"`
	Label     string `json:"label"`
	Type      string `json:"type"`
}

var DriveInfoError = errors.New("invalid drive info from system")
var DriveInvalidError = errors.New("no valid drive found")

var filterLabels = []string{"boot", "efi", "efi system partition", "reserved", "swap", "system reserved"}

const lsblkColumns = "NAME,RM,RO,PARTLABEL,SIZE,LABEL,TYPE"

func ListExternalStorage() ([]DriveInfo, error) {
	// Retrieves list of devices. This only works on linux systems, but since that is the main target it will be no
	// problem
	cmd := exec.Command("lsblk", "-I", "8", "-b", "-J", "-o", lsblkColumns)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error listing external storage: %w", err)
	}

	var response lsblkJson
	if err := json.Unmarshal(out, &response); err != nil {
		return nil, fmt.Errorf("error parsing lsblk output: %w", err)
	}

	externalDevices := make([]DriveInfo, 0)
	for _, blockDevice := range response.BlockDevices {
		for _, partition := range blockDevice.Children {
			slog.Error(fmt.Sprintf("%+v", partition))
			partitionValid, err := checkLsblkDetails(partition)
			if err != nil {
				return nil, err
			}
			if !partitionValid {
				continue
			}

			externalDevices = append(externalDevices, DriveInfo{
				Name:      partition.Label,
				Size:      partition.Size,
				Available: partition.Size,
			})
		}
	}

	return externalDevices, nil
}

// GetExternalPartition checks if the specified partition is valid
func GetExternalPartition(device string) (DriveInfo, error) {
	// Retrieves information about the drive
	cmd := exec.Command("lsblk", "-I", "8", "-b", "-J", "-o", lsblkColumns, "/dev/"+device)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return DriveInfo{}, fmt.Errorf("error getting drive details: %w", err)
	}

	var response lsblkJsonSingle
	if err := json.Unmarshal(output, &response); err != nil {
		return DriveInfo{}, fmt.Errorf("error parsing lsblk output: %w", err)
	}
	if len(response.BlockDevices) != 1 {
		return DriveInfo{}, DriveInvalidError
	}

	partitionValid, err := checkLsblkDetails(response.BlockDevices[0])
	if err != nil {
		return DriveInfo{}, err
	}
	if !partitionValid {
		return DriveInfo{}, DriveInvalidError
	}

	return DriveInfo{
		Name:      response.BlockDevices[0].Label,
		Size:      response.BlockDevices[0].Size,
		Available: response.BlockDevices[0].Size,
	}, nil
}

func checkLsblkDetails(details lsblkDetails) (bool, error) {
	// If the entry isn't a partition return false
	if details.Type != "part" {
		return false, nil
	}

	// If the drive is not removable return false
	if !details.RM {
		return false, nil
	}

	// If the drive is labelled readonly return false
	if details.RO {
		return false, nil
	}

	// If the drive has a name suspected of being a read only drive return false
	// This is estimated by checking if the name matches common system partitions and has a small size
	// This is extremely unlikely to falsely flag user partitions, since it would need to be both given one of these
	// labels and be below 1GB
	if slices.Contains(filterLabels, strings.ToLower(details.PartLabel)) && details.Size < 1e9 {
		return false, nil
	}

	return true, nil
}
