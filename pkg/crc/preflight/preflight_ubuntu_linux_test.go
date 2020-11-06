// +build linux

package preflight

import (
	"errors"
	"os"
	"testing"

	"github.com/code-ready/crc/pkg/crc/constants"

	"github.com/stretchr/testify/assert"
)

var (
	source = `#include <tunables/global>
profile LIBVIRT_TEMPLATE flags=(attach_disconnected) {
  #include <abstractions/libvirt-qemu>
}`
	expected = `#include <tunables/global>
profile LIBVIRT_TEMPLATE flags=(attach_disconnected) {
  ` + constants.MachineCacheDir + `/*/crc.qcow2 rk,

  #include <abstractions/libvirt-qemu>
}`
)

func TestCheckAppArmor(t *testing.T) {
	assert.EqualError(t, checkAppArmorExceptionIsPresent(mockReader(source))(), "AppArmor profile not configured")
	assert.NoError(t, checkAppArmorExceptionIsPresent(mockReader(expected))())
}

func TestFixAppArmor(t *testing.T) {
	assert.NoError(t, addAppArmorExceptionForQcowDisks(mockReader(source), writerVerifier(expected))())
	assert.EqualError(t, addAppArmorExceptionForQcowDisks(mockReader("invalid"), writerVerifier(expected))(),
		"unexpected AppArmor template file /etc/apparmor.d/libvirt/TEMPLATE.qemu, cannot configure it automatically")
}

func TestCleanupAppArmor(t *testing.T) {
	assert.NoError(t, removeAppArmorExceptionForQcowDisks(mockReader(expected), writerVerifier(source))())
}

func writerVerifier(expected string) func(reason string, content string, filepath string, mode os.FileMode) error {
	return func(reason, content, filepath string, mode os.FileMode) error {
		if content == expected {
			return nil
		}
		return errors.New("unexpected content")
	}
}

func mockReader(input string) func(string) ([]byte, error) {
	return func(string) ([]byte, error) {
		return []byte(input), nil
	}
}
