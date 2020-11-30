package bundle

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/code-ready/crc/pkg/crc/constants"

	"github.com/stretchr/testify/assert"
)

func TestUse(t *testing.T) {
	dir, err := ioutil.TempDir("", "repo")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ocBinDir, err := ioutil.TempDir("", "oc-bin-dir")
	assert.NoError(t, err)
	defer os.RemoveAll(ocBinDir)

	assert.NoError(t, os.Mkdir(filepath.Join(dir, "crc_libvirt_4.6.1"), 0755))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, "crc_libvirt_4.6.1", metadataFilename), []byte(reference), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, "crc_libvirt_4.6.1", constants.OcExecutableName), []byte("openshift-client"), 0600))

	repo := &Repository{
		CacheDir: dir,
		OcBinDir: ocBinDir,
	}

	bundle, err := repo.Use("crc_libvirt_4.6.1.crcbundle")
	assert.NoError(t, err)
	assert.Equal(t, "4.6.1", bundle.ClusterInfo.OpenShiftVersion)

	bin, err := ioutil.ReadFile(filepath.Join(ocBinDir, constants.OcExecutableName))
	assert.NoError(t, err)
	assert.Equal(t, "openshift-client", string(bin))
}

func TestExtract(t *testing.T) {
	dir, err := ioutil.TempDir("", "repo")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ocBinDir, err := ioutil.TempDir("", "oc-bin-dir")
	assert.NoError(t, err)
	defer os.RemoveAll(ocBinDir)

	repo := &Repository{
		CacheDir: dir,
		OcBinDir: ocBinDir,
	}

	assert.NoError(t, repo.Extract(filepath.Join("testdata", testBundle(t))))

	bundle, err := repo.Get(testBundle(t))
	assert.NoError(t, err)
	assert.Equal(t, "4.6.1", bundle.ClusterInfo.OpenShiftVersion)

	assert.NoError(t, bundle.Verify())
	_ = os.Remove(bundle.GetKubeConfigPath())
	assert.EqualError(t, bundle.Verify(), "kubeconfig not found in bundle")
}

func testBundle(t *testing.T) string {
	switch runtime.GOOS {
	case "darwin":
		return "crc_hyperkit_4.6.1.crcbundle"
	case "windows":
		return "crc_hyperv_4.6.1.crcbundle"
	case "linux":
		return "crc_libvirt_4.6.1.crcbundle"
	default:
		t.Fatal("unexpected GOOS")
		return ""
	}
}
