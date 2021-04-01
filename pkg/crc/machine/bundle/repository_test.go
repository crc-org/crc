package bundle

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/code-ready/crc/pkg/crc/constants"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
)

func TestUse(t *testing.T) {
	dir, err := ioutil.TempDir("", "repo")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ocBinDir, err := ioutil.TempDir("", "oc-bin-dir")
	assert.NoError(t, err)
	defer os.RemoveAll(ocBinDir)

	createDummyBundleContent(t, filepath.Join(dir, "crc_libvirt_4.6.1"), "")

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

	createDummyBundleContent(t, filepath.Join(dir, "crc_libvirt_4.6.1"), "")

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

	_ = os.Remove(bundle.GetKubeConfigPath())
	bundle, err = repo.Get(testBundle(t))
	assert.Nil(t, bundle)
	assert.EqualError(t, err, "kubeconfig not found in bundle")
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

	bundlePath := filepath.Join(dir, "crc_libvirt_4.6.1")
	createDummyBundleContent(t, bundlePath, "0.9")
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.EqualError(t, err, "cannot use bundle with version 0.9, bundle version must satisfy ^1.0 constraint")
	os.RemoveAll(bundlePath)

	createDummyBundleContent(t, bundlePath, "1.1")
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.NoError(t, err)
	os.RemoveAll(bundlePath)

	createDummyBundleContent(t, bundlePath, "2.0")
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.EqualError(t, err, "cannot use bundle with version 2.0, bundle version must satisfy ^1.0 constraint")
	os.RemoveAll(bundlePath)
}

func writeMetadata(t *testing.T, dir string, bundleInfo *CrcBundleInfo) bool {
	metadataStr, err := json.Marshal(bundleInfo)
	assert.NoError(t, err)
	return assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, metadataFilename), []byte(metadataStr), 0600))
}

func createDummyBundleContent(t *testing.T, bundlePath string, version string) {
	assert.NoError(t, os.Mkdir(bundlePath, 0755))

	var bundleInfo CrcBundleInfo
	assert.NoError(t, copier.Copy(&bundleInfo, &parsedReference))
	bundleInfo.Storage.Files[0].Name = constants.OcExecutableName
	if version != "" {
		bundleInfo.Version = version
	}
	writeMetadata(t, bundlePath, &bundleInfo)

	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundlePath, constants.OcExecutableName), []byte("openshift-client"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundlePath, "kubeadmin-password"), []byte("kubeadmin-password"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundlePath, "kubeconfig"), []byte("kubeconfig"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundlePath, "id_ecdsa_crc"), []byte("id_ecdsa_crc"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundlePath, "crc.qcow2"), []byte("crc.qcow2"), 0600))
}
