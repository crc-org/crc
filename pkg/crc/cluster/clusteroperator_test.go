package cluster

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/code-ready/crc/pkg/os"

	"github.com/stretchr/testify/assert"
)

func TestGetClusterOperatorsStatus(t *testing.T) {
	status, err := GetClusterOperatorsStatus(ocConfig("co.json", "containers.json"), false)
	assert.NoError(t, err)
	assert.Equal(t, &Status{
		Available: true,
	}, status)
}

func TestGetClusterOperatorsStatusProgressing(t *testing.T) {
	status, err := GetClusterOperatorsStatus(ocConfig("co-progressing.json", "containers.json"), false)
	assert.NoError(t, err)
	assert.Equal(t, &Status{
		Available:   true,
		Progressing: true,
		progressing: []string{"authentication"},
	}, status)
}

func TestGetClusterOperatorsStatusWithoutContainersRunning(t *testing.T) {
	status, err := GetClusterOperatorsStatus(ocConfig("co.json", "containers-missing.json"), false)
	assert.NoError(t, err)
	assert.Equal(t, &Status{
		Available:   false,
		unavailable: []string{"openshift-apiserver"},
	}, status)
}

func ocConfig(operators, containers string) os.CommandRunner {
	return &mockRunner{
		files: map[string]string{
			"timeout 5s oc get co -ojson --context admin --cluster crc --kubeconfig /opt/kubeconfig": filepath.Join("testdata", operators),
			"timeout 5s crictl ps -o json": filepath.Join("testdata", containers),
		},
	}
}

type mockRunner struct {
	files map[string]string
}

func (r *mockRunner) Run(executablePath string, args ...string) (string, string, error) {
	cmd := strings.Join(append([]string{executablePath}, args...), " ")
	filename, ok := r.files[cmd]
	if !ok {
		return "", "", fmt.Errorf("unexpected command %s", cmd)
	}
	bin, err := ioutil.ReadFile(filename)
	return string(bin), "", err
}

func (r *mockRunner) RunPrivate(executablePath string, args ...string) (string, string, error) {
	cmd := strings.Join(append([]string{executablePath}, args...), " ")
	filename, ok := r.files[cmd]
	if !ok {
		return "", "", fmt.Errorf("unexpected command %s", cmd)
	}
	bin, err := ioutil.ReadFile(filename)
	return string(bin), "", err
}

func (r *mockRunner) RunPrivileged(reason string, args ...string) (string, string, error) {
	cmd := strings.Join(args, " ")
	filename, ok := r.files[cmd]
	if !ok {
		return "", "", fmt.Errorf("unexpected command %s", cmd)
	}
	bin, err := ioutil.ReadFile(filename)
	return string(bin), "", err
}

func (r *mockRunner) GetKubeconfigPath() string {
	return ""
}
