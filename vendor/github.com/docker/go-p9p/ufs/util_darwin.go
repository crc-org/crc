// +build darwin
package ufs

import (
	"os"
	"syscall"
	"time"

	p9p "github.com/docker/go-p9p"
)

func dirFromInfo(info os.FileInfo) p9p.Dir {
	dir := p9p.Dir{}

	dir.Qid.Path = info.Sys().(*syscall.Stat_t).Ino
	dir.Qid.Version = uint32(info.ModTime().UnixNano() / 1000000)

	dir.Name = info.Name()
	dir.Mode = uint32(info.Mode() & 0777)
	dir.Length = uint64(info.Size())
	dir.AccessTime = atime(info.Sys().(*syscall.Stat_t))
	dir.ModTime = info.ModTime()
	dir.MUID = "none"

	if info.IsDir() {
		dir.Qid.Type |= p9p.QTDIR
		dir.Mode |= p9p.DMDIR
	}

	return dir
}

func atime(stat *syscall.Stat_t) time.Time {
	return time.Unix(stat.Atimespec.Unix())
}
