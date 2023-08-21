package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/machine/fakemachine"
	"github.com/stretchr/testify/assert"
)

func TestPlainDelete(t *testing.T) {
	cacheDir := t.TempDir()

	out := new(bytes.Buffer)
	assert.NoError(t, runDelete(out, fakemachine.NewClient(), true, cacheDir, true, true, ""))
	assert.Equal(t, "Deleted the instance\n", out.String())

	_, err := os.Stat(cacheDir)
	assert.True(t, os.IsNotExist(err))
}

func TestNonForceDelete(t *testing.T) {
	cacheDir := t.TempDir()

	out := new(bytes.Buffer)
	assert.NoError(t, runDelete(out, fakemachine.NewClient(), true, cacheDir, true, false, ""))
	assert.Equal(t, "", out.String())

	_, err := os.Stat(cacheDir)
	assert.NoError(t, err)
}

func TestJSONDelete(t *testing.T) {
	cacheDir := t.TempDir()

	out := new(bytes.Buffer)
	assert.NoError(t, runDelete(out, fakemachine.NewClient(), true, cacheDir, false, true, jsonFormat))
	assert.JSONEq(t, `{"success": true}`, out.String())

	_, err := os.Stat(cacheDir)
	assert.True(t, os.IsNotExist(err))
}
