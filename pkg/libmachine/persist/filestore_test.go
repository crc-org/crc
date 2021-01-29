package persist

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/code-ready/crc/pkg/drivers/none"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/code-ready/crc/pkg/libmachine/version"
	"github.com/stretchr/testify/assert"
)

func getTestStore() (Filestore, func(), error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		return Filestore{}, nil, err
	}
	return Filestore{
			MachinesDir: tmpDir,
		}, func() {
			_ = os.RemoveAll(tmpDir)
		}, nil
}

func TestStoreSave(t *testing.T) {
	store, cleanup, err := getTestStore()
	assert.NoError(t, err)
	defer cleanup()

	h := testHost()

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.MachinesDir, h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}

	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		r := regexp.MustCompile("config.json.tmp*")
		if r.MatchString(f.Name()) {
			t.Fatalf("Failed to remove temp filestore:%s", f.Name())
		}
	}
}

func TestStoreSaveOmitRawDriver(t *testing.T) {
	store, cleanup, err := getTestStore()
	assert.NoError(t, err)
	defer cleanup()

	h := testHost()

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	configJSONPath := filepath.Join(store.MachinesDir, h.Name, "config.json")

	f, err := os.Open(configJSONPath)
	if err != nil {
		t.Fatal(err)
	}

	configData, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	fakeHost := make(map[string]interface{})

	if err := json.Unmarshal(configData, &fakeHost); err != nil {
		t.Fatal(err)
	}

	if rawDriver, ok := fakeHost["RawDriver"]; ok {
		t.Fatal("Should not have gotten a value for RawDriver reading host from disk but got one: ", rawDriver)
	}

}

func TestStoreRemove(t *testing.T) {
	store, cleanup, err := getTestStore()
	assert.NoError(t, err)
	defer cleanup()

	h := testHost()

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.MachinesDir, h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}

	err = store.Remove(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Host path still exists after remove: %s", path)
	}
}

func TestStoreExists(t *testing.T) {
	store, cleanup, err := getTestStore()
	assert.NoError(t, err)
	defer cleanup()

	h := testHost()

	exists, err := store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("Host should not exist before saving")
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	if err := store.SetExists(h.Name); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Host should exist after saving")
	}

	if err := store.Remove(h.Name); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("Host should not exist after removing")
	}
}

func TestStoreLoad(t *testing.T) {
	store, cleanup, err := getTestStore()
	assert.NoError(t, err)
	defer cleanup()

	h := testHost()

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	h, err = store.Load(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	rawDataDriver, ok := h.Driver.(*host.RawDataDriver)
	if !ok {
		t.Fatal("Expected driver loaded from store to be of type *host.RawDataDriver and it was not")
	}

	realDriver := none.NewDriver(h.Name, store.MachinesDir)

	if err := json.Unmarshal(rawDataDriver.Data, &realDriver); err != nil {
		t.Fatalf("Error unmarshaling rawDataDriver data into concrete 'none' driver: %s", err)
	}
}

func testHost() *host.Host {
	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          "test-host",
		Driver:        none.NewDriver("test-host", "/tmp/artifacts"),
		DriverName:    "none",
	}
}
