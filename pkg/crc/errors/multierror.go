package errors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/crc-org/crc/pkg/crc/logging"
)

type vmNotExist string

func (v vmNotExist) Error() string {
	return string(v)
}

const VMNotExist vmNotExist = "Machine does not exist. Use 'crc start' to create it"

type daemonNotRunning string

func (d daemonNotRunning) Error() string {
	return string(d)
}

const DaemonNotRunning daemonNotRunning = "crc does not seem to be setup correctly, have you run 'crc setup'?"

type PreflightError struct {
	Err error
}

func (p *PreflightError) Error() string {
	return p.Err.Error()
}

func (p *PreflightError) Unwrap() error {
	return p.Err
}

type MultiError struct {
	Errors []error
}

func (m MultiError) Error() string {
	if len(m.Errors) == 0 {
		return ""
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	var aggregatedErrors []string

	count := 1
	current := m.Errors[0].Error()
	for i := 1; i < len(m.Errors); i++ {
		if m.Errors[i].Error() == current {
			count++
			continue
		}
		aggregatedErrors = append(aggregatedErrors, errorWithCount(current, count))
		count = 1
		current = m.Errors[i].Error()
	}
	aggregatedErrors = append(aggregatedErrors, errorWithCount(current, count))

	return strings.Join(aggregatedErrors, "\n")
}

func (m *MultiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

func errorWithCount(current string, count int) string {
	if count == 1 {
		return current
	}
	return fmt.Sprintf("%s (x%d)", current, count)
}

// RetriableError is an error that can be tried again
type RetriableError struct {
	Err error
}

func (r *RetriableError) Error() string {
	return "Temporary error: " + r.Err.Error()
}

// Retry retries for a certain duration, after a delay
func Retry(ctx context.Context, limit time.Duration, callback func() error, d time.Duration) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	m := MultiError{}
	timeLimit := time.Now().Add(limit)
	attempt := 0
	for time.Now().Before(timeLimit) || attempt < 2 {
		logging.Debugf("retry loop: attempt %d", attempt)
		err := callback()
		if err == nil {
			return nil
		}
		attempt++
		m.Collect(err)
		if _, ok := err.(*RetriableError); !ok {
			logging.Debugf("non-retriable error: %v", err)
			return m
		}
		logging.Debugf("error: %v - sleeping %s", err, d)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}
	}
	logging.Debugf("RetryAfter timeout after %d tries", attempt)
	return m
}
