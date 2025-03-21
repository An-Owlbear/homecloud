//go:build linux

package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"
)

type DriveInfo struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Size      uint64 `json:"size"`
	Available uint64 `json:"available"`
}

type LsblkJson struct {
	BlockDevices []LsblkDevice `json:"blockdevices"`
}

type LsblkJsonSingle struct {
	BlockDevices []LsblkDetails `json:"blockdevices"`
}

type LsblkDevice struct {
	LsblkDetails
	Children []LsblkDetails `json:"children"`
}

type LsblkDetails struct {
	Name      string `json:"name"`
	RM        bool   `json:"rm"`
	RO        bool   `json:"ro"`
	PartLabel string `json:"partlabel"`
	Size      uint64 `json:"size"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	FSType    string `json:"fstype"`
}

var DriveInfoError = errors.New("invalid drive info from system")
var DriveInvalidError = errors.New("no valid drive found")

var filterLabels = []string{"boot", "efi", "efi system partition", "reserved", "swap", "system reserved"}

const lsblkColumns = "NAME,RM,RO,PARTLABEL,SIZE,LABEL,TYPE,FSTYPE"

func ListExternalStorage() ([]DriveInfo, error) {
	// Retrieves list of devices. This only works on linux systems, but since that is the main target it will be no
	// problem
	cmd := exec.Command("lsblk", "-I", "8", "-b", "-J", "-o", lsblkColumns)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error listing external storage: %w", err)
	}

	var response LsblkJson
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
				Name:      partition.Name,
				Label:     partition.Label,
				Size:      partition.Size,
				Available: partition.Size,
			})
		}
	}

	return externalDevices, nil
}

// GetExternalPartition checks if the specified partition is valid
func GetExternalPartition(device string) (LsblkDetails, error) {
	// Retrieves information about the drive
	cmd := exec.Command("lsblk", "-I", "8", "-b", "-J", "-o", lsblkColumns, "/dev/"+device)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return LsblkDetails{}, fmt.Errorf("error getting drive details: %w", err)
	}

	var response LsblkJsonSingle
	if err := json.Unmarshal(output, &response); err != nil {
		return LsblkDetails{}, fmt.Errorf("error parsing lsblk output: %w", err)
	}
	if len(response.BlockDevices) != 1 {
		return LsblkDetails{}, DriveInvalidError
	}

	partitionValid, err := checkLsblkDetails(response.BlockDevices[0])
	if err != nil {
		return LsblkDetails{}, err
	}
	if !partitionValid {
		return LsblkDetails{}, DriveInvalidError
	}

	return response.BlockDevices[0], nil
}

func checkLsblkDetails(details LsblkDetails) (bool, error) {
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

// MountPartition mounts the given partition, returning a string indicating the path it was mounted to
func MountPartition(details LsblkDetails) (string, error) {
	source := fmt.Sprintf("/dev/%s", details.Name)
	target := fmt.Sprintf("/media/homecloud/%s", details.Name)

	if err := os.MkdirAll(target, 0755); err != nil {
		return "", fmt.Errorf("error creating mount directory %s: %w", target, err)
	}

	err := unix.Mount(source, target, details.FSType, 0, "")
	if err != nil {
		return "", fmt.Errorf("error mounting partition %s to %s: %w", source, target, err)
	}

	return target, nil
}

// UnmountPartition unmounts the given partition
func UnmountPartition(details LsblkDetails) error {
	err := unix.Unmount(fmt.Sprintf("/media/homecloud/%s", details.Name), 0)
	if err != nil {
		return fmt.Errorf("error unmounting partition: %w", err)
	}
	return nil
}

func ListBackups(targetDevice string, appId string) ([]string, error) {
	details, err := GetExternalPartition(targetDevice)
	if err != nil {
		return nil, err
	}

	mountPath, err := MountPartition(details)
	if err != nil {
		return nil, err
	}
	defer UnmountPartition(details)

	entries, err := os.ReadDir(filepath.Join(mountPath, "backup", appId))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}

	backups := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}
