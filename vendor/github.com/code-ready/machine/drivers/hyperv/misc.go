package hyperv

import (
	"github.com/code-ready/machine/libmachine/mcnutils"
)

// CRCDiskCopier describes the interactions with crc disk image.
type CRCDiskCopier interface {
	CopyDiskToMachineDir(storePath, machineName, isoURL string) error
}

func NewCRCDiskCopier() CRCDiskCopier {
	return &crcDiskUtilsCopier{}
}

type crcDiskUtilsCopier struct{}

func (u *crcDiskUtilsCopier) CopyDiskToMachineDir(storePath, machineName, isoURL string) error {
	return mcnutils.NewB2dUtils(storePath, "crc.vhdx").CopyDiskToMachineDir(isoURL, machineName)
}
