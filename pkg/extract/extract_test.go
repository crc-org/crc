package extract

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type FileMap map[string]string

func copyFileMap(orig FileMap) FileMap {
	copy := FileMap{}
	for key, value := range orig {
		copy[key] = value
	}

	return copy
}

// This checks that the list of files returned by Uncompress matches what we expect
func checkFileList(destDir string, extractedFiles []string, expectedFiles FileMap) error {
	// We are going to remove elements from  the map, but we don't want to modify the map used by the caller
	expectedFiles = copyFileMap(expectedFiles)

	for _, file := range extractedFiles {
		logging.Debugf("Checking %s", file)
		file = strings.TrimPrefix(file, destDir)
		_, found := expectedFiles[file]
		if !found {
			return fmt.Errorf("Unexpected file '%s' in file list", file)
		}
		delete(expectedFiles, file)
	}

	if len(expectedFiles) != 0 {
		return fmt.Errorf("Some expected files were not in file list: %v", expectedFiles)
	}

	return nil
}

// This checks that the files in the destination directory matches what we expect
func checkFiles(destDir string, files FileMap) error {
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
		archivePath := strings.TrimPrefix(path, destDir)
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

func testUncompress(archiveName string, fileFilter func(string) bool, files FileMap) error {
	destDir, err := ioutil.TempDir("", "crc-extract-test")
	if err != nil {
		return err
	}
	err = os.MkdirAll(destDir, 0700)
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

func TestUncompress(t *testing.T) {
	var files FileMap = map[string]string{
		"/a/b/c.txt": "ccc",
		"/a/d/e.txt": "eee",
	}
	var filteredFiles FileMap = map[string]string{
		"/a/b/c.txt": "ccc",
	}

	err := testUncompress("testdata/test.tar.gz", nil, files)
	if err != nil {
		t.Errorf("Failed to uncompress test.tar.gz: %v", err)
	}

	err = testUncompress("testdata/test.zip", nil, files)
	if err != nil {
		t.Errorf("Failed to uncompress test.zip: %v", err)
	}

	err = testUncompress("testdata/test.tar.xz", nil, files)
	if err != nil {
		t.Errorf("Failed to uncompress test.tar.xz: %v", err)
	}

	err = testUncompress("testdata/test.tar.gz", fileFilter, filteredFiles)
	if err != nil {
		t.Errorf("Failed to uncompress c.txt from test.tar.gz: %v", err)
	}

	err = testUncompress("testdata/test.zip", fileFilter, filteredFiles)
	if err != nil {
		t.Errorf("Failed to uncompress c.txt from test.zip: %v", err)
	}

	err = testUncompress("testdata/test.tar.xz", fileFilter, filteredFiles)
	if err != nil {
		t.Errorf("Failed to uncompress c.txt from test.tar.xz: %v", err)
	}
}
