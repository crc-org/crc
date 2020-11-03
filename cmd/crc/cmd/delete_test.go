package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlainDelete(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	out := new(bytes.Buffer)
	assert.NoError(t, runDelete(out, fakemachine.NewClient(), true, cacheDir, true, true, ""))
	assert.Equal(t, "Deleted the OpenShift cluster\n", out.String())

	_, err = os.Stat(cacheDir)
	assert.True(t, os.IsNotExist(err))
}

func TestNonForceDelete(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	out := new(bytes.Buffer)
	assert.NoError(t, runDelete(out, fakemachine.NewClient(), true, cacheDir, true, false, ""))
	assert.Equal(t, "", out.String())

	_, err = os.Stat(cacheDir)
	assert.NoError(t, err)
}

func TestJSONDelete(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	out := new(bytes.Buffer)
	assert.NoError(t, runDelete(out, fakemachine.NewClient(), true, cacheDir, false, true, jsonFormat))
	assert.JSONEq(t, `{"success": true}`, out.String())

	_, err = os.Stat(cacheDir)
	assert.True(t, os.IsNotExist(err))
}
