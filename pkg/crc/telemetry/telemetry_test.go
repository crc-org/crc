package telemetry

import (
	"errors"
	"fmt"
	"os/user"
	"testing"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/stretchr/testify/assert"
)

func TestSetError(t *testing.T) {
	err := errors.New("this is an error string")
	assert.Equal(t, err.Error(), SetError(err))

	user, err := user.Current()
	assert.NoError(t, err)

	err = fmt.Errorf("cannot access storage file '%s/.crc/machines/crc/crc.qcow2' (as uid:64055, gid:129): Permission denied')", constants.GetHomeDir())
	assert.NotEqual(t, err.Error(), SetError(err))
	assert.NotContains(t, SetError(err), constants.GetHomeDir())

	err = fmt.Errorf("user %s may not use sudo", user.Username)
	assert.NotEqual(t, err.Error(), SetError(err))
	assert.NotContains(t, SetError(err), user.Username)
}
