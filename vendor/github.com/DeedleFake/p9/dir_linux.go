package p9

import (
	"errors"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

func infoToEntry(fi os.FileInfo) DirEntry {
	sys, _ := fi.Sys().(*syscall.Stat_t)
	if sys == nil {
		return DirEntry{
			FileMode:  ModeFromOS(fi.Mode()),
			MTime:     fi.ModTime(),
			Length:    uint64(fi.Size()),
			EntryName: fi.Name(),
		}
	}

	var uname string
	uid, err := user.LookupId(strconv.FormatUint(uint64(sys.Uid), 10))
	if err == nil {
		uname = uid.Username
	}

	var gname string
	gid, err := user.LookupGroupId(strconv.FormatUint(uint64(sys.Gid), 10))
	if err == nil {
		gname = gid.Name
	}

	return DirEntry{
		FileMode:  ModeFromOS(fi.Mode()),
		ATime:     time.Unix(sys.Atim.Unix()),
		MTime:     fi.ModTime(),
		Length:    uint64(fi.Size()),
		EntryName: fi.Name(),
		UID:       uname,
		GID:       gname,
	}
}

func (d Dir) GetQID(p string) (QID, error) {
	fi, err := os.Stat(d.path(p))
	if err != nil {
		return QID{}, err
	}

	sys, _ := fi.Sys().(*syscall.Stat_t)
	if sys == nil {
		return QID{}, errors.New("failed to get QID: FileInfo was not Stat_t")
	}

	return QID{
		Type:    ModeFromOS(fi.Mode()).QIDType(),
		Version: 0,
		Path:    sys.Ino,
	}, nil
}
