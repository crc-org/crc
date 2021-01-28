package extract

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fileMap map[string]string

var (
	files fileMap = map[string]string{
		filepath.Join("a", "b", "c.txt"): "ccc",
		filepath.Join("a", "d", "e.txt"): "eee",
	}
	filteredFiles fileMap = map[string]string{
		filepath.Join("a", "b", "c.txt"): "ccc",
	}
	archives = []string{
		"test.tar",
		"test.tar.gz",
		"test.zip",
		"test.tar.xz",
		"test.tar.zst",
	}
)

func TestUncompress(t *testing.T) {
	for _, archive := range archives {
		assert.NoError(t, testUncompress(filepath.Join("testdata", archive), nil, files))
		assert.NoError(t, testUncompress(filepath.Join("testdata", archive), fileFilter, filteredFiles))
	}
}

func TestUnCompressBundle(t *testing.T) {
	dir, err := ioutil.TempDir("", "bundles")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	bundle := filepath.Join(dir, "test.crcbundle")
	for _, archive := range archives {
		require.NoError(t, crcos.CopyFileContents(filepath.Join("testdata", archive), bundle, 0600))
		assert.NoError(t, testUncompress(bundle, nil, files))
		assert.NoError(t, testUncompress(bundle, fileFilter, filteredFiles))
	}
}

func copyFileMap(orig fileMap) fileMap {
	copy := fileMap{}
	for key, value := range orig {
		copy[key] = value
	}
	return copy
}

// This checks that the list of files returned by Uncompress matches what we expect
func checkFileList(destDir string, extractedFiles []string, expectedFiles fileMap) error {
	// We are going to remove elements from  the map, but we don't want to modify the map used by the caller
	expectedFiles = copyFileMap(expectedFiles)

	for _, file := range extractedFiles {
		rel, err := filepath.Rel(destDir, file)
		if err != nil {
			return err
		}
		_, found := expectedFiles[rel]
		if !found {
			return fmt.Errorf("Unexpected file '%s' in file list %v", rel, expectedFiles)
		}
		delete(expectedFiles, rel)
	}

	if len(expectedFiles) != 0 {
		return fmt.Errorf("Some expected files were not in file list: %v", expectedFiles)
	}

	return nil
}

// This checks that the files in the destination directory matches what we expect
func checkFiles(destDir string, files fileMap) error {
	// We are going to remove elements from  the map, but we don't want to modify the map used by the caller
	files = copyFileMap(files)

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		logging.Debugf("Walking %s", path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			logging.Debugf("Skipping directory %s", path)
			return nil
		}
		logging.Debugf("Checking file %s", path)
		archivePath, err := filepath.Rel(destDir, path)
		if err != nil {
			return err
		}
		expectedContent, found := files[archivePath]
		if !found {
			return fmt.Errorf("Unexpected extracted file '%s'", path)
		}
		delete(files, archivePath)

		data, err := ioutil.ReadFile(path) // #nosec G304
		if err != nil {
			return err
		}
		if string(data) != expectedContent {
			return fmt.Errorf("Unexpected content for '%s': expected [%s], got [%s]", path, expectedContent, string(data))
		}
		logging.Debugf("'%s' successfully checked", path)
		return nil
	})
	if err != nil {
		return err
	}
	if len(files) != 0 {
		return fmt.Errorf("Some expected files were not extracted: %v", files)
	}

	return nil
}

func testUncompress(archiveName string, fileFilter func(string) bool, files fileMap) error {
	destDir, err := ioutil.TempDir("", "crc-extract-test")
	if err != nil {
		return err
	}
	defer os.RemoveAll(destDir)

	var fileList []string
	if fileFilter != nil {
		fileList, err = UncompressWithFilter(archiveName, destDir, false, fileFilter)
	} else {
		fileList, err = Uncompress(archiveName, destDir, false)
	}
	if err != nil {
		return err
	}

	err = checkFileList(destDir, fileList, files)
	if err != nil {
		return err
	}

	return checkFiles(destDir, files)
}

func fileFilter(filename string) bool {
	return filepath.Base(filename) == "c.txt"
}
