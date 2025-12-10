package libhvee

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToUnixPath(t *testing.T) {
	assert.Equal(t, "/mnt/c/Users/crc", ConvertToUnixPath("C:\\Users\\crc"))
	assert.Equal(t, "/mnt/d/Users/crc", ConvertToUnixPath("d:\\Users\\crc"))
}
