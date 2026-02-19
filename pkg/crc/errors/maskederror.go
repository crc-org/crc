package errors

import (
	"strings"
)

const mask = "*****"

type MaskedSecretError struct {
	Err    error
	Secret string // nolint:gosec
}

func (err *MaskedSecretError) Error() string {
	return strings.ReplaceAll(err.Err.Error(), err.Secret, mask)
}
