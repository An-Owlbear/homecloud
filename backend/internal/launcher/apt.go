package launcher

import (
	"os/exec"
	"strings"
)

// CheckUpdates returns whether new system updates are available
func CheckUpdates() (bool, error) {
	// Updates package list
	cmd := exec.Command("apt", "update")
	_, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Runs upgrade dry run to check if new packages need installing
	upgrade, err := exec.Command("apt", "upgrade", "-s").Output()
	if err != nil {
		return false, err
	}
	// If any line begins with `Inst ` then return true, otherwise return false
	for _, line := range strings.Split(string(upgrade), "\n") {
		if strings.HasPrefix(line, "Inst ") {
			return true, nil
		}
	}
	return false, nil
}

func ApplyUpdates() error {
	_, err := exec.Command("apt", "upgrade", "-y").Output()
	if err != nil {
		return err
	}

	return nil
}
