package bundle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/stretchr/testify/assert"
)

func TestGenerateBundle(t *testing.T) {
	var b CrcBundleInfo
	assert.NoError(t, json.Unmarshal([]byte(jsonForBundle("crc_4.7.1")), &b))

	tmpBundleDir, err := ioutil.TempDir("", "bundle_data")
	assert.NoError(t, err)
	b.cachedPath = filepath.Join(tmpBundleDir, b.Name)
	defer os.RemoveAll(tmpBundleDir)

	createDummyBundleFiles(t, &b)

	srcDir, err := ioutil.TempDir("", "testdata")
	assert.NoError(t, err)
	defer os.RemoveAll(srcDir)

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

func createDummyBundleFiles(t *testing.T, bundle *CrcBundleInfo) {
	assert.NoError(t, os.MkdirAll(bundle.cachedPath, 0750))

	files := []string{
		bundle.GetOcPath(),
		bundle.GetKubeConfigPath(),
		bundle.GetSSHKeyPath(),
		bundle.GetDiskImagePath(),
		bundle.GetKernelPath(),
		bundle.GetInitramfsPath(),
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
