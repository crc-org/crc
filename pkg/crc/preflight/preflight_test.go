package preflight

import (
	"errors"
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/stretchr/testify/assert"
)

func TestCheckPreflight(t *testing.T) {
	check, calls := sampleCheck(nil, nil)
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})

	assert.NoError(t, doPreflightChecks(cfg, []Check{*check}))
	assert.True(t, calls.checked)
	assert.False(t, calls.fixed)
}

func TestSkipPreflight(t *testing.T) {
	check, calls := sampleCheck(nil, nil)
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})
	_, err := cfg.Set("skip-sample", true)
	assert.NoError(t, err)

	assert.NoError(t, doPreflightChecks(cfg, []Check{*check}))
	assert.False(t, calls.checked)
}

func TestFixPreflight(t *testing.T) {
	check, calls := sampleCheck(errors.New("check failed"), nil)
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})

	assert.NoError(t, doFixPreflightChecks(cfg, []Check{*check}))
	assert.True(t, calls.checked)
	assert.True(t, calls.fixed)
}

func TestWarnPreflight(t *testing.T) {
	check, calls := sampleCheck(errors.New("check failed"), errors.New("fix failed"))
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})
	_, err := cfg.Set("warn-sample", true)
	assert.NoError(t, err)

	assert.NoError(t, doFixPreflightChecks(cfg, []Check{*check}))
	assert.True(t, calls.checked)
	assert.True(t, calls.fixed)
}

func sampleCheck(checkErr, fixErr error) (*Check, *status) {
	status := &status{}
	return &Check{
		configKeySuffix:  "sample",
		checkDescription: "Sample check",
		check: func() error {
			status.checked = true
			return checkErr
		},
		fixDescription: "sample fix",
		fix: func() error {
			status.fixed = true
			return fixErr
		},
	}, status
}

type status struct {
	checked, fixed bool
}
