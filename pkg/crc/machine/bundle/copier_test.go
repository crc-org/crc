package bundle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/stretchr/testify/assert"
)

func TestGenerateBundle(t *testing.T) {
	var b CrcBundleInfo
	assert.NoError(t, json.Unmarshal([]byte(jsonForBundle("crc_4.7.1")), &b))

	tmpBundleDir := t.TempDir()
	b.cachedPath = filepath.Join(tmpBundleDir, b.Name)

	createDummyBundleFiles(t, &b)

	srcDir := t.TempDir()

	customBundleName := "custom_bundle"
	copier, err := NewCopier(&b, srcDir, customBundleName)
	assert.NoError(t, err)

	assert.NoError(t, copier.CopyKubeConfig())

	assert.NoError(t, copier.CopyPrivateSSHKey(copier.srcBundle.GetSSHKeyPath()))
	assert.NoError(t, copier.CopyFilesFromFileList())

	err = crcos.CopyFileContents(copier.srcBundle.GetDiskImagePath(), copier.copiedBundle.GetDiskImagePath(), 0644)
	assert.NoError(t, err)
	assert.NoError(t, copier.SetDiskImage(copier.copiedBundle.GetDiskImagePath(), "qcow2"))

	assert.NoError(t, copier.GenerateBundle(customBundleName))
	defer os.Remove(fmt.Sprintf("%s%s", customBundleName, bundleExtension))
}

func TestGetType(t *testing.T) {
	type data struct {
		value         string
		expectedValue string
	}
	testdata := []data{
		{value: "snc", expectedValue: "snc_custom"},
		{value: "podman", expectedValue: "podman_custom"},
		{value: "snc_custom", expectedValue: "snc_custom"},
		{value: "podman_custom", expectedValue: "podman_custom"},
	}
	for _, d := range testdata {
		assert.EqualValues(t, d.expectedValue, getType(d.value))
	}
}

func createDummyBundleFiles(t *testing.T, bundle *CrcBundleInfo) {
	assert.NoError(t, os.MkdirAll(bundle.cachedPath, 0750))

	files := []string{
		bundle.GetOcPath(),
		bundle.GetKubeConfigPath(),
		bundle.GetSSHKeyPath(),
		bundle.GetDiskImagePath(),
	}

	for _, file := range files {
		if file == "" {
			continue
		}
		fd, err := os.Create(file)
		fd.Close()
		assert.NoError(t, err)
	}
}
