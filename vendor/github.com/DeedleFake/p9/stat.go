package p9

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
	"unsafe"

	"github.com/DeedleFake/p9/internal/util"
	"github.com/DeedleFake/p9/proto"
)

var (
	// ErrLargeStat is returned during decoding when a stat is larger
	// than its own declared size.
	ErrLargeStat = errors.New("stat larger than declared size")
)

// FileMode stores permission and type information about a file or
// directory.
type FileMode uint32

// FileMode type bitmasks.
const (
	ModeDir FileMode = 1 << (31 - iota)
	ModeAppend
	ModeExclusive
	ModeMount
	ModeAuth
	ModeTemporary
	ModeSymlink
	_
	ModeDevice
	ModeNamedPipe
	ModeSocket
	ModeSetuid
	ModeSetgid
)

// ModeFromOS converts an os.FileMode to a FileMode.
func ModeFromOS(m os.FileMode) FileMode {
	r := FileMode(m.Perm())

	if m&os.ModeDir != 0 {
		r |= ModeDir
	}
	if m&os.ModeAppend != 0 {
		r |= ModeAppend
	}
	if m&os.ModeExclusive != 0 {
		r |= ModeExclusive
	}
	if m&os.ModeTemporary != 0 {
		r |= ModeTemporary
	}
	if m&os.ModeSymlink != 0 {
		r |= ModeSymlink
	}
	if m&os.ModeDevice != 0 {
		r |= ModeDevice
	}
	if m&os.ModeNamedPipe != 0 {
		r |= ModeNamedPipe
	}
	if m&os.ModeSocket != 0 {
		r |= ModeSocket
	}
	if m&os.ModeSetuid != 0 {
		r |= ModeSetuid
	}
	if m&os.ModeSetgid != 0 {
		r |= ModeSetgid
	}

	return r
}

// OS converts a FileMode to an os.FileMode.
func (m FileMode) OS() os.FileMode {
	r := os.FileMode(m.Perm())

	if m&ModeDir != 0 {
		r |= os.ModeDir
	}
	if m&ModeAppend != 0 {
		r |= os.ModeAppend
	}
	if m&ModeExclusive != 0 {
		r |= os.ModeExclusive
	}
	if m&ModeTemporary != 0 {
		r |= os.ModeTemporary
	}
	if m&ModeSymlink != 0 {
		r |= os.ModeSymlink
	}
	if m&ModeDevice != 0 {
		r |= os.ModeDevice
	}
	if m&ModeNamedPipe != 0 {
		r |= os.ModeNamedPipe
	}
	if m&ModeSocket != 0 {
		r |= os.ModeSocket
	}
	if m&ModeSetuid != 0 {
		r |= os.ModeSetuid
	}
	if m&ModeSetgid != 0 {
		r |= os.ModeSetgid
	}

	return r
}

// QIDType converts a FileMode to a QIDType. Note that this will
// result in a loss of information, as the information stored by
// QIDType is a direct subset of that handled by FileMode.
func (m FileMode) QIDType() QIDType {
	return QIDType(m >> 24)
}

// Type returns a FileMode containing only the type bits of m.
func (m FileMode) Type() FileMode {
	return m & 0xFFFF0000
}

// Perm returns a FileMode containing only the permission bits of m.
func (m FileMode) Perm() FileMode {
	return m & 0777
}

func (m FileMode) String() string {
	buf := []byte("----------")

	const types = "dalMATL!DpSug"
	for i := range types {
		if m&(1<<uint(31-i)) != 0 {
			buf[0] = types[i]
		}
	}

	const perms = "rwx"
	for i := 1; i < len(buf); i++ {
		if m&(1<<uint32(len(buf)-1-i)) != 0 {
			buf[i] = perms[(i-1)%len(perms)]
		}
	}

	return *(*string)(unsafe.Pointer(&buf))
}

// Stat is a stat value.
type Stat struct {
	Type   uint16
	Dev    uint32
	QID    QID
	Mode   FileMode
	ATime  time.Time
	MTime  time.Time
	Length uint64
	Name   string
	UID    string
	GID    string
	MUID   string
}

// DirEntry returns a DirEntry that corresponds to the Stat.
func (s Stat) DirEntry() DirEntry {
	return DirEntry{
		FileMode:  s.Mode,
		ATime:     s.ATime,
		MTime:     s.MTime,
		Length:    s.Length,
		EntryName: s.Name,
		UID:       s.UID,
		GID:       s.GID,
		MUID:      s.MUID,

		Path:    s.QID.Path,
		Version: s.QID.Version,
	}
}

func (s Stat) size() uint16 {
	return uint16(47 + len(s.Name) + len(s.UID) + len(s.GID) + len(s.MUID))
}

func (s Stat) P9Encode() (r []byte, err error) {
	var buf bytes.Buffer
	write := func(v interface{}) {
		if err != nil {
			return
		}

		err = proto.Write(&buf, v)
	}

	write(s.size())
	write(s.Type)
	write(s.Dev)
	write(s.QID)
	write(s.Mode)
	write(s.ATime)
	write(s.MTime)
	write(s.Length)
	write(s.Name)
	write(s.UID)
	write(s.GID)
	write(s.MUID)

	return buf.Bytes(), err
}

func (s *Stat) P9Decode(r io.Reader) (err error) {
	var size uint16
	err = proto.Read(r, &size)
	if err != nil {
		return err
	}

	lr := &util.LimitedReader{
		R: r,
		N: uint32(size),
		E: ErrLargeStat,
	}

	read := func(v interface{}) {
		if err != nil {
			return
		}

		err = proto.Read(lr, v)
	}

	read(&s.Type)
	read(&s.Dev)
	read(&s.QID)
	read(&s.Mode)
	read(&s.ATime)
	read(&s.MTime)
	read(&s.Length)
	read(&s.Name)
	read(&s.UID)
	read(&s.GID)
	read(&s.MUID)

	return err
}

// DirEntry is a smaller version of Stat that eliminates unnecessary
// or duplicate fields.
type DirEntry struct {
	FileMode  FileMode
	ATime     time.Time
	MTime     time.Time
	Length    uint64
	EntryName string
	UID       string
	GID       string
	MUID      string

	Path    uint64
	Version uint32
}

// Stat returns a Stat that corresponds to the DirEntry.
func (d DirEntry) Stat() Stat {
	return Stat{
		Type: uint16(d.FileMode >> 16),
		QID: QID{
			Type:    QIDType(d.FileMode >> 24),
			Version: d.Version,
			Path:    d.Path,
		},
		Mode:   d.FileMode,
		ATime:  d.ATime,
		MTime:  d.MTime,
		Length: d.Length,
		Name:   d.EntryName,
		UID:    d.UID,
		GID:    d.GID,
		MUID:   d.MUID,
	}
}

func (d DirEntry) Name() string {
	return d.EntryName
}

func (d DirEntry) Size() int64 {
	return int64(d.Length)
}

func (d DirEntry) Mode() os.FileMode {
	return d.FileMode.OS()
}

func (d DirEntry) ModTime() time.Time {
	return d.MTime
}

func (d DirEntry) IsDir() bool {
	return d.FileMode&ModeDir != 0
}

func (d DirEntry) Sys() interface{} {
	return d
}

// StatChanges is a wrapper around DirEntry that is used in wstat
// requests. If one of its methods returns false, that field should be
// considered unset in the DirEntry.
type StatChanges struct {
	DirEntry
}

func (c StatChanges) Mode() (FileMode, bool) {
	return c.DirEntry.FileMode, c.DirEntry.FileMode != 0xFFFFFFFF
}

func (c StatChanges) ATime() (time.Time, bool) {
	return c.DirEntry.ATime, c.DirEntry.ATime.Unix() != -1
}

func (c StatChanges) MTime() (time.Time, bool) {
	return c.DirEntry.MTime, c.DirEntry.MTime.Unix() != -1
}

func (c StatChanges) Length() (uint64, bool) {
	return c.DirEntry.Length, c.DirEntry.Length != 0xFFFFFFFFFFFFFFFF
}

func (c StatChanges) Name() (string, bool) {
	return c.DirEntry.EntryName, c.DirEntry.EntryName != ""
}

func (c StatChanges) UID() (string, bool) {
	return c.DirEntry.UID, c.DirEntry.UID != ""
}

func (c StatChanges) GID() (string, bool) {
	return c.DirEntry.GID, c.DirEntry.GID != ""
}

func (c StatChanges) MUID() (string, bool) {
	return c.DirEntry.MUID, c.DirEntry.MUID != ""
}
