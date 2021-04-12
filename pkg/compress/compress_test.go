package compress

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/extract"

	"github.com/stretchr/testify/require"
)

const testArchiveName = "testdata.tar.zst"

var (
	files fileMap = map[string]string{
		filepath.Join("a"):      "a",
		filepath.Join("c"):      "c",
		filepath.Join("b", "d"): "d",
	}
)

func testCompress(t *testing.T, baseDir string) {
	// This is useful to check that the top-level directory of the archive
	// was created with the correct permissions, and was not created by the
	// `os.MkdirAll(0750)` call at the beginning of `untarFile`
	require.NoError(t, os.Chmod(baseDir, 0700))
	require.NoError(t, Compress(baseDir, testArchiveName))
	defer os.Remove(testArchiveName)

	destDir, err := ioutil.TempDir("", "testdata-extracted")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	fileList, err := extract.Uncompress(testArchiveName, destDir, false)
	require.NoError(t, err)

	_, d := filepath.Split(baseDir)
	fi, err := os.Stat(filepath.Join(destDir, d))
	require.NoError(t, err)
	fMode := os.FileMode(0700)
	if runtime.GOOS == "windows" {
		// https://golang.org/pkg/os/#Chmod
		fMode = os.FileMode(0777)
	}
	require.Equal(t, fi.Mode().Perm(), fMode)

	require.NoError(t, checkFileList(filepath.Join(destDir, d), fileList, files))
	require.NoError(t, checkFiles(filepath.Join(destDir, d), files))
}

func TestCompressRelative(t *testing.T) {
	// Test with relative path
	testCompress(t, "testdata")

	// Test with absolute path
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	testCompress(t, filepath.Join(currentDir, "testdata"))
}

/* The code below is duplicated from pkg/extract/extract_test.go */
type fileMap map[string]string

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
