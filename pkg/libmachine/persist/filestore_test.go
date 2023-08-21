package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/pkg/drivers/none"
	"github.com/crc-org/crc/v2/pkg/libmachine/host"
	"github.com/stretchr/testify/assert"
)

func getTestStore(t *testing.T) Filestore {
	return Filestore{
		MachinesDir: t.TempDir(),
	}
}

func TestStoreSave(t *testing.T) {
	store := getTestStore(t)

	h := testHost()

	assert.NoError(t, store.Save(h))

	path := filepath.Join(store.MachinesDir, h.Name)
	assert.DirExists(t, path)
}

func TestStoreSaveOmitRawDriver(t *testing.T) {
	store := getTestStore(t)

	h := testHost()

	assert.NoError(t, store.Save(h))

	configJSONPath := filepath.Join(store.MachinesDir, h.Name, "config.json")

	configData, err := os.ReadFile(configJSONPath)
	assert.NoError(t, err)

	fakeHost := make(map[string]interface{})

	assert.NoError(t, json.Unmarshal(configData, &fakeHost))

	_, ok := fakeHost["RawDriver"]
	assert.False(t, ok)
}

func TestStoreRemove(t *testing.T) {
	store := getTestStore(t)

	h := testHost()

	assert.NoError(t, store.Save(h))

	path := filepath.Join(store.MachinesDir, h.Name)
	assert.DirExists(t, path)

	err := store.Remove(h.Name)
	assert.NoError(t, err)

	assert.NoDirExists(t, path)
}

func TestStoreExists(t *testing.T) {
	store := getTestStore(t)

	h := testHost()

	exists, err := store.Exists(h.Name)
	assert.NoError(t, err)
	assert.False(t, exists)

	assert.NoError(t, store.Save(h))

	assert.NoError(t, store.SetExists(h.Name))

	exists, err = store.Exists(h.Name)
	assert.NoError(t, err)
	assert.True(t, exists)

	assert.NoError(t, store.Remove(h.Name))

	exists, err = store.Exists(h.Name)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestStoreLoad(t *testing.T) {
	store := getTestStore(t)

	h := testHost()

	assert.NoError(t, store.Save(h))

	var err error
	h, err = store.Load(h.Name)
	assert.NoError(t, err)

	rawDataDriver, ok := h.Driver.(*host.RawDataDriver)
	assert.True(t, ok)

	realDriver := none.NewDriver(h.Name, store.MachinesDir)
	assert.NoError(t, json.Unmarshal(rawDataDriver.Data, &realDriver))
}

func testHost() *host.Host {
	return &host.Host{
		ConfigVersion: host.Version,
		Name:          "test-host",
		Driver:        none.NewDriver("test-host", "/tmp/artifacts"),
		DriverName:    "none",
	}
}
