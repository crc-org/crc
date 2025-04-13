package go9p

import (
	"fmt"
	"os"
	"syscall"
)

func postfd(name string) (pipe *os.File, srvHandle *os.File, err error) {
	f1, f2, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	sf, err := os.OpenFile("/srv/"+name, os.O_CREATE|os.O_EXCL|os.O_WRONLY|64|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return nil, nil, err
	}

	_, err = sf.Write([]byte(fmt.Sprintf("%d", f2.Fd())))
	if err != nil {
		sf.Close()
		return nil, nil, err
	}
	f2.Close()
	return f1, sf, nil
}
