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

	createDummyBundleContent(t, dir, "crc_libvirt_4.6.1", "1.0")

	repo := &Repository{
		CacheDir: dir,
		OcBinDir: ocBinDir,
	}

	bundle, err := repo.Use("crc_libvirt_4.6.1.crcbundle")
	assert.NoError(t, err)
	assert.Equal(t, "4.6.1", bundle.ClusterInfo.OpenShiftVersion.String())

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

	createDummyBundleContent(t, dir, "crc_libvirt_4.6.1", "1.0")

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
	assert.Equal(t, "4.6.1", bundle.GetOpenshiftVersion())

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
	createDummyBundleContent(t, dir, "crc_libvirt_4.6.1", "0.9")
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.EqualError(t, err, "cannot use bundle with version 0.9, bundle version must satisfy ^1.0 constraint")
	os.RemoveAll(bundlePath)

	createDummyBundleContent(t, dir, "crc_libvirt_4.6.1", "1.1")
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.NoError(t, err)
	os.RemoveAll(bundlePath)

	createDummyBundleContent(t, dir, "crc_libvirt_4.6.1", "2.0")
	_, err = repo.Get("crc_libvirt_4.6.1.crcbundle")
	assert.EqualError(t, err, "cannot use bundle with version 2.0, bundle version must satisfy ^1.0 constraint")
	os.RemoveAll(bundlePath)
}

func TestListBundles(t *testing.T) {
	dir, err := ioutil.TempDir("", "repo")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ocBinDir, err := ioutil.TempDir("", "oc-bin-dir")
	assert.NoError(t, err)
	defer os.RemoveAll(ocBinDir)

	createDummyBundleContent(t, dir, "crc_libvirt_4.6.15", "1.0")
	createDummyBundleContent(t, dir, "crc_libvirt_4.7.0", "1.0")
	createDummyBundleContent(t, dir, "crc_libvirt_4.7.11_1621489482", "1.0")
	createDummyBundleContent(t, dir, "crc_libvirt_4.8.0-0.nightly-2021-05-26-021757", "1.0")
	createDummyBundleContent(t, dir, "crc_libvirt_4.10.0", "1.0")

	repo := &Repository{
		CacheDir: dir,
		OcBinDir: ocBinDir,
	}

	bundles, err := repo.List()
	assert.NoError(t, err)
	var names []string
	for _, bundle := range bundles {
		names = append(names, bundle.GetBundleName())
	}
	assert.Equal(t, []string{
		"crc_libvirt_4.7.11_1621489482",
		"crc_libvirt_4.10.0",
		"crc_libvirt_4.8.0-0.nightly-2021-05-26-021757",
		"crc_libvirt_4.7.0",
		"crc_libvirt_4.6.15",
	}, names)
}

func createDummyBundleContent(t *testing.T, dir, name, version string) {
	bundleDir := filepath.Join(dir, name)
	assert.NoError(t, os.MkdirAll(bundleDir, 0755))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundleDir, metadataFilename), []byte(jsonForBundleWithVersion(version, name)), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundleDir, constants.OcExecutableName), []byte("openshift-client"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundleDir, "kubeadmin-password"), []byte("kubeadmin-password"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundleDir, "kubeconfig"), []byte("kubeconfig"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundleDir, "id_ecdsa_crc"), []byte("id_ecdsa_crc"), 0600))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(bundleDir, "crc.qcow2"), []byte("crc.qcow2"), 0600))
}
