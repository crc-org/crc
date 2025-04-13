package real

import (
	"fmt"
	"hash/crc64"
	"log"
	"os"
	"path"

	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

type Dir struct {
	Path string
}

var _ fs.Dir = &Dir{}

func (f *Dir) Parent() fs.Dir {
	if f.Path == "/" {
		return nil
	}
	return &Dir{path.Dir(f.Path)}
}

func (f *Dir) SetParent(d fs.Dir) {
	panic("THIS SHOULD NOT HAPPEN")
}

var crc64Table = crc64.MakeTable(0xC96C5795D7870F42)

func (f *Dir) Stat() proto.Stat {
	info, err := os.Stat(f.Path)
	if err != nil {
		log.Printf("Failed to stat %s: %s", f.Path, err)
		return proto.Stat{}
	}
	u, g, err := getUserGroup(info)
	if err != nil {
		u = "?"
		g = "?"
	}
	mode := uint32(info.Mode())
	return proto.Stat{
		Qid: proto.Qid{
			Qtype: uint8(mode >> 24),
			Vers:  uint32(info.ModTime().Unix()),
			Uid:   crc64.Checksum([]byte(f.Path), crc64Table),
		},
		Mode:   uint32(mode),
		Atime:  uint32(0),
		Mtime:  uint32(info.ModTime().Unix()),
		Length: uint64(info.Size()),
		Name:   info.Name(),
		Uid:    u,
		Gid:    g,
		Muid:   "",
	}
}

func (f *Dir) WriteStat(s *proto.Stat) error {
	current := f.Stat()
	if s.Mode != current.Mode {
		return os.Chmod(f.Path, os.FileMode(s.Mode))
	}
	if s.Length != current.Length {
		return os.Truncate(f.Path, int64(s.Length))
	}
	if s.Name != current.Name {
		dir := path.Dir(f.Path)
		newPath := path.Join(dir, s.Name)
		err := os.Rename(f.Path, newPath)
		if err != nil {
			return err
		}
		f.Path = newPath
		return nil
	}
	if s.Uid != current.Uid {
		//log.Printf("OLD Uid: %s NEW Uid: %s\n", current.Uid, s.Uid)
		return fmt.Errorf("Owner change not implemented")
	}
	if s.Gid != current.Gid {
		//log.Printf("OLD Gid: %s NEW Gid: %s\n", current.Gid, s.Gid)
		return fmt.Errorf("Group change not implemented")
	}
	return nil
}

func (d *Dir) Children() map[string]fs.FSNode {
	f, err := os.Open(d.Path)
	defer f.Close()
	if err != nil {
		log.Printf("Failed to list path %s: %s", d.Path, err)
		return nil
	}
	infos, err := f.Readdir(-1)
	if err != nil {
		log.Printf("Failed to list path %s: %s", d.Path, err)
		return nil
	}
	m := make(map[string]fs.FSNode)
	for i := range infos {
		if infos[i].IsDir() {
			m[infos[i].Name()] = &Dir{Path: path.Join(d.Path, infos[i].Name())}
		} else {
			//m[infos[i].Name()] = &RealFile{BaseFile: *fs.NewBaseFile(exportFS.NewStat(infos[i].Name(), user, group, uint32(infos[i].Mode()))), Path: path.Join(d.Path, infos[i].Name()), opens: make(map[uint64]*os.File)}
			m[infos[i].Name()] = NewFile(path.Join(d.Path, infos[i].Name()))
		}
	}
	return m
}

// CreateDir is a function meant to be passed to WithCreateDir.
// It creates a real directory under the parent
func CreateDir(filesystem *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.Dir, error) {
	fullPath := path.Join(fs.FullPath(parent), name)
	err := os.Mkdir(fullPath, os.FileMode(perm))
	if err != nil {
		return nil, err
	}
	return &Dir{fullPath}, nil
}
