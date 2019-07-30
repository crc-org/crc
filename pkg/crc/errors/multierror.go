package errors

import (
	"fmt"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type MultiError struct {
	Errors []error
}

func (m *MultiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

func (m MultiError) ToError() error {
	if len(m.Errors) == 0 {
		return nil
	}

	errStrings := []string{}
	for _, err := range m.Errors {
		errStrings = append(errStrings, err.Error())
	}
	return fmt.Errorf(strings.Join(errStrings, "\n"))
}

// RetriableError is an error that can be tried again
type RetriableError struct {
	Err error
}

func (r RetriableError) Error() string {
	return "Temporary Error: " + r.Err.Error()
}

// RetryAfter retries a number of attempts, after a delay
func RetryAfter(attempts int, callback func() error, d time.Duration) (err error) {
	m := MultiError{}
	for i := 0; i < attempts; i++ {
		if i > 0 {
			logging.Debugf("retry loop %d", i)
		}
		err = callback()
		if err == nil {
			return nil
		}
		m.Collect(err)
		if _, ok := err.(*RetriableError); !ok {
			logging.Debugf("non-retriable error: %v", err)
			return m.ToError()
		}
		logging.Debugf("error: %v - sleeping %s", err, d)
		time.Sleep(d)
	}
	return m.ToError()
}
