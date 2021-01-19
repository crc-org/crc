package mcnutils

import (
	"fmt"
	"time"
)

func WaitForSpecificOrError(f func() (bool, error), maxAttempts int, waitInterval time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		stop, err := f()
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
		time.Sleep(waitInterval)
	}
	return fmt.Errorf("Maximum number of retries (%d) exceeded", maxAttempts)
}

func WaitForSpecific(f func() bool, maxAttempts int, waitInterval time.Duration) error {
	return WaitForSpecificOrError(func() (bool, error) {
		return f(), nil
	}, maxAttempts, waitInterval)
}

func WaitFor(f func() bool) error {
	return WaitForSpecific(f, 60, 3*time.Second)
}
