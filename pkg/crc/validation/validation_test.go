package validation

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEnoughFreeSpace(t *testing.T) {
	dir, err := os.UserHomeDir()
	assert.NoError(t, err)

	// Should pass: any real filesystem has more than 1 byte free
	assert.NoError(t, ValidateEnoughFreeSpace(dir, 1))

	// Should fail: no filesystem has MaxUint64 bytes free
	assert.Error(t, ValidateEnoughFreeSpace(dir, math.MaxUint64))

	// Non-existent path should return nil (fail open)
	assert.NoError(t, ValidateEnoughFreeSpace("/nonexistent/path/that/does/not/exist", 1))
}
