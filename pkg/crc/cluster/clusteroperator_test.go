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

func TestGetClusterOperatorStatus(t *testing.T) {
	status, err := GetClusterOperatorStatus(ocConfig("co.json"), "authentication")
	assert.NoError(t, err)
	assert.Equal(t, available, status)

	status, err = GetClusterOperatorStatus(ocConfig("co-progressing.json"), "authentication")
	assert.NoError(t, err)
	assert.Equal(t, progressing, status)

	status, err = GetClusterOperatorStatus(ocConfig("co-progressing.json"), "cloud-credential")
	assert.NoError(t, err)
	assert.Equal(t, available, status)
}

func TestGetClusterOperatorStatusNotFound(t *testing.T) {
	_, err := GetClusterOperatorStatus(ocConfig("co-progressing.json"), "not-found")
	assert.EqualError(t, err, "no cluster operator found")
}

func ocConfig(s string) oc.Config {
	return oc.Config{
		Runner: &mockRunner{file: filepath.Join("testdata", s)},
	}
}

type mockRunner struct {
	file string
}

func (r *mockRunner) Run(binaryPath string, args ...string) (string, string, error) {
	bin, err := ioutil.ReadFile(r.file)
	return string(bin), "", err
}

func (r *mockRunner) RunPrivate(binaryPath string, args ...string) (string, string, error) {
	bin, err := ioutil.ReadFile(r.file)
	return string(bin), "", err
}

func (r *mockRunner) GetKubeconfigPath() string {
	return ""
}
