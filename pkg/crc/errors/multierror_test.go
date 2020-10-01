package errors

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryAfter(t *testing.T) {
	calls := 0
	ret := RetryAfter(time.Second, func() error {
		calls++
		return nil
	}, 0)
	assert.NoError(t, ret)
	assert.Equal(t, 1, calls)
}

func TestRetryAfterFailure(t *testing.T) {
	calls := 0
	ret := RetryAfter(time.Second, func() error {
		calls++
		return errors.New("failed")
	}, 0)
	assert.EqualError(t, ret, "failed")
	assert.Equal(t, 1, calls)
}

func TestRetryAfterMaxAttempts(t *testing.T) {
	calls := 0
	ret := RetryAfter(10*time.Millisecond, func() error {
		calls++
		return &RetriableError{Err: errors.New("failed")}
	}, 0)
	assert.EqualError(t, ret, fmt.Sprintf("Temporary error: failed (x%d)", calls))
	assert.Greater(t, calls, 5)
}

func TestRetryAfterSuccessAfterFailures(t *testing.T) {
	calls := 0
	ret := RetryAfter(time.Second, func() error {
		calls++
		if calls < 3 {
			return &RetriableError{Err: errors.New("failed")}
		}
		return nil
	}, 0)
	assert.NoError(t, ret)
	assert.Equal(t, 3, calls)
}

func TestMultiErrorString(t *testing.T) {
	assert.Equal(t, "Temporary Error: No Pending CSR (x4)", MultiError{
		Errors: []error{
			errors.New("Temporary Error: No Pending CSR"),
			errors.New("Temporary Error: No Pending CSR"),
			errors.New("Temporary Error: No Pending CSR"),
			errors.New("Temporary Error: No Pending CSR"),
		},
	}.Error())

	assert.Equal(t, "No Pending CSR (x2)\nConnection refused (x2)\nNo Pending CSR", MultiError{
		Errors: []error{
			errors.New("No Pending CSR"),
			errors.New("No Pending CSR"),
			errors.New("Connection refused"),
			errors.New("Connection refused"),
			errors.New("No Pending CSR"),
		},
	}.Error())
}
