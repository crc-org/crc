package persist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/libmachine/host"
)

type Filestore struct {
	Path string
}

func NewFilestore(path string) *Filestore {
	return &Filestore{
		Path: path,
	}
}

func (s Filestore) GetMachinesDir() string {
	return filepath.Join(s.Path, "machines")
}

func (s Filestore) saveToFile(data []byte, file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return ioutil.WriteFile(file, data, 0600)
	}

	tmpfi, err := ioutil.TempFile(filepath.Dir(file), "config.json.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfi.Name())

	if err = ioutil.WriteFile(tmpfi.Name(), data, 0600); err != nil {
		return err
	}

	if err = tmpfi.Close(); err != nil {
		return err
	}

	if err = os.Remove(file); err != nil {
		return err
	}

	err = os.Rename(tmpfi.Name(), file)
	return err
}

func (s Filestore) Save(host *host.Host) error {
	data, err := json.MarshalIndent(host, "", "    ")
	if err != nil {
		return err
	}

	hostPath := filepath.Join(s.GetMachinesDir(), host.Name)

	// Ensure that the directory we want to save to exists.
	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return err
	}

	return s.saveToFile(data, filepath.Join(hostPath, "config.json"))
}

func (s Filestore) Remove(name string) error {
	hostPath := filepath.Join(s.GetMachinesDir(), name)
	return os.RemoveAll(hostPath)
}

func (s Filestore) SetExists(name string) error {
	filename := filepath.Join(s.GetMachinesDir(), name, fmt.Sprintf(".%s-exist", name))
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	file.Close()
	log.Debugf("Created %s", filename)

	return nil
}

func (s Filestore) Exists(name string) (bool, error) {
	filename := filepath.Join(s.GetMachinesDir(), name, fmt.Sprintf(".%s-exist", name))
	_, err := os.Stat(filename)
	log.Debugf("Checking file: %s", filename)

	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}

	return false, err
}

func (s Filestore) Load(name string) (*host.Host, error) {
	hostPath := filepath.Join(s.GetMachinesDir(), name)

	if _, err := os.Stat(hostPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("machine %s does not exist", name)
	}
	data, err := ioutil.ReadFile(filepath.Join(s.GetMachinesDir(), name, "config.json"))
	if err != nil {
		return nil, err
	}
	return host.MigrateHost(name, data)
}
