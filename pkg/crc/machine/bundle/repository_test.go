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

	addBundle(t, dir, "crc_libvirt_4.6.1")

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

func TestUseWithOCFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "repo")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	fd, err := ioutil.TempFile("", "oc-bin-dir")
	assert.NoError(t, err)
	fd.Close()
	defer os.RemoveAll(fd.Name())

	addBundle(t, dir, "crc_libvirt_4.6.1")

	repo := &Repository{
		CacheDir: dir,
		OcBinDir: fd.Name(),
	}

	_, err = repo.Use("crc_libvirt_4.6.1.crcbundle")
	assert.NoError(t, err)
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

func TestVersionCheck(t *testing.T) {
	dir, err := ioutil.TempDir("", "repo")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	repo := &Repository{
		CacheDir: dir,
	}

	assert.NoError(t, os.Mkdir(filepath.Join(dir, "crc_libvirt_4.6.1"), 0755))

	writeMetadata(t, filepath.Join(dir, "crc_libvirt_4.6.1"), `{"version":"0.9"}`)
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.EqualError(t, err, "cannot use bundle with version 0.9, bundle version must satisfy ^1.0 constraint")

	writeMetadata(t, filepath.Join(dir, "crc_libvirt_4.6.1"), `{"version":"1.1"}`)
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.NoError(t, err)

	writeMetadata(t, filepath.Join(dir, "crc_libvirt_4.6.1"), `{"version":"2.0"}`)
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.EqualError(t, err, "cannot use bundle with version 2.0, bundle version must satisfy ^1.0 constraint")
}

func writeMetadata(t *testing.T, dir string, s string) bool {
	return assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, metadataFilename), []byte(s), 0600))
}

func addBundle(t *testing.T, dir, name string) {
	assert.NoError(t, os.Mkdir(filepath.Join(dir, name), 0755))
	writeMetadata(t, filepath.Join(dir, name), reference)
	assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, name, constants.OcExecutableName), []byte("openshift-client"), 0600))
}
