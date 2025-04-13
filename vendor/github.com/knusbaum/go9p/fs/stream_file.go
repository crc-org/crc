package fs

import (
	"errors"
	"fmt"

	"github.com/knusbaum/go9p/proto"
)

// StreamFile implements a one-way stream. Writes by clients
// return errors.
type StreamFile struct {
	*BaseFile
	s         Stream
	fidReader map[uint64]StreamReader
}

// BiDiStreamFile implements a two-way (full-duplex) stream. Writes
// to the stream will be read by clients who open the file. Writes
// by clients will be received by calls to Read() on the stream.
type BiDiStreamFile struct {
	*BaseFile
	s         BiDiStream
	fidReader map[uint64]StreamReadWriter
}

// NewStreamFile creates a file that serves a stream to clients.
// If the Stream s implements the BiDiStream protocol, a
// BiDiStreamFile is returned. Otherwise a StreamFile is
// returned. Each Open() creates a new Reader or ReadWriter on
// the stream, and subsequent reads and writes on that fid operate
// on that Reader or ReadWriter. A Close() on a fid will close the
// Reader/ReadWriter.
func NewStreamFile(stat *proto.Stat, s Stream) File {
	if bidi, ok := s.(BiDiStream); ok {
		return &BiDiStreamFile{
			BaseFile:  NewBaseFile(stat),
			s:         bidi,
			fidReader: make(map[uint64]StreamReadWriter),
		}
	}
	return &StreamFile{
		BaseFile:  NewBaseFile(stat),
		s:         s,
		fidReader: make(map[uint64]StreamReader),
	}
}

func (f *StreamFile) Stat() proto.Stat {
	stat := f.fStat
	stat.Length = f.s.length()
	return stat
}

func (f *StreamFile) Open(fid uint64, omode proto.Mode) error {
	if omode == proto.Owrite ||
		omode == proto.Ordwr {
		return errors.New("Cannot open this stream for writing.")
	}
	f.fidReader[fid] = f.s.AddReader()
	return nil
}

func (f *StreamFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	bs := make([]byte, count)
	r, ok := f.fidReader[fid]
	if !ok {
		// This really shouldn't happen.
		return nil, fmt.Errorf("Failed to read stream. Not opened for read.")
	}
	n, err := r.Read(bs)
	if err != nil {
		return nil, err
	}
	bs = bs[:n]
	return bs, nil
}

func (f *StreamFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	return 0, errors.New("Cannot write to this stream.")
}

func (f *StreamFile) Close(fid uint64) error {
	r, ok := f.fidReader[fid]
	if ok {
		f.s.RemoveReader(r)
		delete(f.fidReader, fid)
	}
	return nil
}

func (f *BiDiStreamFile) Stat() proto.Stat {
	stat := f.fStat
	stat.Length = f.s.length()
	return stat
}

func (f *BiDiStreamFile) Open(fid uint64, omode proto.Mode) error {
	f.fidReader[fid] = f.s.AddReadWriter()
	return nil
}

func (f *BiDiStreamFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	bs := make([]byte, count)
	r, ok := f.fidReader[fid]
	if !ok {
		// This really shouldn't happen.
		return nil, fmt.Errorf("Failed to read stream. Server error.")
	}
	n, err := r.Read(bs)
	if err != nil {
		return nil, err
	}
	bs = bs[:n]
	return bs, nil
}

func (f *BiDiStreamFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	r, ok := f.fidReader[fid]
	if !ok {
		// This really shouldn't happen.
		return 0, fmt.Errorf("Failed to write stream. Server error.")
	}
	n, err := r.Write(data)
	return uint32(n), err
}

func (f *BiDiStreamFile) Close(fid uint64) error {
	r, ok := f.fidReader[fid]
	if ok {
		f.s.RemoveReader(r)
		delete(f.fidReader, fid)
	}
	return nil
}

type streamWithReader struct {
	s   BiDiStream
	srw StreamReadWriter
}

// PipeFile creates a new full-duplex stream for each call to Open()
// The stream will be passed as a BiDiStream to the handler function
// passed to NewPipeFile, which will be run in a new goroutine. When
// that function returns, the stream will be closed.
type PipeFile struct {
	*BaseFile
	fidReader map[uint64]streamWithReader
	handler   func(s BiDiStream)
}

// NewPipeFile creates a new PipeFile that will create a new bi-directional
// stream for every Open(), starting a goroutine calling handler on the stream.
// handler should loop to read the stream. When handler returns, the stream will
// be closed.
func NewPipeFile(stat *proto.Stat, handler func(s BiDiStream)) *PipeFile {
	return &PipeFile{
		BaseFile:  NewBaseFile(stat),
		fidReader: make(map[uint64]streamWithReader),
		handler:   handler,
	}
}

func (f *PipeFile) Open(fid uint64, omode proto.Mode) error {
	s := NewBlockingStream(10)
	go func() {
		defer s.Close()
		f.handler(s)
	}()
	f.fidReader[fid] = streamWithReader{s, s.AddReadWriter()}
	return nil
}

func (f *PipeFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	bs := make([]byte, count)
	s, ok := f.fidReader[fid]
	if !ok {
		// This really shouldn't happen.
		return nil, fmt.Errorf("Failed to read stream. Server error.")
	}
	n, err := s.srw.Read(bs)
	if err != nil {
		return nil, err
	}
	bs = bs[:n]
	return bs, nil
}

func (f *PipeFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	s, ok := f.fidReader[fid]
	if !ok {
		// This really shouldn't happen.
		return 0, fmt.Errorf("Failed to write stream. Server error.")
	}
	n, err := s.srw.Write(data)
	return uint32(n), err
}

func (f *PipeFile) Close(fid uint64) error {
	s, ok := f.fidReader[fid]
	if ok {
		s.s.Close()
		delete(f.fidReader, fid)
	}
	return nil
}
