package cluster

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/stretchr/testify/assert"
)

func TestGetClusterOperatorStatus(t *testing.T) {
	status, err := GetClusterOperatorStatus(oc.OcConfig{
		Runner: &mockRunner{file: filepath.Join("testdata", "co.json")},
	})
	assert.NoError(t, err)
	assert.Equal(t, &ClusterStatus{
		Available: true,
	}, status)
}

func TestGetClusterOperatorStatusProgressing(t *testing.T) {
	status, err := GetClusterOperatorStatus(oc.OcConfig{
		Runner: &mockRunner{file: filepath.Join("testdata", "co-progressing.json")},
	})
	assert.NoError(t, err)
	assert.Equal(t, &ClusterStatus{
		Available:   true,
		Progressing: true,
	}, status)
}

type mockRunner struct {
	file string
}

func (r *mockRunner) Run(args ...string) (string, string, error) {
	bin, err := ioutil.ReadFile(r.file)
	return string(bin), "", err
}

func (r *mockRunner) RunPrivate(args ...string) (string, string, error) {
	return "", "", errors.New("not implemented")
}

func (r *mockRunner) GetKubeconfigPath() string {
	return ""
}
