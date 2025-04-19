package p9

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"sync"
	"time"
	"unsafe"

	"github.com/DeedleFake/p9/internal/debug"
	"github.com/DeedleFake/p9/proto"
)

// FileSystem is an interface that allows high-level implementations
// of 9P servers by allowing the implementation to ignore the majority
// of the details of the protocol.
type FileSystem interface {
	// Auth returns an authentication file. This file can be used to
	// send authentication information back and forth between the server
	// and the client.
	//
	// The file returned from this method must report its own type as
	// QTAuth.
	Auth(user, aname string) (File, error)

	// Attach attachs to a file hierarchy provided by the filesystem.
	// If the client is attempting to authenticate, a File previously
	// returned from Auth() will be passed as afile.
	Attach(afile File, user, aname string) (Attachment, error)
}

// Attachment is a file hierarchy provided by an FS.
//
// All paths passed to the methods of this system begin with the aname
// used to attach the attachment, use forward slashes as separators,
// and have been cleaned using path.Clean. For example, if the client
// attaches using the aname "/example" and then tries to open the file
// located at "some/other/example", the Open method will be called
// with the path "/example/some/other/example".
type Attachment interface {
	// Stat returns a DirEntry giving info about the file or directory
	// at the given path. If an error is returned, the text of the error
	// will be transmitted to the client.
	Stat(path string) (DirEntry, error)

	// WriteStat applies changes to the metadata of the file at path.
	// If a method of the changes argument returns false, it should not
	// be changed.
	//
	// If an error is returned, it is assumed that the entire operation
	// failed. In particular, name changes will not be tracked by the
	// system if this returns an error.
	WriteStat(path string, changes StatChanges) error

	// Open opens the file at path in the given mode. If an error is
	// returned, it will be transmitted to the client.
	Open(path string, mode uint8) (File, error)

	// Create creates and opens a file at path with the given perms and
	// mode. If an error is returned, it will be transmitted to the
	// client.
	//
	// When creating a directory, this method will be called with the
	// DMDIR bit set in perm. Even in this situation it should still
	// open and return the newly created file.
	Create(path string, perm FileMode, mode uint8) (File, error)

	// Remove deletes the file at path, returning any errors encountered.
	Remove(path string) error
}

// IOUnitFS is implemented by Attachments that want to report an
// IOUnit value to clients when open and create requests are made. An
// IOUnit value lets the client know the maximum amount of data during
// reads and writes that is guaranteed to be an atomic operation.
type IOUnitFS interface {
	IOUnit() uint32
}

// QIDFS is implemented by Attachments that want to deal with QIDs
// manually. A QID represents a unique identifier for a given file. In
// particular, the Path field must be unique for every path, even if
// the file at that path has been deleted and replaced with a
// completely new file.
type QIDFS interface {
	GetQID(p string) (QID, error)
}

// File is the interface implemented by files being dealt with by a
// FileSystem.
//
// Note that although this interface implements io.ReaderAt and
// io.WriterAt, a number of the restrictions placed on those
// interfaces may be ignored. The implementation need only follow
// restrictions placed by the 9P protocol specification.
type File interface {
	// Used to handle 9P read requests.
	io.ReaderAt

	// Used to handle 9P write requests.
	io.WriterAt

	// Used to handle 9P clunk requests.
	io.Closer

	// Readdir is called when an attempt is made to read a directory. It
	// should return a list of entries in the directory or an error. If
	// an error is returned, the error will be transmitted to the
	// client.
	//
	// When a client attempts to read a file, if file reports itself as
	// a QTDir, then this method will be used to read it instead of
	// ReadAt().
	Readdir() ([]DirEntry, error)
}

type fsFile struct {
	sync.RWMutex

	path string

	a Attachment

	file File
	dir  bytes.Buffer
}

type fsHandler struct {
	fs    FileSystem
	msize uint32

	fids sync.Map // map[uint32]*fsFile
}

// FSHandler returns a MessageHandler that provides a virtual
// filesystem using the provided FileSystem implementation. msize is
// the maximum size that messages from either the server or the client
// are allowed to be.
//
// The returned MessageHandler implementation will print debug
// messages to stderr if the p9debug build tag is set.
//
// BUG: Tflush requests are not currently handled at all by this
// implementation due to no clear method of stopping a pending call to
// ReadAt or WriteAt.
func FSHandler(fs FileSystem, msize uint32) proto.MessageHandler {
	return &fsHandler{
		fs:    fs,
		msize: msize,
	}
}

// FSConnHandler returns a ConnHandler that calls FSHandler to
// generate MessageHandlers.
//
// The returned ConnHandler implementation will print debug messages
// to stderr if the p9debug build tag is set.
func FSConnHandler(fs FileSystem, msize uint32) proto.ConnHandler {
	return proto.ConnHandlerFunc(func() proto.MessageHandler {
		debug.Log("Got new connection to FSConnHandler.\n")
		return FSHandler(fs, msize)
	})
}

func (h *fsHandler) getQID(p string, attach Attachment) (QID, error) {
	if q, ok := attach.(QIDFS); ok {
		return q.GetQID(p)
	}

	stat, err := attach.Stat(p)
	if err != nil {
		return QID{}, err
	}

	sum := sha256.Sum256(unsafe.Slice(unsafe.StringData(p), len(p)))
	path := binary.LittleEndian.Uint64(sum[:])

	return QID{
		Type: stat.FileMode.QIDType(),
		Path: path,
	}, nil
}

func (h *fsHandler) getFile(fid uint32, create bool) (*fsFile, bool) {
	if create {
		f, ok := h.fids.LoadOrStore(fid, new(fsFile))
		return f.(*fsFile), ok
	}

	f, ok := h.fids.Load(fid)
	if !ok {
		return nil, false
	}
	return f.(*fsFile), true
}

func (h *fsHandler) largeCount(count uint32) bool {
	return IOHeaderSize+count > h.msize
}

func (h *fsHandler) version(msg *Tversion) any {
	if msg.Version != Version {
		return &Rerror{
			Ename: ErrUnsupportedVersion.Error(),
		}
	}

	if h.msize > msg.Msize {
		h.msize = msg.Msize
	}

	return &Rversion{
		Msize:   h.msize,
		Version: Version,
	}
}

func (h *fsHandler) auth(msg *Tauth) any {
	file, err := h.fs.Auth(msg.Uname, msg.Aname)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	f, _ := h.getFile(msg.AFID, true)
	f.Lock()
	defer f.Unlock()

	f.file = file

	return &Rauth{
		AQID: QID{
			Type: QTAuth,
			Path: uint64(time.Now().UnixNano()),
		},
	}
}

func (h *fsHandler) flush(msg *Tflush) any {
	// TODO: Implement this.
	return &Rerror{
		Ename: "flush is not supported",
	}
}

func (h *fsHandler) attach(msg *Tattach) any {
	var afile File
	if msg.AFID != NoFID {
		tmp, ok := h.getFile(msg.AFID, false)
		if !ok {
			return &Rerror{
				Ename: "no such AFID",
			}
		}

		tmp.RLock()
		afile = tmp.file
		tmp.RUnlock()
	}

	attach, err := h.fs.Attach(afile, msg.Uname, msg.Aname)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	qid, err := h.getQID(msg.Aname, attach)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	file, ok := h.getFile(msg.FID, true)
	if ok {
		return &Rerror{
			Ename: "FID in use",
		}
	}
	file.Lock()
	defer file.Unlock()

	file.path = msg.Aname
	file.a = attach

	return &Rattach{
		QID: qid,
	}
}

func (h *fsHandler) walk(msg *Twalk) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}

	file.RLock()
	base := file.path
	a := file.a
	file.RUnlock()

	qids := make([]QID, 0, len(msg.Wname))
	for i, name := range msg.Wname {
		next := path.Join(base, name)

		qid, err := h.getQID(next, a)
		if err != nil {
			if i == 0 {
				return &Rerror{
					Ename: err.Error(),
				}
			}

			return &Rwalk{
				WQID: qids,
			}
		}

		qids = append(qids, qid)
		base = next
	}

	file, ok = h.getFile(msg.NewFID, true)
	if ok {
		return &Rerror{
			Ename: "FID in use",
		}
	}
	file.Lock()
	defer file.Unlock()

	file.path = base
	file.a = a

	return &Rwalk{
		WQID: qids,
	}
}

func (h *fsHandler) open(msg *Topen) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.Lock()
	defer file.Unlock()

	if file.file != nil {
		return &Rerror{
			Ename: "file already open",
		}
	}

	f, err := file.a.Open(file.path, msg.Mode)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}
	file.file = f

	qid, err := h.getQID(file.path, file.a)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	var iounit uint32
	if unit, ok := file.a.(IOUnitFS); ok {
		iounit = unit.IOUnit()
	}

	return &Ropen{
		QID:    qid,
		IOUnit: iounit,
	}
}

func (h *fsHandler) create(msg *Tcreate) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.Lock()
	defer file.Unlock()

	if file.file != nil {
		return &Rerror{
			Ename: "file already open",
		}
	}

	p := path.Join(file.path, msg.Name)

	f, err := file.a.Create(p, msg.Perm, msg.Mode)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	file.path = p
	file.file = f

	qid, err := h.getQID(p, file.a)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	var iounit uint32
	if unit, ok := file.a.(IOUnitFS); ok {
		iounit = unit.IOUnit()
	}

	return &Rcreate{
		QID:    qid,
		IOUnit: iounit,
	}
}

func (h *fsHandler) read(msg *Tread) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.Lock()
	defer file.Unlock()

	if file.file == nil {
		return &Rerror{
			Ename: "file not open",
		}
	}

	qid, err := h.getQID(file.path, file.a)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	if h.largeCount(msg.Count) {
		return &Rerror{
			Ename: "read too large",
		}
	}

	var n int
	buf := make([]byte, msg.Count)

	switch {
	case qid.Type&QTDir != 0:
		if msg.Offset == 0 {
			dir, err := file.file.Readdir()
			if err != nil {
				return &Rerror{
					Ename: err.Error(),
				}
			}

			for i := range dir {
				qid, err := h.getQID(path.Join(file.path, dir[i].EntryName), file.a)
				if err != nil {
					return &Rerror{
						Ename: err.Error(),
					}
				}

				dir[i].Version = qid.Version
				dir[i].Path = qid.Path
			}

			file.dir.Reset()
			err = WriteDir(&file.dir, dir)
			if err != nil {
				return &Rerror{
					Ename: err.Error(),
				}
			}
		}

		// This technically isn't quite accurate to the 9P specification.
		// The specification states that all reads of a directory must
		// either be at offset 0 or at the previous offset plus the length
		// of the previous read. Instead, this implemenation just ignores
		// the offset if it's not zero. This is backwards compatible with
		// the specification, however, so it's probably not really an
		// issue.
		tmp, err := file.dir.Read(buf)
		if (err != nil) && (err != io.EOF) {
			return &Rerror{
				Ename: err.Error(),
			}
		}
		n = tmp

	default:
		tmp, err := file.file.ReadAt(buf, int64(msg.Offset))
		if (err != nil) && (err != io.EOF) {
			return &Rerror{
				Ename: err.Error(),
			}
		}
		n = tmp
	}

	return &Rread{
		Data: buf[:n],
	}
}

func (h *fsHandler) write(msg *Twrite) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.RLock()         // Somewhat ironic that this doesn't require a
	defer file.RUnlock() // full lock like read() does.

	if file.file == nil {
		return &Rerror{
			Ename: "file not open",
		}
	}

	n, err := file.file.WriteAt(msg.Data, int64(msg.Offset))
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	return &Rwrite{
		Count: uint32(n),
	}
}

func (h *fsHandler) clunk(msg *Tclunk) any {
	defer h.fids.Delete(msg.FID)

	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.RLock()
	defer file.RUnlock()

	if file.file == nil {
		return &Rclunk{}
	}

	err := file.file.Close()
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	return &Rclunk{}
}

func (h *fsHandler) remove(msg *Tremove) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.RLock()
	defer file.RUnlock()

	rsp := h.clunk(&Tclunk{
		FID: msg.FID,
	})
	if _, ok := rsp.(error); ok {
		return rsp
	}

	err := file.a.Remove(file.path)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	return &Rremove{}
}

func (h *fsHandler) stat(msg *Tstat) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.RLock()
	defer file.RUnlock()

	stat, err := file.a.Stat(file.path)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	qid, err := h.getQID(file.path, file.a)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}
	stat.Version = qid.Version
	stat.Path = qid.Path

	return &Rstat{
		Stat: stat.Stat(),
	}
}

func (h *fsHandler) wstat(msg *Twstat) any {
	file, ok := h.getFile(msg.FID, false)
	if !ok {
		return &Rerror{
			Ename: "unknown FID",
		}
	}
	file.RLock()
	defer file.RUnlock()

	changes := StatChanges{
		DirEntry: msg.Stat.DirEntry(),
	}

	err := file.a.WriteStat(file.path, changes)
	if err != nil {
		return &Rerror{
			Ename: err.Error(),
		}
	}

	return &Rwstat{}
}

func (h *fsHandler) HandleMessage(msg any) (r any) {
	defer func() {
		debug.Log("%#v\n", r)
	}()

	debug.Log("%#v\n", msg)

	switch msg := msg.(type) {
	case *Tversion:
		return h.version(msg)

	case *Tauth:
		return h.auth(msg)

	case *Tflush:
		return h.flush(msg)

	case *Tattach:
		return h.attach(msg)

	case *Twalk:
		return h.walk(msg)

	case *Topen:
		return h.open(msg)

	case *Tcreate:
		return h.create(msg)

	case *Tread:
		return h.read(msg)

	case *Twrite:
		return h.write(msg)

	case *Tclunk:
		return h.clunk(msg)

	case *Tremove:
		return h.remove(msg)

	case *Tstat:
		return h.stat(msg)

	case *Twstat:
		return h.wstat(msg)

	default:
		return &Rerror{
			Ename: fmt.Sprintf("unexpected message type: %T", msg),
		}
	}
}

func (h *fsHandler) Close() error {
	h.fids.Range(func(k, v any) bool {
		file := v.(*fsFile)
		if file.file != nil {
			file.file.Close()
		}
		return true
	})

	return nil
}
