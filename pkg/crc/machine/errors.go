package machine

import "github.com/code-ready/crc/pkg/crc/errors"

func startError(name string, description string, err error) (StartResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return StartResult{
		Name:  name,
		Error: fullErr.Error(),
	}, fullErr
}

func stopError(name string, description string, err error) (StopResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return StopResult{
		Name:    name,
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}

func powerOffError(name string, description string, err error) (PowerOffResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return PowerOffResult{
		Name:    name,
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}

func deleteError(name string, description string, err error) (DeleteResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return DeleteResult{
		Name:    name,
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}

func ipError(name string, description string, err error) (IPResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return IPResult{
		Name:    name,
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}

func statusError(name string, description string, err error) (ClusterStatusResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return ClusterStatusResult{
		Name:    name,
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}

func consoleURLError(description string, err error) (ConsoleResult, error) {
	fullErr := errors.Newf("%s: %v", description, err)
	return ConsoleResult{
		Success: false,
		Error:   fullErr.Error(),
	}, fullErr
}
