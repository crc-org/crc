package oc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	ocConfig := OcConfig{
		Runner: OcLocalRunner{
			OcBinaryPath:   "/bin/echo",
			KubeconfigPath: "kubeconfig-file",
		},
		Context: "a-context",
		Cluster: "a-cluster",
	}
	stdout, _, err := ocConfig.RunOcCommand("a-command")
	assert.NoError(t, err)
	assert.Equal(t, "a-command --context a-context --cluster a-cluster --kubeconfig kubeconfig-file\n", stdout)
}

func TestEnvRunner(t *testing.T) {
	runner := OcEnvRunner{
		OcBinaryPath:  "printenv",
		KubeconfigEnv: "some-env",
	}
	stdout, _, err := runner.Run()
	assert.NoError(t, err)
	assert.Contains(t, stdout, "KUBECONFIG=some-env\n")
}

func TestRunCommandWithoutContextAndCluster(t *testing.T) {
	ocConfig := OcConfig{
		Runner: OcLocalRunner{
			OcBinaryPath:   "/bin/echo",
			KubeconfigPath: "kubeconfig-file",
		},
	}
	stdout, _, err := ocConfig.RunOcCommand("a-command")
	assert.NoError(t, err)
	assert.Equal(t, "a-command --kubeconfig kubeconfig-file\n", stdout)
}
