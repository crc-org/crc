package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainVersion(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runPrintVersion(out, &version{
		Version:             "1.13",
		Commit:              "aabbcc",
		OpenshiftVersion:    "4.5.4",
		PodmanVersion:       "3.4.4",
		InstalledBundlePath: "",
	}, ""))
	assert.Equal(t, `CodeReady Containers version: 1.13+aabbcc
OpenShift version: 4.5.4 (not embedded in executable)
Podman version: 3.4.4
`, out.String())
}

func TestJsonVersion(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runPrintVersion(out, &version{
		Version:             "1.13",
		Commit:              "aabbcc",
		OpenshiftVersion:    "4.5.4",
		PodmanVersion:       "3.4.4",
		InstalledBundlePath: "",
	}, "json"))

	expected := `{"version": "1.13", "commit": "aabbcc", "openshiftVersion": "4.5.4", "podmanVersion": "3.4.4"}`
	assert.JSONEq(t, expected, out.String())
}
