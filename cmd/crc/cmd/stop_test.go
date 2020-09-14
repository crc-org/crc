package cmd

import (
	"bytes"
	"testing"

	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/stretchr/testify/assert"
)

func TestStopPlainSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runStop(out, fakemachine.NewClient(), true, false, ""))
	assert.Equal(t, "Stopped the OpenShift cluster\n", out.String())
}

func TestStopPlainError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.EqualError(t, runStop(out, fakemachine.NewFailingClient(), true, false, ""), "stop failed")
}

func TestStopWithForcePlainError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.EqualError(t, runStop(out, fakemachine.NewFailingClient(), true, true, ""), "poweroff failed")
}

func TestStopJSONSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runStop(out, fakemachine.NewClient(), false, false, jsonFormat))
	assert.JSONEq(t, `{"success": true, "forced": false}`, out.String())
}

func TestStopJSONError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runStop(out, fakemachine.NewFailingClient(), false, false, jsonFormat))
	assert.JSONEq(t, `{"success": false, "forced": false, "error": "stop failed"}`, out.String())
}

func TestStopWithForceJSONError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runStop(out, fakemachine.NewFailingClient(), false, true, jsonFormat))
	assert.JSONEq(t, `{"success": false, "forced": true, "error": "poweroff failed"}`, out.String())
}
