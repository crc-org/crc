package errors

import (
	goerrors "errors"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func New(text string) error {
	logging.Error("Error occured: %s", text)
	return goerrors.New(text)
}

func NewF(text string, args ...interface{}) error {
	logging.ErrorF("Error occured: %s", text, args)
	return fmt.Errorf(text, args)
}
