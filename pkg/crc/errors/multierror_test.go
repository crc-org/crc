package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryAfter(t *testing.T) {
	calls := 0
	ret := RetryAfter(10, func() error {
		calls++
		return nil
	}, 0)
	assert.NoError(t, ret)
	assert.Equal(t, 1, calls)
}

func TestRetryAfterFailure(t *testing.T) {
	calls := 0
	ret := RetryAfter(10, func() error {
		calls++
		return errors.New("failed")
	}, 0)
	assert.EqualError(t, ret, "failed")
	assert.Equal(t, 1, calls)
}

func TestRetryAfterMaxAttempts(t *testing.T) {
	calls := 0
	ret := RetryAfter(3, func() error {
		calls++
		return &RetriableError{Err: errors.New("failed")}
	}, 0)
	assert.EqualError(t, ret, "Temporary error: failed\nTemporary error: failed\nTemporary error: failed")
	assert.Equal(t, 3, calls)
}

func TestRetryAfterSuccessAfterFailures(t *testing.T) {
	calls := 0
	ret := RetryAfter(5, func() error {
		calls++
		if calls < 3 {
			return &RetriableError{Err: errors.New("failed")}
		}
		return nil
	}, 0)
	assert.NoError(t, ret)
	assert.Equal(t, 3, calls)
}
