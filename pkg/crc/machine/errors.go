package machine

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func startError(name string, description string, err error) (StartResult, error) {
	fullErr := logErrorf("%s: %v", description, err)
	return StartResult{
		Name:  name,
		Error: fullErr.Error(),
	}, fullErr
}

func statusError(name string, description string, err error) (ClusterStatusResult, error) {
	fullErr := logErrorf("%s: %v", description, err)
	return ClusterStatusResult{
		Name:    name,
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}

func consoleURLError(description string, err error) (ConsoleResult, error) {
	fullErr := logErrorf("%s: %v", description, err)
	return ConsoleResult{
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}

func logErrorf(format string, args ...interface{}) error {
	logging.Errorf(format, args...)
	return fmt.Errorf(format, args...)
}
