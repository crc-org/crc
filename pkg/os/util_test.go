package os

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceEnv(t *testing.T) {
	env := []string{"HOME=/home/user/", "PATH=/bin:/sbin:/usr/bin", "LC_ALL=de_DE.UTF8"}
	replaced := ReplaceOrAddEnv(env, "LC_ALL", "C")

	assert.Len(t, replaced, 3)
	assert.Equal(t, env[2], "LC_ALL=de_DE.UTF8")
	assert.Equal(t, replaced[2], "LC_ALL=C")
}

func TestAddEnv(t *testing.T) {
	env := []string{"HOME=/home/user/", "PATH=/bin:/sbin:/usr/bin", "LC_ALL=de_DE.UTF8"}
	replaced := ReplaceOrAddEnv(env, "KUBECONFIG", "some-data")

	assert.Len(t, replaced, 4)
	assert.Equal(t, replaced[3], "KUBECONFIG=some-data")
}
