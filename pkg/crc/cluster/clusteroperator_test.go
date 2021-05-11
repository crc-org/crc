package cluster

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/stretchr/testify/assert"
)

var (
	available = &Status{
		Available: true,
	}
	progressing = &Status{
		Available:   true,
		Progressing: true,
		progressing: []string{"authentication"},
	}
)

func TestGetClusterOperatorsStatus(t *testing.T) {
	status, err := GetClusterOperatorsStatus(ocConfig("co.json"))
	assert.NoError(t, err)
	assert.Equal(t, available, status)
}

func TestGetClusterOperatorsStatusProgressing(t *testing.T) {
	status, err := GetClusterOperatorsStatus(ocConfig("co-progressing.json"))
	assert.NoError(t, err)
	assert.Equal(t, progressing, status)
}

func ocConfig(s string) oc.Config {
	return oc.Config{
		Runner: &mockRunner{file: filepath.Join("testdata", s)},
	}
}

type mockRunner struct {
	file string
}

func (r *mockRunner) Run(executablePath string, args ...string) (string, string, error) {
	bin, err := ioutil.ReadFile(r.file)
	return string(bin), "", err
}

func (r *mockRunner) RunPrivate(executablePath string, args ...string) (string, string, error) {
	bin, err := ioutil.ReadFile(r.file)
	return string(bin), "", err
}

func (r *mockRunner) RunPrivileged(reason string, args ...string) (string, string, error) {
	bin, err := ioutil.ReadFile(r.file)
	return string(bin), "", err
}

func (r *mockRunner) GetKubeconfigPath() string {
	return ""
}
