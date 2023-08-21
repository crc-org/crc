package cmd

import (
	"bytes"
	"errors"
	"testing"

	crcErrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/stretchr/testify/assert"
)

func TestSetupRenderActionPlainSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&setupResult{
		Success: true,
	}, out, ""))
	assert.Equal(t, "Your system is correctly setup for using CRC. Use 'crc start' to start the instance\n", out.String())
}

func TestSetupRenderActionPlainFailure(t *testing.T) {
	out := new(bytes.Buffer)
	err := errors.New("broken")
	assert.EqualError(t, render(&setupResult{
		Success: false,
		Error:   crcErrors.ToSerializableError(err),
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
		Error:   crcErrors.ToSerializableError(errors.New("broken")),
	}, out, jsonFormat))
	assert.JSONEq(t, `{"success": false, "error": "broken"}`, out.String())
}
