package p9

import (
	"errors"
	"os"
	"syscall"
	"time"
)

func infoToEntry(fi os.FileInfo) DirEntry {
	sys, _ := fi.Sys().(*syscall.Dir)
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
		ATime:     time.Unix(int64(sys.Atime), 0),
		MTime:     fi.ModTime(),
		Length:    uint64(fi.Size()),
		EntryName: fi.Name(),
		UID:       sys.Uid,
		GID:       sys.Gid,
		MUID:      sys.Muid,
	}
}

func (d Dir) GetQID(p string) (QID, error) {
	fi, err := os.Stat(d.path(p))
	if err != nil {
		return QID{}, err
	}

	sys, _ := fi.Sys().(*syscall.Dir)
	if sys == nil {
		return QID{}, errors.New("failed to get QID: FileInfo was not Dir")
	}

	return QID{
		Type:    sys.Qid.Type,
		Version: sys.Qid.Vers,
		Path:    sys.Qid.Path,
	}, nil
}
