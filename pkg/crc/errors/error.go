package errors

import (
	goerrors "errors"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func New(text string) error {
	logging.Error(fmt.Sprintf("Error occurred: %s", text))
	return goerrors.New(text)
}

func Newf(text string, args ...interface{}) error {
	logging.Errorf(fmt.Sprintf("Error occurred: %s", text), args...)
	return fmt.Errorf(text, args...)
}
