package util

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testArguments = map[string]struct {
	commandOutput          string
	expectedParsedOutput   string
	expectedParsedExitCode string
}{
	"When command output line contains additional output along with exit code, then parse output and exit code": {
		"remainderOutputexitCodeOfLastCommandInShell=0", "remainderOutput\n", "0",
	},
	"When command output line contains only exit code, then parse exit code": {
		"exitCodeOfLastCommandInShell=1", "", "1",
	},
}

func TestScanPipeShouldCorrectlyParseOutputNotEndingWithNewLine(t *testing.T) {
	for name, test := range testArguments {
		t.Run(name, func(t *testing.T) {
			// Given
			shell.ConfigureTypeOfShell("bash")
			shell.exitCodeChannel = make(chan string, 2)
			scanner := bufio.NewScanner(strings.NewReader(test.commandOutput))
			b := &bytes.Buffer{}
			// When
			shell.ScanPipe(scanner, b, "stdout")
			// Then
			assert.Equal(t, test.expectedParsedOutput, b.String())
			assert.Equal(t, test.expectedParsedExitCode, <-shell.exitCodeChannel)
		})
	}
}
