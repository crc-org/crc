package os

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunPrivileged(t *testing.T) {
	privilegedCommand = "echo"
	_, _, err := RunPrivileged("it should fail", "i-dont-exist")
	assert.ErrorContains(t, err, "i-dont-exist executable not found")
}
