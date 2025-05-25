// Copyright 2012 The Ninep Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ufs

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lionkov/ninep"
	"github.com/lionkov/ninep/srv"
)

type Fid struct {
	path       string
	file       *os.File
	dirs       []os.FileInfo
	diroffset  uint64
	direntends []int
	dirents    []byte
	st         os.FileInfo
}

type Ufs struct {
	srv.Srv
	Root string
}

var root = flag.String("root", "/", "root filesystem")
var Enoent = &ninep.Error{"file not found", ninep.ENOENT}

func toError(err error) *ninep.Error {
	var ecode uint32

	ename := err.Error()
	if e, ok := err.(syscall.Errno); ok {
		ecode = uint32(e)
	} else {
		ecode = ninep.EIO
	}

	return &ninep.Error{ename, ecode}
}

// IsBlock reports if the file is a block device
func isBlock(d os.FileInfo) bool {
	sysif := d.Sys()
	if sysif == nil {
		return false
	}
	stat := sysif.(*syscall.Stat_t)
	return (stat.Mode & syscall.S_IFMT) == syscall.S_IFBLK
}

// IsChar reports if the file is a character device
func isChar(d os.FileInfo) bool {
	sysif := d.Sys()
	if sysif == nil {
		return false
	}
	stat := sysif.(*syscall.Stat_t)
	return (stat.Mode & syscall.S_IFMT) == syscall.S_IFCHR
}

func (fid *Fid) stat() *ninep.Error {
	var err error

	fid.st, err = os.Lstat(fid.path)
	if err != nil {
		return toError(err)
	}

	return nil
}

func omode2uflags(mode uint8) int {
	ret := int(0)
	switch mode & 3 {
	case ninep.OREAD:
		ret = os.O_RDONLY
		break

	case ninep.ORDWR:
		ret = os.O_RDWR
		break

	case ninep.OWRITE:
		ret = os.O_WRONLY
		break

	case ninep.OEXEC:
		ret = os.O_RDONLY
		break
	}

	if mode&ninep.OTRUNC != 0 {
		ret |= os.O_TRUNC
	}

	return ret
}

func dir2Qid(d os.FileInfo) *ninep.Qid {
	var qid ninep.Qid
	sysif := d.Sys()
	if sysif == nil {
		return nil
	}
	stat := sysif.(*syscall.Stat_t)

	qid.Path = stat.Ino
	qid.Version = uint32(d.ModTime().UnixNano() / 1000000)
	qid.Type = dir2QidType(d)

	return &qid
}

func dir2QidType(d os.FileInfo) uint8 {
	ret := uint8(0)
	if d.IsDir() {
		ret |= ninep.QTDIR
	}

	if d.Mode()&os.ModeSymlink != 0 {
		ret |= ninep.QTSYMLINK
	}

	return ret
}

func dir2Npmode(d os.FileInfo, dotu bool) uint32 {
	ret := uint32(d.Mode() & 0777)
	if d.IsDir() {
		ret |= ninep.DMDIR
	}

	if dotu {
		mode := d.Mode()
		if mode&os.ModeSymlink != 0 {
			ret |= ninep.DMSYMLINK
		}

		if mode&os.ModeSocket != 0 {
			ret |= ninep.DMSOCKET
		}

		if mode&os.ModeNamedPipe != 0 {
			ret |= ninep.DMNAMEDPIPE
		}

		if mode&os.ModeDevice != 0 {
			ret |= ninep.DMDEVICE
		}

		if mode&os.ModeSetuid != 0 {
			ret |= ninep.DMSETUID
		}

		if mode&os.ModeSetgid != 0 {
			ret |= ninep.DMSETGID
		}
	}

	return ret
}

// Dir is an instantiation of the ninep.Dir structure
// that can act as a receiver for local methods.
type Dir struct {
	ninep.Dir
}

func dir2Dir(s string, d os.FileInfo, dotu bool, upool ninep.Users) (*ninep.Dir, error) {
	sysif := d.Sys()
	if sysif == nil {
		return nil, &os.PathError{"dir2Dir", s, nil}
	}
	var sysMode *syscall.Stat_t
	switch t := sysif.(type) {
	case *syscall.Stat_t:
		sysMode = t
	default:
		return nil, &os.PathError{"dir2Dir: sysif has wrong type", s, nil}
	}

	dir := new(Dir)
	dir.Qid = *dir2Qid(d)
	dir.Mode = dir2Npmode(d, dotu)
	dir.Atime = uint32(atime(sysMode).Unix())
	dir.Mtime = uint32(d.ModTime().Unix())
	dir.Length = uint64(d.Size())
	dir.Name = s[strings.LastIndex(s, "/")+1:]

	if dotu {
		dir.dotu(s, d, upool, sysMode)
		return &dir.Dir, nil
	}

	unixUid := int(sysMode.Uid)
	unixGid := int(sysMode.Gid)
	dir.Uid = strconv.Itoa(unixUid)
	dir.Gid = strconv.Itoa(unixGid)
	dir.Muid = "none"

	// BUG(akumar): LookupId will never find names for
	// groups, as it only operates on user ids.
	u, err := user.LookupId(dir.Uid)
	if err == nil {
		dir.Uid = u.Username
	}
	g, err := user.LookupId(dir.Gid)
	if err == nil {
		dir.Gid = g.Username
	}

	return &dir.Dir, nil
}

func (dir *Dir) dotu(path string, d os.FileInfo, upool ninep.Users, sysMode *syscall.Stat_t) {
	u := upool.Uid2User(int(sysMode.Uid))
	g := upool.Gid2Group(int(sysMode.Gid))
	dir.Uid = u.Name()
	if dir.Uid == "" {
		dir.Uid = "none"
	}

	dir.Gid = g.Name()
	if dir.Gid == "" {
		dir.Gid = "none"
	}
	dir.Muid = "none"
	dir.Ext = ""
	dir.Uidnum = uint32(u.Id())
	dir.Gidnum = uint32(g.Id())
	dir.Muidnum = ninep.NOUID
	if d.Mode()&os.ModeSymlink != 0 {
		var err error
		dir.Ext, err = os.Readlink(path)
		if err != nil {
			dir.Ext = ""
		}
	} else if isBlock(d) {
		dir.Ext = fmt.Sprintf("b %d %d", sysMode.Rdev>>24, sysMode.Rdev&0xFFFFFF)
	} else if isChar(d) {
		dir.Ext = fmt.Sprintf("c %d %d", sysMode.Rdev>>24, sysMode.Rdev&0xFFFFFF)
	}
}

func (*Ufs) ConnOpened(conn *srv.Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("connected")
	}
}

func (*Ufs) ConnClosed(conn *srv.Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("disconnected")
	}
}

func (*Ufs) FidDestroy(sfid *srv.Fid) {
	var fid *Fid

	if sfid.Aux == nil {
		return
	}

	fid = sfid.Aux.(*Fid)
	if fid.file != nil {
		fid.file.Close()
	}
}

func (u *Ufs) Attach(req *srv.Req) {
	if req.Afid != nil {
		req.RespondError(srv.Enoauth)
		return
	}

	tc := req.Tc
	fid := new(Fid)

	// You can think of the ufs.Root as a 'chroot' of a sort.
	// client attaches are not allowed to go outside the
	// directory represented by ufs.Root
	fid.path = path.Join(u.Root, tc.Aname)

	req.Fid.Aux = fid
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	qid := dir2Qid(fid.st)
	req.RespondRattach(qid)
}

func (*Ufs) Flush(req *srv.Req) {}

func (*Ufs) Walk(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	tc := req.Tc

	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	if req.Newfid.Aux == nil {
		req.Newfid.Aux = new(Fid)
	}

	nfid := req.Newfid.Aux.(*Fid)
	wqids := make([]ninep.Qid, len(tc.Wname))
	path := fid.path
	i := 0
	for ; i < len(tc.Wname); i++ {
		p := path + "/" + tc.Wname[i]
		st, err := os.Lstat(p)
		if err != nil {
			if i == 0 {
				req.RespondError(Enoent)
				return
			}

			break
		}

		wqids[i] = *dir2Qid(st)
		path = p
	}

	nfid.path = path
	req.RespondRwalk(wqids[0:i])
}

func (*Ufs) Open(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	tc := req.Tc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	var e error
	fid.file, e = os.OpenFile(fid.path, omode2uflags(tc.Mode), 0)
	if e != nil {
		req.RespondError(toError(e))
		return
	}

	req.RespondRopen(dir2Qid(fid.st), 0)
}

func (*Ufs) Create(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	tc := req.Tc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	path := fid.path + "/" + tc.Name
	var e error = nil
	var file *os.File = nil
	switch {
	case tc.Perm&ninep.DMDIR != 0:
		e = os.Mkdir(path, os.FileMode(tc.Perm&0777))

	case tc.Perm&ninep.DMSYMLINK != 0:
		e = os.Symlink(tc.Ext, path)

	case tc.Perm&ninep.DMLINK != 0:
		n, e := strconv.ParseUint(tc.Ext, 10, 0)
		if e != nil {
			break
		}

		ofid := req.Conn.FidGet(uint32(n))
		if ofid == nil {
			req.RespondError(srv.Eunknownfid)
			return
		}

		e = os.Link(ofid.Aux.(*Fid).path, path)
		ofid.DecRef()

	case tc.Perm&ninep.DMNAMEDPIPE != 0:
	case tc.Perm&ninep.DMDEVICE != 0:
		req.RespondError(&ninep.Error{"not implemented", ninep.EIO})
		return

	default:
		var mode uint32 = tc.Perm & 0777
		if req.Conn.Dotu {
			if tc.Perm&ninep.DMSETUID > 0 {
				mode |= syscall.S_ISUID
			}
			if tc.Perm&ninep.DMSETGID > 0 {
				mode |= syscall.S_ISGID
			}
		}
		file, e = os.OpenFile(path, omode2uflags(tc.Mode)|os.O_CREATE, os.FileMode(mode))
	}

	if file == nil && e == nil {
		file, e = os.OpenFile(path, omode2uflags(tc.Mode), 0)
	}

	if e != nil {
		req.RespondError(toError(e))
		return
	}

	fid.path = path
	fid.file = file
	err = fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	req.RespondRcreate(dir2Qid(fid.st), 0)
}

func (u *Ufs) Read(req *srv.Req) {
	dbg := u.Debuglevel&srv.DbgLogFcalls != 0
	fid := req.Fid.Aux.(*Fid)
	tc := req.Tc
	rc := req.Rc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	ninep.InitRread(rc, tc.Count)
	var count int
	var e error
	if fid.st.IsDir() {
		if tc.Offset == 0 {
			var e error
			// If we got here, it was open. Can't really seek
			// in most cases, just close and reopen it.
			fid.file.Close()
			if fid.file, e = os.OpenFile(fid.path, omode2uflags(req.Fid.Omode), 0); e != nil {
				req.RespondError(toError(e))
				return
			}

			if fid.dirs, e = fid.file.Readdir(-1); e != nil {
				req.RespondError(toError(e))
				return
			}

			if dbg {
				log.Printf("Read: read %d entries", len(fid.dirs))
			}
			fid.dirents = nil
			fid.direntends = nil
			for i := 0; i < len(fid.dirs); i++ {
				path := fid.path + "/" + fid.dirs[i].Name()
				st, err := dir2Dir(path, fid.dirs[i], req.Conn.Dotu, req.Conn.Srv.Upool)
				if err != nil {
					if dbg {
						log.Printf("dbg: stat of %v: %v", path, err)
					}
					continue
				}
				if dbg {
					log.Printf("Stat: %v is %v", path, st)
				}
				b := ninep.PackDir(st, req.Conn.Dotu)
				fid.dirents = append(fid.dirents, b...)
				count += len(b)
				fid.direntends = append(fid.direntends, count)
				if dbg {
					log.Printf("fid.direntends is %v\n", fid.direntends)
				}
			}
		}

		switch {
		case tc.Offset > uint64(len(fid.dirents)):
			count = 0
		case len(fid.dirents[tc.Offset:]) > int(tc.Count):
			count = int(tc.Count)
		default:
			count = len(fid.dirents[tc.Offset:])
		}

		if dbg {
			log.Printf("readdir: count %v @ offset %v", count, tc.Offset)
		}
		nextend := sort.SearchInts(fid.direntends, int(tc.Offset)+count)
		if nextend < len(fid.direntends) {
			if fid.direntends[nextend] > int(tc.Offset)+count {
				if nextend > 0 {
					count = fid.direntends[nextend-1] - int(tc.Offset)
				} else {
					count = 0
				}
			}
		}
		if dbg {
			log.Printf("readdir: count adjusted %v @ offset %v", count, tc.Offset)
		}
		if count == 0 && int(tc.Offset) < len(fid.dirents) && len(fid.dirents) > 0 {
			req.RespondError(&ninep.Error{"too small read size for dir entry", ninep.EINVAL})
			return
		}
		copy(rc.Data, fid.dirents[tc.Offset:int(tc.Offset)+count])
	} else {
		count, e = fid.file.ReadAt(rc.Data, int64(tc.Offset))
		if e != nil && e != io.EOF {
			req.RespondError(toError(e))
			return
		}
	}

	ninep.SetRreadCount(rc, uint32(count))
	req.Respond()
}

func (*Ufs) Write(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	tc := req.Tc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	n, e := fid.file.WriteAt(tc.Data, int64(tc.Offset))
	if e != nil {
		req.RespondError(toError(e))
		return
	}

	req.RespondRwrite(uint32(n))
}

func (*Ufs) Clunk(req *srv.Req) { req.RespondRclunk() }

func (*Ufs) Remove(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	e := os.Remove(fid.path)
	if e != nil {
		req.RespondError(toError(e))
		return
	}

	req.RespondRremove()
}

func (*Ufs) Stat(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	if err := fid.stat(); err != nil {
		req.RespondError(err)
		return
	}

	st, err := dir2Dir(fid.path, fid.st, req.Conn.Dotu, req.Conn.Srv.Upool)
	if err != nil {
		req.RespondError(err)
		return
	}
	req.RespondRstat(st)
}

func lookup(uid string, group bool) (uint32, *ninep.Error) {
	if uid == "" {
		return ninep.NOUID, nil
	}
	usr, e := user.Lookup(uid)
	if e != nil {
		return ninep.NOUID, toError(e)
	}
	conv := usr.Uid
	if group {
		conv = usr.Gid
	}
	u, e := strconv.Atoi(conv)
	if e != nil {
		return ninep.NOUID, toError(e)
	}
	return uint32(u), nil
}

func (u *Ufs) Wstat(req *srv.Req) {
	var changed bool
	fid := req.Fid.Aux.(*Fid)
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	dir := &req.Tc.Dir
	if dir.Mode != 0xFFFFFFFF {
		changed = true
		mode := dir.Mode & 0777
		if req.Conn.Dotu {
			if dir.Mode&ninep.DMSETUID > 0 {
				mode |= syscall.S_ISUID
			}
			if dir.Mode&ninep.DMSETGID > 0 {
				mode |= syscall.S_ISGID
			}
		}
		e := os.Chmod(fid.path, os.FileMode(mode))
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	uid, gid := ninep.NOUID, ninep.NOUID
	if req.Conn.Dotu {
		uid = dir.Uidnum
		gid = dir.Gidnum
	}

	// Try to find local uid, gid by name.
	if (dir.Uid != "" || dir.Gid != "") && !req.Conn.Dotu {
		changed = true
		uid, err = lookup(dir.Uid, false)
		if err != nil {
			req.RespondError(err)
			return
		}

		// BUG(akumar): Lookup will never find gids
		// corresponding to group names, because
		// it only operates on user names.
		gid, err = lookup(dir.Gid, true)
		if err != nil {
			req.RespondError(err)
			return
		}
	}

	if uid != ninep.NOUID || gid != ninep.NOUID {
		changed = true
		e := os.Chown(fid.path, int(uid), int(gid))
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	if dir.Name != "" {
		changed = true
		// If we path.Join dir.Name to / before adding it to
		// the fid path, that ensures nobody gets to walk out of the
		// root of this server.
		newname := path.Join(path.Dir(fid.path), path.Join("/", dir.Name))

		// absolute renaming. Ufs can do this, so let's support it.
		// We'll allow an absolute path in the Name and, if it is,
		// we will make it relative to root. This is a gigantic performance
		// improvement in systems that allow it.
		if filepath.IsAbs(dir.Name) {
			newname = path.Join(u.Root, dir.Name)
		}

		err := syscall.Rename(fid.path, newname)
		if err != nil {
			req.RespondError(toError(err))
			return
		}
		fid.path = newname
	}

	if dir.Length != 0xFFFFFFFFFFFFFFFF {
		changed = true
		e := os.Truncate(fid.path, int64(dir.Length))
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	// If either mtime or atime need to be changed, then
	// we must change both.
	if dir.Mtime != ^uint32(0) || dir.Atime != ^uint32(0) {
		changed = true
		mt, at := time.Unix(int64(dir.Mtime), 0), time.Unix(int64(dir.Atime), 0)
		if cmt, cat := (dir.Mtime == ^uint32(0)), (dir.Atime == ^uint32(0)); cmt || cat {
			st, e := os.Stat(fid.path)
			if e != nil {
				req.RespondError(toError(e))
				return
			}
			switch cmt {
			case true:
				mt = st.ModTime()
			default:
				at = atime(st.Sys().(*syscall.Stat_t))
			}
		}
		e := os.Chtimes(fid.path, at, mt)
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	if !changed && fid.file != nil {
		fid.file.Sync()
	}
	req.RespondRwstat()
}

func New() *Ufs {
	return &Ufs{Root: *root}
}
