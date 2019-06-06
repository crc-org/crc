package os

import (
	"strings"

	"github.com/code-ready/crc/pkg/crc/errors"
)

const (
	FEDORA DistroID = "Fedora"
	RHEL   DistroID = "RHEL"
	CENTOS DistroID = "CentOS"
	UBUNTU DistroID = "Ubuntu"
)

var (
	fedoraDistribution LinuxDistribution = LinuxDistribution{
		ID:             FEDORA,
		PackageManager: "dnf",
	}
	elDistribution LinuxDistribution = LinuxDistribution{
		ID:             RHEL,
		PackageManager: "yum",
	}
	ubuntuDistribution LinuxDistribution = LinuxDistribution{
		ID:             UBUNTU,
		PackageManager: "apt-get",
	}
)

type LinuxDistribution struct {
	ID             DistroID
	PackageManager string
}

func CurrentDistribution() (*LinuxDistribution, error) {
	distro, err := getLSBRelease()

	if err != nil {
		return nil, err
	}

	switch distro {
	case FEDORA.String():
		return &fedoraDistribution, nil

	case CENTOS.String():
		fallthrough
	case RHEL.String():
		return &elDistribution, nil

	case UBUNTU.String():
		return &ubuntuDistribution, nil
	}

	return nil, errors.Newf("Unknown distribution: %s", distro)
}

type DistroID string

func (t DistroID) String() string {
	return string(t)
}

func getLSBRelease() (string, error) {
	stdOut, _, err := RunWithDefaultLocale("lsb_release", "-is")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdOut), nil
}
