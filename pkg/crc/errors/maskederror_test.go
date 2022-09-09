package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskedError(t *testing.T) {
	err := errors.New("The password is: pass@122")
	maskedErr := &MaskedSecretError{err, "pass@122"}
	expectedErrMsg := fmt.Sprintf("The password is: %s", mask)
	assert.EqualError(t, maskedErr, expectedErrMsg)
}
