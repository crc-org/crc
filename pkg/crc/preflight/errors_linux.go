package preflight

import (
	"errors"
	"fmt"
)

type unitStatusErr struct {
	shouldBeRunning bool
	unitName        string
}

func (err *unitStatusErr) Error() string {
	if err.shouldBeRunning {
		return fmt.Sprintf("%s is not running", err.unitName)
	}

	return fmt.Sprintf("%s should not be running", err.unitName)
}

func (err *unitStatusErr) Is(target error) bool {
	var castTarget *unitStatusErr
	if !errors.As(target, &castTarget) {
		return false
	}

	return err.shouldBeRunning == castTarget.shouldBeRunning && err.unitName == castTarget.unitName
}

func unitShouldBeRunningErr(unitName string) error {
	return &unitStatusErr{
		shouldBeRunning: true,
		unitName:        unitName,
	}
}

func unitShouldNotBeRunningErr(unitName string) error {
	return &unitStatusErr{
		shouldBeRunning: false,
		unitName:        unitName,
	}
}
