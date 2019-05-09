package config

import (
	validations "github.com/code-ready/crc/pkg/crc/config"
)

var (
	// Preflight checks
	SkipCheckResolverFilePermissions = createSetting("skip-check-resolver-file-permissions", nil, []validationFnType{validations.ValidateBool})
	WarnCheckResolverFilePermissions = createSetting("warn-check-resolver-file-permissions", nil, []validationFnType{validations.ValidateBool})
)
