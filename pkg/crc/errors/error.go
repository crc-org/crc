package errors

import (
	goerrors "errors"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func New(text string) error {
	logging.Error(text)
	return goerrors.New(text)
}

func Newf(format string, args ...interface{}) error {
	logging.Errorf(format, args...)
	return fmt.Errorf(format, args...)
}
