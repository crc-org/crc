//go:build !windows
// +build !windows

package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBash(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "/bin/bash")

	shell, err := detect()

	assert.Equal(t, "bash", shell)
	assert.NoError(t, err)
}

func TestDetectFish(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "/bin/fish")

	shell, err := detect()

	assert.Equal(t, "fish", shell)
	assert.NoError(t, err)
}
