package p9

import (
	"errors"
	"os"
	"path/filepath"
)

// Dir is an implementation of FileSystem that serves from the local
// filesystem. It accepts attachments of either "" or "/", but rejects
// all others.
//
// Note that Dir does not support authentication, simply returning an
// error for any attempt to do so. If authentication is necessary,
// wrap a Dir in an AuthFS instance.
type Dir string

func (d Dir) path(p string) string {
	return filepath.Join(string(d), filepath.FromSlash(p))
}

// Stat implements Attachment.Stat.
func (d Dir) Stat(p string) (DirEntry, error) {
	fi, err := os.Stat(d.path(p))
	if err != nil {
		return DirEntry{}, err
	}

	e := infoToEntry(fi)
	if e.EntryName == "." {
		e.EntryName = ""
	}

	return e, nil
}

// WriteStat implements Attachment.WriteStat.
func (d Dir) WriteStat(p string, changes StatChanges) error {
	// TODO: Add support for other values.

	p = d.path(p)
	base := filepath.Dir(p)

	mode, ok := changes.Mode()
	if ok {
		err := os.Chmod(p, mode.OS())
		if err != nil {
			return err
		}
	}

	atime, ok1 := changes.ATime()
	mtime, ok2 := changes.MTime()
	if ok1 || ok2 {
		err := os.Chtimes(p, atime, mtime)
		if err != nil {
			return err
		}
	}

	length, ok := changes.Length()
	if ok {
		err := os.Truncate(p, int64(length))
		if err != nil {
			return err
		}
	}

	name, ok := changes.Name()
	if ok {
		err := os.Rename(p, filepath.Join(base, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
	}

	return nil
}

// Auth implements FileSystem.Auth.
func (d Dir) Auth(user, aname string) (File, error) {
	return nil, errors.New("auth not supported")
}

// Attach implements FileSystem.Attach.
func (d Dir) Attach(afile File, user, aname string) (Attachment, error) {
	switch aname {
	case "", "/":
		return d, nil
	}

	return nil, errors.New("unknown attachment")
}

// Open implements Attachment.Open.
func (d Dir) Open(p string, mode uint8) (File, error) {
	flag := toOSFlags(mode)

	file, err := os.OpenFile(d.path(p), flag, 0644)
	return &dirFile{
		File: file,
	}, err
}

// Create implements Attachment.Create.
func (d Dir) Create(p string, perm FileMode, mode uint8) (File, error) {
	p = d.path(p)

	if perm&ModeDir != 0 {
		err := os.Mkdir(p, os.FileMode(perm.Perm()))
		if err != nil {
			return nil, err
		}
	}

	flag := toOSFlags(mode)

	file, err := os.OpenFile(p, flag|os.O_CREATE, os.FileMode(perm.Perm()))
	return &dirFile{
		File: file,
	}, err
}

// Remove implements Attachment.Remove.
func (d Dir) Remove(p string) error {
	return os.Remove(d.path(p))
}

type dirFile struct {
	*os.File
}

func (f *dirFile) Readdir() ([]DirEntry, error) {
	fi, err := f.File.Readdir(-1)
	if err != nil {
		return nil, err
	}

	entries := make([]DirEntry, 0, len(fi))
	for _, info := range fi {
		entries = append(entries, infoToEntry(info))
	}
	return entries, nil
}

// ReadOnlyFS wraps a filesystem implementation with an implementation
// that rejects any attempts to cause changes to the filesystem with
// the exception of writing to an authfile.
func ReadOnlyFS(fs FileSystem) FileSystem {
	return &readOnlyFS{fs}
}

type readOnlyFS struct {
	FileSystem
}

func (ro readOnlyFS) Attach(afile File, user, aname string) (Attachment, error) {
	a, err := ro.FileSystem.Attach(afile, user, aname)
	if err != nil {
		return nil, err
	}

	return passQIDFS(&readOnlyAttachment{a}, a), nil
}

type qidfsPasser struct {
	Attachment
	QIDFS
}

func passQIDFS(w Attachment, u Attachment) Attachment {
	if q, ok := u.(QIDFS); ok {
		return &qidfsPasser{
			Attachment: w,
			QIDFS:      q,
		}
	}

	return w
}

type readOnlyAttachment struct {
	Attachment
}

func (ro readOnlyAttachment) WriteStat(path string, changes StatChanges) error {
	return errors.New("read-only filesystem")
}

func (ro readOnlyAttachment) Open(path string, mode uint8) (File, error) {
	if mode&(OWRITE|ORDWR|OEXEC|OTRUNC|OCEXEC|ORCLOSE) != 0 {
		return nil, errors.New("read-only filesystem")
	}

	return ro.Attachment.Open(path, mode)
}

func (ro readOnlyAttachment) Create(path string, perm FileMode, mode uint8) (File, error) {
	return nil, errors.New("read-only filesystem")
}

func (ro readOnlyAttachment) Remove(path string) error {
	return errors.New("read-only filesystem")
}

// AuthFS allows simple wrapping and overwriting of the Auth and
// Attach methods of an existing FileSystem implementation, allowing
// the user to add authentication support to a FileSystem that does
// not have it, or to change the implementation of that support for
// FileSystems that do.
type AuthFS struct {
	FileSystem

	// AuthFunc is the function called when the Auth() method is called.
	AuthFunc func(user, aname string) (File, error)

	// AttachFunc is the function called when the Attach() method is
	// called. Note that this function, unlike the method, does not
	// return an Attachment. Instead, if this function returns a nil
	// error, the underlying implementation's Attach() method is called
	// with the returned file as its afile argument.
	AttachFunc func(afile File, user, aname string) (File, error)
}

// Auth implements FileSystem.Auth.
func (a AuthFS) Auth(user, aname string) (File, error) {
	return a.AuthFunc(user, aname)
}

// Attach implements FileSystem.Attach.
func (a AuthFS) Attach(afile File, user, aname string) (Attachment, error) {
	file, err := a.AttachFunc(afile, user, aname)
	if err != nil {
		return nil, err
	}
	return a.FileSystem.Attach(file, user, aname)
}

func toOSFlags(mode uint8) (flag int) {
	if mode&OREAD != 0 {
		flag |= os.O_RDONLY
	}
	if mode&OWRITE != 0 {
		flag |= os.O_WRONLY
	}
	if mode&ORDWR != 0 {
		flag |= os.O_RDWR
	}
	if mode&OTRUNC != 0 {
		flag |= os.O_TRUNC
	}
	//if mode&OEXCL != 0 {
	//	flag |= os.O_EXCL
	//}
	//if mode&OAPPEND != 0 {
	//	flag |= os.O_APPEND
	//}

	return flag
}
