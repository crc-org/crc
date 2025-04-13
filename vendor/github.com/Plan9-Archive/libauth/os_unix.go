// +build !plan9

package libauth

import (
	"io"
	"os"

	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
)

func openRPC() (io.ReadWriteCloser, error) {
	fsys, err := client.MountService("factotum")
	if err != nil {
		return nil, err
	}

	fid, err := fsys.Open("rpc", plan9.ORDWR)
	if err != nil {
		return nil, err
	}

	return fid, nil
}

func openCtl() (io.ReadWriteCloser, error) {
	fsys, err := client.MountService("factotum")
	if err != nil {
		return nil, err
	}

	fid, err := fsys.Open("ctl", plan9.ORDWR)
	if err != nil {
		return nil, err
	}

	return fid, nil
}

var factotum = os.Getenv("PLAN9") + "/bin/factotum"
