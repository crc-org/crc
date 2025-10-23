//go:build !windows

package manpages

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/stretchr/testify/assert"
)

func TestShouldGenerateManPages_WhenManPagesAlreadyExists_ShouldReturnTrue(t *testing.T) {
	// Given
	dir := t.TempDir()
	crcManFile := filepath.Join(dir, "crc.1.gz")
	_, err := os.Create(crcManFile)
	assert.NoError(t, err)
	// When
	result := manPagesAlreadyGenerated(dir)

	// Then
	assert.Equal(t, true, result)
}

func TestGenerateManPagesInTemporaryDirectory(t *testing.T) {
	// Given
	dir := t.TempDir()
	osEnv := make(map[string]string)
	osEnvGetter = func(key string) string {
		return osEnv[key]
	}
	osEnvSetter = func(key, value string) error {
		osEnv[key] = value
		return nil
	}

	// When
	err := GenerateManPages(func(targetDir string) error {
		return doc.GenManTree(&cobra.Command{
			Use: "crc",
		}, CrcManPageHeader, targetDir)
	}, dir)

	// Then
	assert.NoError(t, err)
	assert.NotEmptyf(t, dir, "No manpages man1 directory created")
	userCommandManDir := filepath.Join(dir, "man1")
	assert.DirExists(t, userCommandManDir)
	files, err := os.ReadDir(userCommandManDir)
	assert.NoError(t, err)
	var manPagesFiles []string
	for _, manPage := range files {
		manPagesFiles = append(manPagesFiles, manPage.Name())
	}
	assert.ElementsMatch(t, []string{
		"crc.1.gz",
	}, manPagesFiles)
	assert.Equal(t, osEnvGetter("MANPATH"), dir)
}

func TestUpdateManPathEnvironmentVariable_GivenManPath_ShouldWriteOrAppendToManPath(t *testing.T) {
	tests := []struct {
		manPathOldValue         string
		expectedManPathNewValue string
	}{
		{"", "/home/foo/.local/share/man"},
		{"/usr/local/share/man/:/usr/share/man", "/usr/local/share/man/:/usr/share/man:/home/foo/.local/share/man"},
	}

	for _, tt := range tests {
		t.Run(tt.manPathOldValue, func(t *testing.T) {
			// Given
			osEnv := make(map[string]string)
			osEnv["MANPATH"] = tt.manPathOldValue
			osEnvGetter = func(key string) string {
				return osEnv[key]
			}
			osEnvSetter = func(key, value string) error {
				osEnv[key] = value
				return nil
			}

			// When
			err := appendToManPathEnvironmentVariable("/home/foo/.local/share/man")

			// Then
			assert.NoError(t, err)
			assert.Equal(t, osEnv["MANPATH"], tt.expectedManPathNewValue)
		})
	}
}

func TestManPathAlreadyContains(t *testing.T) {
	tests := []struct {
		manPathValue    string
		manPathToSearch string
		expected        bool
	}{
		{"/usr/local/share/man/:/usr/share/man:/home/foo/.local/share/man", "/home/foo/.local/share/man", true},
		{"/usr/local/share/man/:/usr/share/man", "/home/foo/.local/share/man", false},
		{"/usr/local/share/man/:/usr/share/man", "/usr/local/man", false},
	}

	for _, test := range tests {
		t.Run(test.manPathValue, func(t *testing.T) {
			// Given
			// When
			result := manPathAlreadyContains(test.manPathValue, test.manPathToSearch)
			// Then
			if result != test.expected {
				t.Errorf("manPathAlreadyContains('%s'), expected '%t' but got '%t'", test.manPathValue, test.expected, result)
			}
		})
	}
}

func TestRemoveCrcManPages(t *testing.T) {
	// Given
	dir := t.TempDir()
	osEnv := make(map[string]string)
	osEnv["MANPATH"] = fmt.Sprintf("/usr/local/share/man%c%s", os.PathListSeparator, dir)
	osEnvGetter = func(key string) string {
		return osEnv[key]
	}
	osEnvSetter = func(key, value string) error {
		osEnv[key] = value
		return nil
	}

	// When
	err := GenerateManPages(func(targetDir string) error {
		return doc.GenManTree(&cobra.Command{
			Use: "crc",
		}, CrcManPageHeader, targetDir)
	}, dir)
	assert.NoError(t, err)
	err = RemoveCrcManPages(dir)

	// Then
	assert.NoError(t, err)
	userCommandManDir := filepath.Join(dir, "man1")
	assert.DirExists(t, userCommandManDir)
	manPages, err := os.ReadDir(userCommandManDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(manPages))
	assert.Equal(t, "/usr/local/share/man", osEnv["MANPATH"])
}
