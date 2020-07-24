package machine

import "github.com/code-ready/crc/pkg/crc/errors"

func startError(name string, description string, err error) (StartResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return StartResult{
		Name:  name,
		Error: fullErr.Error(),
	}, fullErr
}
