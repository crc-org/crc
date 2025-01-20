package manpages

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/stretchr/testify/assert"
)

func TestGenerateManPages_ShouldDoNothingOnWindows(t *testing.T) {
	// Given
	dir := t.TempDir()
	// When
	err := GenerateManPages(func(targetDir string) error {
		return doc.GenManTree(&cobra.Command{
			Use: "crc",
		}, CrcManPageHeader, targetDir)
	}, dir)
	// Then
	assert.NoError(t, err)
}

func TestRemoveCrcManPages_ShouldDoNothingOnWindows(t *testing.T) {
	// Given
	dir := t.TempDir()
	// When
	err := RemoveCrcManPages(dir)
	// Then
	assert.NoError(t, err)
}
