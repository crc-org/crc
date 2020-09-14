package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupRenderActionPlainSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&setupResult{
		Success: true,
	}, out, ""))
	assert.Equal(t, "Setup is complete, you can now run 'crc start -b $bundlename' to start the OpenShift cluster\n", out.String())
}

func TestSetupRenderActionPlainFailure(t *testing.T) {
	out := new(bytes.Buffer)
	assert.EqualError(t, render(&setupResult{
		Success: false,
		Error:   "broken",
	}, out, ""), "broken")
	assert.Equal(t, "", out.String())
}

func TestSetupRenderActionJSONSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&setupResult{
		Success: true,
	}, out, jsonFormat))
	assert.JSONEq(t, `{"success": true}`, out.String())
}

func TestSetupRenderActionJSONFailure(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&setupResult{
		Success: false,
		Error:   "broken",
	}, out, jsonFormat))
	assert.JSONEq(t, `{"success": false, "error": "broken"}`, out.String())
}
