package p9

import (
	"os"
	"syscall"
	"time"
)

func infoToEntry(fi os.FileInfo) DirEntry {
	sys, _ := fi.Sys().(*syscall.Win32FileAttributeData)
	if sys == nil {
		return DirEntry{
			FileMode:  ModeFromOS(fi.Mode()),
			MTime:     fi.ModTime(),
			Length:    uint64(fi.Size()),
			EntryName: fi.Name(),
		}
	}

	return DirEntry{
		FileMode:  ModeFromOS(fi.Mode()),
		ATime:     time.Unix(0, sys.LastAccessTime.Nanoseconds()),
		MTime:     fi.ModTime(),
		Length:    uint64(fi.Size()),
		EntryName: fi.Name(),
	}
}
