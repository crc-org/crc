package ufs

import (
	"os"

	p9p "github.com/docker/go-p9p"
)

func oflags(mode p9p.Flag) int {
	flags := 0

	switch mode & 3 {
	case p9p.OREAD:
		flags = os.O_RDONLY
		break

	case p9p.ORDWR:
		flags = os.O_RDWR
		break

	case p9p.OWRITE:
		flags = os.O_WRONLY
		break

	case p9p.OEXEC:
		flags = os.O_RDONLY
		break
	}

	if mode&p9p.OTRUNC != 0 {
		flags |= os.O_TRUNC
	}

	return flags
}
