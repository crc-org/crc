// +build windows
package ufs

import (
	"crypto/rand"
	"encoding/binary"
	"os"
	"syscall"
	"time"

	p9p "github.com/docker/go-p9p"
)

func dirFromInfo(info os.FileInfo) p9p.Dir {
	dir := p9p.Dir{}

	sys, _ := info.Sys().(*syscall.Win32FileAttributeData)
	if sys != nil {
		dir.AccessTime = time.Unix(0, sys.LastAccessTime.Nanoseconds())
	}

	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	dir.Qid.Path = binary.LittleEndian.Uint64(randomBytes)

	dir.Qid.Version = uint32(info.ModTime().UnixNano() / 1000000)

	dir.Name = info.Name()
	dir.Mode = uint32(info.Mode() & 0777)
	dir.Length = uint64(info.Size())
	dir.ModTime = info.ModTime()
	dir.MUID = "none"

	if info.IsDir() {
		dir.Qid.Type |= p9p.QTDIR
		dir.Mode |= p9p.DMDIR
	}

	return dir
}
