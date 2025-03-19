package testutils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

const SharedDirectory = "/tmp/homecloud_testing"
const TempDirectory = "/tmp/homecloud_testing_temp"

func SetupTempStorage() config.Storage {
	os.MkdirAll(TempDirectory, 0755)
	return config.Storage{
		DataPath: filepath.Join(TempDirectory, "data"),
	}
}

func CleanupSharedStorage() {
	entries, err := os.ReadDir(SharedDirectory)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(SharedDirectory, entry.Name())); err != nil {
			panic(err)
		}
	}
}

func CheckTarArchive(t *testing.T, archivePath string, expectedFiles map[string]string) {
	fileReader, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		t.Fatal(err)
	}
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			filename := strings.TrimPrefix(header.Name, "./")
			expected, ok := expectedFiles[filename]
			if !ok {
				t.Fatalf("file %s in tar not expected", filename)
			}
			fileContents, err := io.ReadAll(tarReader)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(expected, string(fileContents)); diff != "" {
				t.Fatal(diff)
			}
			delete(expectedFiles, filename)
		}
	}

	if len(expectedFiles) > 0 {
		t.Fatalf("did not find expected files: %v", expectedFiles)
	}
}

func CreateTarArchive(archiveDir string, testData map[string]string) (string, error) {
	filesDir := filepath.Join(archiveDir, "files")
	for filename, content := range testData {
		if err := os.MkdirAll(filepath.Join(filesDir, filepath.Dir(filename)), 0777); err != nil {
			return "", err
		}

		writer, err := os.Create(filepath.Join(filesDir, filename))
		if err != nil {
			return "", err
		}
		if _, err := writer.WriteString(content); err != nil {
			return "", err
		}
		writer.Close()
	}

	archivePath := filepath.Join(archiveDir, "data.tar.gz")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer archiveFile.Close()
	gzipWriter := gzip.NewWriter(archiveFile)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	dirfs := os.DirFS(filesDir)
	if err := tarWriter.AddFS(dirfs); err != nil {
		return "", err
	}

	return archivePath, nil
}
