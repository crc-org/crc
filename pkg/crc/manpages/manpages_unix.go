//go:build !windows

package manpages

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra/doc"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

var (
	rootCrcManPage             = "crc.1.gz"
	osEnvGetter                = os.Getenv
	osEnvSetter                = os.Setenv
	ManPathEnvironmentVariable = "MANPATH"
	CrcManPageHeader           = &doc.GenManHeader{
		Title:   "CRC",
		Section: "1",
	}
)

// GenerateManPages generates manual pages for user commands and places them
// in the specified target directory. It performs the following steps:
//
// 1. Checks if the man pages should be generated based on the target folder.
// 2. Creates the necessary directory structure if it does not exist.
// 3. Generates man pages in a temporary directory.
// 4. Compresses the generated man pages and moves them to the target folder.
// 5. Updates the MANPATH environment variable to include the target directory.
// 6. Cleans up the temporary directory.
//
// manPageGenerator: Function that generates man pages in the specified directory.
// targetDir: Directory where the generated man pages should be placed.
//
// Returns an error if any step in the process fails.
func GenerateManPages(manPageGenerator func(targetDir string) error, targetDir string) error {
	manUserCommandTargetFolder := filepath.Join(targetDir, "man1")
	if !manPagesAlreadyGenerated(manUserCommandTargetFolder) {
		if _, err := os.Stat(manUserCommandTargetFolder); os.IsNotExist(err) {
			err = os.MkdirAll(manUserCommandTargetFolder, 0755)
			if err != nil {
				logging.Errorf("error in creating dir for man pages: %s", err.Error())
			}
		}
		temporaryManPagesDir, err := generateManPagesInTemporaryDirectory(manPageGenerator)
		if err != nil {
			return err
		}
		err = compressManPages(temporaryManPagesDir, manUserCommandTargetFolder)
		if err != nil {
			return fmt.Errorf("error in compressing man pages: %s", err.Error())
		}
		err = appendToManPathEnvironmentVariable(targetDir)
		if err != nil {
			return fmt.Errorf("error updating MANPATH environment variable: %s", err.Error())
		}
		err = os.RemoveAll(temporaryManPagesDir)
		if err != nil {
			return fmt.Errorf("error removing temporary man pages directory: %s", err.Error())
		}
	}
	return nil
}

func appendToManPathEnvironmentVariable(folder string) error {
	manPath := osEnvGetter(ManPathEnvironmentVariable)
	if !manPathAlreadyContains(folder, manPath) {
		if manPath == "" {
			manPath = folder
		} else {
			manPath = fmt.Sprintf("%s%c%s", manPath, os.PathListSeparator, folder)
		}
		err := osEnvSetter(ManPathEnvironmentVariable, manPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func manPathAlreadyContains(manPathEnvVarValue string, folder string) bool {
	manDirs := strings.Split(manPathEnvVarValue, string(os.PathListSeparator))
	return slices.Contains(manDirs, folder)
}

func removeFromManPathEnvironmentVariable(manPathEnvVarValue string, folder string) error {
	manDirs := strings.Split(manPathEnvVarValue, string(os.PathListSeparator))
	var updatedManPathEnvVarValues []string
	for _, manDir := range manDirs {
		if manDir != folder {
			updatedManPathEnvVarValues = append(updatedManPathEnvVarValues, manDir)
		}
	}
	return osEnvSetter(ManPathEnvironmentVariable, strings.Join(updatedManPathEnvVarValues, string(os.PathListSeparator)))
}

func generateManPagesInTemporaryDirectory(manPageGenerator func(targetDir string) error) (string, error) {
	tempDir, err := os.MkdirTemp("", "crc-manpages")
	if err != nil {
		return "", err
	}
	manPagesGenerationErr := manPageGenerator(tempDir)
	if manPagesGenerationErr != nil {
		return "", manPagesGenerationErr
	}
	logging.Debugf("Successfully generated manpages in %s", tempDir)
	return tempDir, nil
}

func compressManPages(manPagesSourceFolder string, manPagesTargetFolder string) error {
	return filepath.Walk(manPagesSourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		compressedFilePath := filepath.Join(manPagesTargetFolder, info.Name()+".gz")
		compressedFile, err := os.Create(compressedFilePath)
		if err != nil {
			return err
		}
		defer compressedFile.Close()

		gzipWriter := gzip.NewWriter(compressedFile)
		defer gzipWriter.Close()

		_, err = io.Copy(gzipWriter, srcFile)
		if err != nil {
			return err
		}
		return nil
	})
}

func manPagesAlreadyGenerated(manPagesTargetFolder string) bool {
	rootCrcManPageFilePath := filepath.Join(manPagesTargetFolder, rootCrcManPage)
	if _, err := os.Stat(rootCrcManPageFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func RemoveCrcManPages(manPageDir string) error {
	manUserCommandTargetFolder := filepath.Join(manPageDir, "man1")
	err := filepath.Walk(manUserCommandTargetFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path)[:len("crc")] == "crc" {
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return removeFromManPathEnvironmentVariable(osEnvGetter(ManPathEnvironmentVariable), manPageDir)
}
