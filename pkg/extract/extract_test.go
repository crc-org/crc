package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcos "github.com/crc-org/crc/v2/pkg/os"
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
		assert.NoError(t, testUncompress(t, filepath.Join("testdata", archive), nil, files))
		assert.NoError(t, testUncompress(t, filepath.Join("testdata", archive), fileFilter, filteredFiles))
	}
}

func TestUnCompressBundle(t *testing.T) {
	dir := t.TempDir()

	bundle := filepath.Join(dir, "test.crcbundle")
	for _, archive := range archives {
		require.NoError(t, crcos.CopyFileContents(filepath.Join("testdata", archive), bundle, 0600))
		assert.NoError(t, testUncompress(t, bundle, nil, files))
		assert.NoError(t, testUncompress(t, bundle, fileFilter, filteredFiles))
	}
}

// check that archives containing ./ are extracted as expected
func TestDotSlash(t *testing.T) {
	var files fileMap = map[string]string{
		"a.txt": "",
	}
	assert.NoError(t, testUncompress(t, filepath.Join("testdata", "dotslash.tar.gz"), nil, files))
}

func TestZipSlip(t *testing.T) {
	archiveName := filepath.Join("testdata", "zipslip.tar.gz")
	_, err := Uncompress(archiveName, t.TempDir())
	logging.Infof("error: %v", err)
	assert.ErrorContains(t, err, "illegal file path")
}

func copyFileMap(orig fileMap) fileMap {
	copiedMap := fileMap{}
	for key, value := range orig {
		copiedMap[key] = value
	}
	return copiedMap
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

		data, err := os.ReadFile(path) // #nosec G304
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

func testUncompress(t *testing.T, archiveName string, fileFilter func(string) bool, files fileMap) error {
	destDir := t.TempDir()

	var fileList []string
	var err error
	if fileFilter != nil {
		fileList, err = UncompressWithFilter(archiveName, destDir, fileFilter)
	} else {
		fileList, err = Uncompress(archiveName, destDir)
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
