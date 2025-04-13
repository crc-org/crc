package fs

import (
	"io"
	"log"
	"os"
	"sync"
	"time"
)

func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		// Need to do this for some reason.
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}

// A StreamReader reads from a Stream. See Stream.AddReader.
type StreamReader interface {
	Read(p []byte) (n int, err error)
	Close()
}

// A StreamReadWriter is just like a StreamReader, but allows the Reader to send data back to the Stream.
type StreamReadWriter interface {
	StreamReader
	Write(p []byte) (n int, err error)
}

type chanReader struct {
	read         chan []byte
	write        chan []byte
	writerClosed chan struct{}
	unread       []byte
	live         bool
}

func (r *chanReader) Read(p []byte) (n int, err error) {
	for len(p) > 0 {
		if len(r.unread) == 0 {
			if n > 0 {
				select {
				case bs, ok := <-r.read:
					if !ok {
						return
					}
					r.unread = bs
				default:
					return
				}
			} else {
				bs, ok := <-r.read
				if !ok {
					// Return 0, nil on EOF.
					// Returning io.EOF will cause RError response,
					// which is incorrect. 9p EOF is be 0-length RRead.
					return 0, nil
				}
				r.unread = bs
			}
		}
		newn := copy(p, r.unread)
		r.unread = r.unread[newn:]
		p = p[newn:]
		n += newn
		if len(p) == 0 {
			return
		}
	}
	return
}

func (r *chanReader) Write(p []byte) (n int, err error) {
	bs := make([]byte, len(p))
	copy(bs, p)
	select {
	case <-r.writerClosed:
		return 0, io.EOF
	default:
		select {
		case r.write <- bs:
			return len(p), nil
		case <-r.writerClosed:
			return 0, io.EOF
		}
	}
}

func (r *chanReader) Close() {
	if r.live {
		r.live = false
		close(r.read)
	}
}

// A Stream is a fan-out pipe of data. Stream implements io.Writer, and can have multiple readers
// associated with it, created with AddReader. How and whether data written with Write() is
// delivered to every reader is up to the implementation.
//
// For instance, when calling stream.Write([]byte("abc")) on a stream with 3 readers, it is equally
// valid for a stream to send "abc" to one and nothing to the others, or send "a" to one, "b" to
// the second, and "c" to the third, or to send "abc" to all three. See individual Stream
// implementations for their specific behavior.
type Stream interface {
	AddReader() StreamReader
	RemoveReader(r StreamReader)
	Write(p []byte) (n int, err error)
	Close() error
	length() uint64 // length of the stream (or 0 if unknown)
}

// A BiDiStream is a bi-directional stream. It adds two methods to the Stream interface,
// AddReadWriter and Read. BiDiStream allows Read()s of data written by StreamReadWriters. Although
// there is no way to determine which StreamReadWriter wrote the bytes read by Read(), it is
// guaranteed that a Write()s done by StreamReadWriters are delivered as complete chunks. That is,
// all bytes from a single StreamReadWriter.Write() call will be read by Read() before bytes from
// another StreamReader.Write() will be read.
type BiDiStream interface {
	Stream
	AddReadWriter() StreamReadWriter
	Read(p []byte) (n int, err error)
}

type baseStream struct {
	readers  []*chanReader
	read     []byte
	bufflen  int
	incoming chan []byte
	close    chan struct{}
	sync.Mutex
}

func (s *baseStream) length() uint64 {
	return 0
}

func (s *baseStream) AddReader() StreamReader {
	return s.AddReadWriter()
}

func (s *baseStream) AddReadWriter() StreamReadWriter {
	s.Lock()
	defer s.Unlock()
	reader := &chanReader{
		read:         make(chan []byte, s.bufflen),
		write:        s.incoming,
		writerClosed: s.close,
		live:         true,
	}
	if s.closed() {
		reader.Close()
	} else {
		s.readers = append(s.readers, reader)
	}
	return reader
}

func (s *baseStream) RemoveReader(r StreamReader) {
	s.Lock()
	defer s.Unlock()
	k := 0
	for i, reader := range s.readers {
		if r != reader {
			if i != k {
				s.readers[k] = reader
			}
			k++
		}
	}
	s.readers = s.readers[:k]
	r.Close()
}

func (s *baseStream) Read(p []byte) (n int, err error) {
	if s.read == nil || len(s.read) == 0 {
		for {
			var (
				in []byte
				ok bool
			)
			select {
			case in, ok = <-s.incoming:
			// Hacky way to ensure s.incoming gets preferentially selected.
			default:
				select {
				case in, ok = <-s.incoming:
				case <-s.close:
					return 0, io.EOF
				}
			}
			if !ok {
				return 0, io.EOF
			}
			s.read = in
			if len(s.read) > 0 {
				break
			}
		}
	}
	n = copy(p, s.read)
	s.read = s.read[n:]
	return
}

func (s *baseStream) closed() bool {
	select {
	case <-s.close:
		return true
	default:
		return false
	}
}

func (s *baseStream) Close() error {
	s.Lock()
	defer s.Unlock()
	if s.closed() {
		return nil
	}
	for _, reader := range s.readers {
		reader.Close()
	}
	s.readers = nil
	close(s.close)
	return nil
}

// DroppingStream is a stream which will disconnect readers who aren't able to keep up with the
// writer. The writer will slow down when writing (pausing for 50ms) to allow slower clients to
// catch up, but those that cannot keep up will be closed.
type DroppingStream struct {
	baseStream
}

func NewDroppingStream(buffer int) *DroppingStream {
	return &DroppingStream{
		baseStream{
			bufflen:  buffer,
			incoming: make(chan []byte, 10),
			close:    make(chan struct{}, 0),
		},
	}
}

func (s *DroppingStream) Write(p []byte) (n int, err error) {
	s.Lock()
	defer s.Unlock()
	k := 0
	t := time.NewTimer(10 * time.Millisecond)
	for i, reader := range s.readers {
		resetTimer(t, 50*time.Millisecond)
		cp := make([]byte, len(p))
		copy(cp, p)
		select {
		case reader.read <- cp:
			if i != k {
				s.readers[k] = reader
			}
			k++
		case <-t.C:
			// Writing to writer Timed out.
			reader.Close()
		}
	}
	s.readers = s.readers[:k]
	return len(p), nil
}

// A BlockingStream will ensure all data is written to active StreamReaders, or block if the
// readers are unable to receive the data. A certain number of bytes of data is buffered
// (see NewBlockingStream). If an unresponsive reader is removed, Write will be able to continue.
type BlockingStream struct {
	baseStream
	writeLock sync.Mutex
}

func NewBlockingStream(buffer int) *BlockingStream {
	return &BlockingStream{
		baseStream: baseStream{
			bufflen:  buffer,
			incoming: make(chan []byte, 10),
			close:    make(chan struct{}, 0),
		},
	}
}

func (s *BlockingStream) Write(p []byte) (n int, err error) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	laggers := s.tryWrite(s.readers, p)
	for len(laggers) > 0 {
		time.Sleep(10 * time.Millisecond)
		laggers = s.tryWrite(laggers, p)
	}
	return len(p), nil
}

func (s *BlockingStream) tryWrite(readers []*chanReader, p []byte) []*chanReader {
	s.Lock()
	defer s.Unlock()
	var laggers []*chanReader
	for _, reader := range readers {
		if !reader.live {
			continue
		}
		cp := make([]byte, len(p))
		copy(cp, p)
		select {
		case reader.read <- cp:
		default:
			laggers = append(laggers, reader)
		}
	}
	return laggers
}

// A SkippingStream Sends all Writes() to all StreamReaders, skipping those that are not able to
// keep up. Once a StreamReader starts reading again, it will get new Writes, having skipped some
// of the data. This may be good for things like event streams, audio, etc.
type SkippingStream struct {
	baseStream
}

func NewSkippingStream(buffer int) *SkippingStream {
	return &SkippingStream{
		baseStream{
			bufflen:  buffer,
			incoming: make(chan []byte, 10),
			close:    make(chan struct{}, 0),
		},
	}
}

func (s *SkippingStream) Write(p []byte) (n int, err error) {
	s.Lock()
	defer s.Unlock()
	t := time.NewTimer(50 * time.Millisecond)
	for _, reader := range s.readers {
		resetTimer(t, 50*time.Millisecond)
		cp := make([]byte, len(p))
		copy(cp, p)
		select {
		case reader.read <- cp:
		case <-t.C:
			// Timed out. Skip this message.
		}
	}
	return len(p), nil
}

type fileReader struct {
	f      *os.File
	signal chan struct{}
	live   bool
	t      *time.Timer
}

func (r *fileReader) Read(p []byte) (n int, err error) {
	if !r.live {
		return 0, io.EOF
	}

	for {
		n, err = r.f.Read(p)
		if err == nil || err != io.EOF {
			return
		}
		resetTimer(r.t, 500*time.Millisecond)
		select {
		case _, ok := <-r.signal:
			if !ok {
				r.Close()
				return 0, io.EOF
			}
		case <-r.t.C:
		}
	}
}

func (r *fileReader) Close() {
	if r.f != nil {
		r.f.Close()
		r.f = nil
	}
	r.live = false
}

// A SavedStream is a file-backed stream. Readers receive the full contents of the stream starting
// at the beginning and receive any new Writes.
type SavedStream struct {
	f       *os.File
	path    string
	readers []*fileReader
	closed  bool
	sync.Mutex
}

func NewSavedStream(path string) (*SavedStream, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return nil, err
	}
	return &SavedStream{
		f:      f,
		path:   path,
		closed: false,
	}, nil
}

func (s *SavedStream) length() uint64 {
	stat, err := s.f.Stat()
	if err != nil {
		log.Printf("fs streams: %s", err)
		return 0
	}
	return uint64(stat.Size())
}

func (s *SavedStream) AddReader() StreamReader {
	s.Lock()
	defer s.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		return &fileReader{
			f:      nil,
			signal: make(chan struct{}, 0),
			live:   false,
		}
	}

	reader := &fileReader{
		f:      f,
		signal: make(chan struct{}, 1),
		live:   true,
		t:      time.NewTimer(500 * time.Millisecond),
	}

	if s.closed {
		close(reader.signal)
	} else {
		s.readers = append(s.readers, reader)
	}
	return reader
}

func (s *SavedStream) RemoveReader(r StreamReader) {
	s.Lock()
	defer s.Unlock()
	k := 0
	for i, reader := range s.readers {
		if r != reader {
			if i != k {
				s.readers[k] = reader
			}
			k++
		}
	}
	s.readers = s.readers[:k]
	r.Close()
}

func (s *SavedStream) Write(p []byte) (n int, err error) {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return 0, io.EOF // TODO: Should this be EOF?
	}
	n, err = s.f.Write(p)
	if err != nil {
		return
	}
	k := 0
	for i, reader := range s.readers {
		if reader.live {
			select {
			case reader.signal <- struct{}{}:
			default:
			}
			if i != k {
				s.readers[k] = reader
			}
			k++
		}
	}
	s.readers = s.readers[:k]
	return
}

func (s *SavedStream) Close() error {
	s.Lock()
	defer s.Unlock()
	for _, reader := range s.readers {
		close(reader.signal)
	}
	s.closed = true
	s.readers = nil
	s.f.Close()
	s.f = nil
	return nil
}
