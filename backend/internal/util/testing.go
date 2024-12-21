package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// RootDir retrieves the root directory of the project. Mainly for use with tests, where the
// working directory isn't the root directory
func RootDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found "))
		}
		currentDir = parent
	}

	return currentDir
}
