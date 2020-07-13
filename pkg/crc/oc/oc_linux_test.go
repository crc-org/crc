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
	assert.Equal(t, "a-command --kubeconfig kubeconfig-file --context a-context --cluster a-cluster\n", stdout)
}
