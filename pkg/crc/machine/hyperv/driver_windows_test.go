package hyperv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToUnixPath(t *testing.T) {
	assert.Equal(t, "/mnt/c/Users/crc", convertToUnixPath("C:\\Users\\crc"))
	assert.Equal(t, "/mnt/d/Users/crc", convertToUnixPath("d:\\Users\\crc"))
}
