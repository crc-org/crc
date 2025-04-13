package libauth

import (
	"io"
	"os"
)

func openRPC() (io.ReadWriteCloser, error) {
	return os.OpenFile("/mnt/factotum/rpc", os.O_RDWR, 0)
}

func openCtl() (io.ReadWriteCloser, error) {
	return os.OpenFile("/mnt/factotum/ctl", os.O_RDWR, 0)
}

var factotum = "/boot/factotum"
