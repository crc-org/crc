package p9

import (
	"bufio"
	"errors"
	"io"
	"path"
	"strings"
	"sync"

	"github.com/DeedleFake/p9/internal/util"
)

// Remote provides a file-like interface for performing operations on
// files presented by a 9P server.
//
// Remote implements File, allowing it to be itself served using
// FileSystem.
type Remote struct {
	client *Client

	fid uint32
	qid QID

	m   sync.Mutex
	pos uint64
}

// Type returns the type of the file represented by the Remote.
func (file *Remote) Type() QIDType {
	return file.qid.Type
}

func (file *Remote) walk(p string) (*Remote, error) {
	fid := file.client.nextFID()

	w := []string{path.Clean(p)}
	if w[0] != "/" {
		w = strings.Split(w[0], "/")
	}
	if (len(w) == 1) && (w[0] == ".") {
		w = nil
	}
	rsp, err := file.client.Send(&Twalk{
		FID:    file.fid,
		NewFID: fid,
		Wname:  w,
	})
	if err != nil {
		return nil, err
	}
	walk := rsp.(*Rwalk)

	qid := file.qid
	if len(walk.WQID) != 0 {
		qid = walk.WQID[len(walk.WQID)-1]
	}
	if len(walk.WQID) != len(w) {
		qid = QID{
			Type:    0xFF,
			Version: 0xFFFFFFFF,
			Path:    0xFFFFFFFFFFFFFFFF,
		}
	}

	return &Remote{
		client: file.client,
		fid:    fid,
		qid:    qid,
	}, nil
}

// Open opens and returns a file relative to the current one. In many
// cases, this will likely be relative to the filesystem root. For
// example:
//
//	root, _ := client.Attach(nil, "anyone", "/")
//	file, _ := root.Open("some/file/or/another", p9.OREAD)
func (file *Remote) Open(p string, mode uint8) (*Remote, error) {
	next, err := file.walk(p)
	if err != nil {
		return nil, err
	}

	rsp, err := file.client.Send(&Topen{
		FID:  next.fid,
		Mode: mode,
	})
	if err != nil {
		return nil, err
	}
	open := rsp.(*Ropen)

	next.qid = open.QID

	return next, nil
}

// Create creates, with the given permissions, and opens, with the
// given mode, a new file at p relative to the current file.
func (file *Remote) Create(p string, perm FileMode, mode uint8) (*Remote, error) {
	dir, name := path.Split(p)
	next, err := file.walk(path.Clean(dir))
	if err != nil {
		return nil, err
	}

	rsp, err := file.client.Send(&Tcreate{
		FID:  next.fid,
		Name: name,
		Perm: perm,
		Mode: mode,
	})
	if err != nil {
		return nil, err
	}
	create := rsp.(*Rcreate)

	next.qid = create.QID

	return next, nil
}

// Remove deletes the file at p, relative to the current file. If p is
// "", it closes the current file, if open, and deletes it.
func (file *Remote) Remove(p string) error {
	if p != "" {
		file, err := file.walk(p)
		if err != nil {
			return err
		}
		// Close is not necessary. Remove is also a clunk.

		return file.Remove("")
	}

	_, err := file.client.Send(&Tremove{
		FID: file.fid,
	})
	return err
}

// Seek seeks a file. As 9P requires clients to track their own
// positions in files, this is purely a local operation with the
// exception of the case of whence being io.SeekEnd, in which case a
// request will be made to the server in order to get the file's size.
func (file *Remote) Seek(offset int64, whence int) (int64, error) {
	file.m.Lock()
	defer file.m.Unlock()

	switch whence {
	case io.SeekStart:
		if offset < 0 {
			return int64(file.pos), errors.New("negative offset")
		}

		file.pos = uint64(offset)
		return offset, nil

	case io.SeekCurrent:
		npos := int64(file.pos) + offset
		if npos < 0 {
			return int64(file.pos), errors.New("negative offset")
		}

		file.pos = uint64(npos)
		return npos, nil

	case io.SeekEnd:
		stat, err := file.Stat("")
		if err != nil {
			return int64(file.pos), err
		}

		npos := int64(stat.Length) + offset
		if npos < 0 {
			return int64(file.pos), errors.New("negative offset")
		}

		file.pos = uint64(npos)
		return npos, nil
	}

	panic(util.Errorf("Invalid whence: %v", whence))
}

// Read reads from the file at the internally-tracked offset. For more
// information, see ReadAt.
func (file *Remote) Read(buf []byte) (int, error) {
	file.m.Lock()
	defer file.m.Unlock()

	n, err := file.ReadAt(buf, int64(file.pos))
	file.pos += uint64(n)
	return n, err
}

func (file *Remote) maxBufSize() int {
	return int(file.client.Msize() - IOHeaderSize)
}

func (file *Remote) readPart(buf []byte, off int64) (int, error) {
	rsp, err := file.client.Send(&Tread{
		FID:    file.fid,
		Offset: uint64(off),
		Count:  uint32(len(buf)),
	})
	if err != nil {
		return 0, err
	}
	read := rsp.(*Rread)
	if len(read.Data) == 0 {
		return 0, io.EOF
	}

	n := copy(buf, read.Data)
	return n, nil
}

// ReadAt reads from the file at the given offset. If the buffer given
// will result in a response that is larger than the currently allowed
// message size, as established by the handshake, it will perform
// multiple read requests in sequence, reading each into the
// appropriate parts of the buffer. It returns the number of bytes
// read and an error, if any occurred.
//
// If an error occurs while performing the sequential requests, it
// will return immediately.
func (file *Remote) ReadAt(buf []byte, off int64) (int, error) {
	size := min(len(buf), file.maxBufSize())

	var total int
	for start := 0; start < len(buf); start += size {
		end := min(start+size, len(buf))

		n, err := file.readPart(buf[start:end], off+int64(start))
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// Write writes to the file at the internally-tracked offset. For more
// information, see WriteAt.
func (file *Remote) Write(data []byte) (int, error) {
	file.m.Lock()
	defer file.m.Unlock()

	n, err := file.WriteAt(data, int64(file.pos))
	file.pos += uint64(n)
	return n, err
}

func (file *Remote) writePart(data []byte, off int64) (int, error) {
	rsp, err := file.client.Send(&Twrite{
		FID:    file.fid,
		Offset: uint64(off),
		Data:   data,
	})
	if err != nil {
		return 0, err
	}
	write := rsp.(*Rwrite)

	if write.Count < uint32(len(data)) {
		return int(write.Count), io.EOF
	}
	return int(write.Count), nil
}

// WriteAt writes from the file at the given offset. If the buffer
// given will result in a request that is larger than the currently
// allowed message size, as established by the handshake, it will
// perform multiple write requests in sequence, writing each with
// appropriate offsets such that the entire buffer is written. It
// returns the number of bytes written and an error, if any occurred.
//
// If an error occurs while performing the sequential requests, it
// will return immediately.
func (file *Remote) WriteAt(data []byte, off int64) (int, error) {
	size := min(len(data), file.maxBufSize())

	var total int
	for start := 0; start < len(data); start += size {
		end := min(start+size, len(data))

		n, err := file.writePart(data[start:end], off+int64(start))
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// Close closes the file on the server. Further usage of the file will
// produce errors.
func (file *Remote) Close() error {
	_, err := file.client.Send(&Tclunk{
		FID: file.fid,
	})
	return err
}

// Stat fetches and returns the DirEntry for the file located at p,
// relative to the current file. If p is "", it is considered to be
// the current file.
func (file *Remote) Stat(p string) (DirEntry, error) {
	if p != "" {
		file, err := file.walk(p)
		if err != nil {
			return DirEntry{}, err
		}
		defer file.Close()

		return file.Stat("")
	}

	rsp, err := file.client.Send(&Tstat{
		FID: file.fid,
	})
	if err != nil {
		return DirEntry{}, err
	}
	stat := rsp.(*Rstat)

	return stat.Stat.DirEntry(), nil
}

// Readdir reads the file as a directory, returning the list of
// directory entries returned by the server.
//
// As it returns the full list of entries, it never returns io.EOF.
//
// Note that to read this list again, the file must first be seeked to
// the beginning.
func (file *Remote) Readdir() ([]DirEntry, error) {
	return ReadDir(bufio.NewReaderSize(file, file.maxBufSize()))
}
