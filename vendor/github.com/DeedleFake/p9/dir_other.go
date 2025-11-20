//go:build !linux && !darwin && !plan9 && !windows
// +build !linux,!darwin,!plan9,!windows

package p9

import "os"

func infoToEntry(fi os.FileInfo) DirEntry {
	return DirEntry{
		FileMode:  ModeFromOS(fi.Mode()),
		MTime:     fi.ModTime(),
		Length:    uint64(fi.Size()),
		EntryName: fi.Name(),
	}
}
