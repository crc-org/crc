package shell

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "")

	shell, err := detect()

	assert.Contains(t, supportedShell, shell)
	assert.NoError(t, err)
}

func TestGetNameAndItsPpidOfCurrent(t *testing.T) {
	pid := os.Getpid()
	if pid < 0 || pid > math.MaxUint32 {
		assert.Fail(t, "integer overflow detected")
	}
	shell, shellppid, err := getNameAndItsPpid(uint32(pid))
	assert.Equal(t, "shell.test.exe", shell)
	ppid := os.Getppid()
	if ppid < 0 || ppid > math.MaxUint32 {
		assert.Fail(t, "integer overflow detected")
	}
	assert.Equal(t, uint32(ppid), shellppid)
	assert.NoError(t, err)
}

func TestGetNameAndItsPpidOfParent(t *testing.T) {
	pid := os.Getppid()
	if pid < 0 || pid > math.MaxUint32 {
		assert.Fail(t, "integer overflow detected")
	}
	shell, _, err := getNameAndItsPpid(uint32(pid))

	assert.Equal(t, "go.exe", shell)
	assert.NoError(t, err)
}
