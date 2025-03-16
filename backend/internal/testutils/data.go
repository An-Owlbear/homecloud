package testutils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

func SetupTempStorage() config.Storage {
	// TODO: clear storage directory
	return config.Storage{
		DataPath: "/tmp/homecloud_test/storage",
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
