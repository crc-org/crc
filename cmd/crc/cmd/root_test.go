package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrcManPageGenerator_WhenInvoked_GeneratesManPagesForAllCrcSubCommands(t *testing.T) {
	// Given
	dir := t.TempDir()

	// When
	err := crcManPageGenerator(dir)

	// Then
	assert.NoError(t, err)
	files, readErr := os.ReadDir(dir)
	assert.NoError(t, readErr)
	var manPagesFiles []string
	for _, manPage := range files {
		manPagesFiles = append(manPagesFiles, manPage.Name())
	}
	assert.ElementsMatch(t, []string{
		"crc-bundle-generate.1",
		"crc-bundle.1",
		"crc-cleanup.1",
		"crc-config-get.1",
		"crc-config-set.1",
		"crc-config-unset.1",
		"crc-config-view.1",
		"crc-config.1",
		"crc-console.1",
		"crc-delete.1",
		"crc-generate-kubeconfig.1",
		"crc-ip.1",
		"crc-oc-env.1",
		"crc-podman-env.1",
		"crc-setup.1",
		"crc-start.1",
		"crc-status.1",
		"crc-stop.1",
		"crc-version.1",
		"crc.1",
	}, manPagesFiles)
}
