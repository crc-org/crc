package fs

import (
	"fmt"
	"sync"

	"github.com/knusbaum/go9p/proto"
)

// StaticFile implements File. It is a very simple
// implementation that allows the reading and writing
// of a byte slice that every client sees. Writes modify
// the content and reads serve the content.
type StaticFile struct {
	BaseFile
	Data []byte
}

// NewStaticFile returns a StaticFile that contains the
// byte slice data.
func NewStaticFile(s *proto.Stat, data []byte) *StaticFile {
	s.Length = uint64(len(data))
	return &StaticFile{
		BaseFile: BaseFile{fStat: *s},
		Data:     data,
	}
}

func (f *StaticFile) Stat() proto.Stat {
	f.Lock()
	defer f.Unlock()
	f.fStat.Length = uint64(len(f.Data))
	return f.fStat
}

func (f *StaticFile) Open(fid uint64, omode proto.Mode) error {
	if omode&proto.Otrunc > 0 {
		f.Lock()
		defer f.Unlock()
		f.Data = make([]byte, 0)
	}
	return nil
}

func (f *StaticFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	f.RLock()
	defer f.RUnlock()
	flen := uint64(len(f.Data))
	if offset >= flen {
		return []byte{}, nil
	}
	if offset+count > flen {
		count = flen - offset
	}
	return f.Data[offset : offset+count], nil
}

func (f *StaticFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	f.Lock()
	defer f.Unlock()
	flen := uint64(len(f.Data))
	count := uint64(len(data))
	if offset+count > flen {
		newlen := offset + count
		f.fStat.Length = newlen
		// TODO: Maybe this can be optimized
		f.Data = append(f.Data, make([]byte, newlen-flen)...)
	}

	copy(f.Data[offset:offset+count], data)
	return uint32(len(data)), nil
}

// StaticDir is a Dir that simply keeps track of a
// set of child Files.
type StaticDir struct {
	dStat proto.Stat
	//children map[string]FSNode
	children []FSNode
	parent   Dir
	sync.RWMutex
}

func NewStaticDir(stat *proto.Stat) *StaticDir {
	dir := &StaticDir{
		dStat: *stat,
		//children: make(map[string]FSNode),
		children: make([]FSNode, 0),
	}
	// Make sure stat is marked as a directory.
	dir.dStat.Mode |= proto.DMDIR
	// qtype bits should be consistent with Stat mode.
	dir.dStat.Qid.Qtype = uint8(dir.dStat.Mode >> 24)
	return dir
}

func (d *StaticDir) Stat() proto.Stat {
	return d.dStat
}

func (d *StaticDir) WriteStat(s *proto.Stat) error {
	d.Lock()
	defer d.Unlock()
	d.dStat = *s
	return nil
}

func (d *StaticDir) SetParent(p Dir) {
	d.Lock()
	defer d.Unlock()
	d.parent = p
}

func (d *StaticDir) Parent() Dir {
	d.RLock()
	defer d.RUnlock()
	return d.parent
}

func (d *StaticDir) Children() map[string]FSNode {
	d.RLock()
	defer d.RUnlock()
	ret := make(map[string]FSNode)
	for _, n := range d.children {
		ret[n.Stat().Name] = n
	}
	return ret
}

func (d *StaticDir) AddChild(n FSNode) error {
	d.Lock()
	defer d.Unlock()
	stat := n.Stat()
	for _, n := range d.children {
		if n.Stat().Name == stat.Name {
			return fmt.Errorf("%s already exists", stat.Name)
		}
	}
	d.children = append(d.children, n)
	n.SetParent(d)
	return nil
}

func (d *StaticDir) DeleteChild(name string) error {
	d.Lock()
	defer d.Unlock()
	k := 0
	for _, c := range d.children {
		if c.Stat().Name != name {
			d.children[k] = c
			k++
		} else {
			c.SetParent(nil)
		}
	}
	d.children = d.children[:k]
	return nil
}

// CreateStaticFile is a function meant to be passed to WithCreateFile.
// It will add an empty StaticFile to the FS whenever a client attempts to
// create a file.
func CreateStaticFile(fs *FS, parent Dir, user, name string, perm uint32, mode uint8) (File, error) {
	modParent, ok := parent.(ModDir)
	if !ok {
		return nil, fmt.Errorf("%s does not support modification.", FullPath(parent))
	}
	f := NewStaticFile(fs.NewStat(name, user, user, perm), []byte{})
	err := modParent.AddChild(f)
	return f, err
}

// CreateStaticDir is a function meant to be passed to WithCreateDir.
// It will add an empty StaticDir to the FS whenever a client attempts to
// create a directory.
func CreateStaticDir(fs *FS, parent Dir, user, name string, perm uint32, mode uint8) (Dir, error) {
	modParent, ok := parent.(ModDir)
	if !ok {
		return nil, fmt.Errorf("%s does not support modification.", FullPath(parent))
	}
	f := NewStaticDir(fs.NewStat(name, user, user, perm))
	err := modParent.AddChild(f)
	return f, err
}
