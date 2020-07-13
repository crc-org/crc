package oc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	ocConfig := NewOcConfig(OcLocalRunner{
		OcBinaryPath:   "/bin/echo",
		KubeconfigPath: "kubeconfig-file",
	}, "a-context", "a-cluster")
	stdout, _, _ := ocConfig.RunOcCommand("a-command")
	assert.Equal(t, "a-command --context a-context --cluster a-cluster --kubeconfig kubeconfig-file\n", stdout)
}
