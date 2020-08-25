package oc

import (
	"testing"

	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	ocConfig := Config{
		Runner:         crcos.NewLocalCommandRunner(),
		OcBinaryPath:   "/bin/echo",
		KubeconfigPath: "kubeconfig-file",
		Context:        "a-context",
		Cluster:        "a-cluster",
	}
	stdout, _, err := ocConfig.RunOcCommand("a-command")
	assert.NoError(t, err)
	assert.Equal(t, "a-command --context a-context --cluster a-cluster --kubeconfig kubeconfig-file\n", stdout)
}

func TestRunCommandWithoutContextAndCluster(t *testing.T) {
	ocConfig := Config{
		Runner:         crcos.NewLocalCommandRunner(),
		OcBinaryPath:   "/bin/echo",
		KubeconfigPath: "kubeconfig-file",
	}
	stdout, _, err := ocConfig.RunOcCommand("a-command")
	assert.NoError(t, err)
	assert.Equal(t, "a-command --kubeconfig kubeconfig-file\n", stdout)
}
