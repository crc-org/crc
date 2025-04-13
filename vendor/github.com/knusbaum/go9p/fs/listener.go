package fs

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/knusbaum/go9p/proto"
)

// ListenFile implements a net.Listener as a 9p File.
// The file is a stream, so offsets for Read and Write are ignored.
type ListenFile struct {
	BaseFile
	closed   bool
	conns    map[uint64]*fileConn
	incoming chan *fileConn
	m        sync.RWMutex
}

type ListenFileListener ListenFile

type fileConn struct {
	fid    uint64
	closed bool

	outReader *io.PipeReader
	outWriter *io.PipeWriter
	inReader  *io.PipeReader
	inWriter  *io.PipeWriter

	m sync.Mutex
}

type addr9p struct {
	resource string
}

var _ File = &ListenFile{}
var _ net.Listener = &ListenFileListener{}
var _ net.Conn = &fileConn{}
var _ net.Addr = &addr9p{}

func NewListenFile(s *proto.Stat) *ListenFile {
	return &ListenFile{
		BaseFile: BaseFile{fStat: *s},
		conns:    make(map[uint64]*fileConn),
		incoming: make(chan *fileConn, 10),
	}
}

func (f *ListenFile) connForFid(fid uint64) *fileConn {
	f.m.RLock()
	defer f.m.RUnlock()
	return f.conns[fid]
}

func (f *ListenFile) Open(fid uint64, omode proto.Mode) error {
	f.m.Lock()
	log.Printf("OPEN [%d]\n", fid)
	if f.closed {
		f.m.Unlock()
		return fmt.Errorf("Server closed the connection.")
	}
	outReader, outWriter := io.Pipe()
	inReader, inWriter := io.Pipe()
	conn := &fileConn{
		fid:       fid,
		outReader: outReader,
		outWriter: outWriter,
		inReader:  inReader,
		inWriter:  inWriter,
	}
	f.conns[fid] = conn
	f.m.Unlock()
	f.incoming <- conn
	return nil
}

func (f *ListenFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	//log.Printf("READ [%d]", fid)
	conn := f.connForFid(fid)
	if conn == nil {
		return nil, fmt.Errorf("Bad FID")
	}
	return conn.handleRead(count)
}

func (f *ListenFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	conn := f.connForFid(fid)
	if conn == nil {
		return 0, fmt.Errorf("Bad FID")
	}
	return conn.handleWrite(data)
}

func (f *ListenFile) Close(fid uint64) error {
	f.m.Lock()
	defer f.m.Unlock()
	conn := f.conns[fid]
	if conn == nil {
		return fmt.Errorf("Bad FID.")
	}
	delete(f.conns, fid)
	return conn.Close()
}

// Accept waits for and returns the next connection to the listener.
func (l *ListenFileListener) Accept() (net.Conn, error) {
	conn, ok := <-l.incoming
	if !ok {
		return nil, fmt.Errorf("Connection Closed")
	}
	return conn, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *ListenFileListener) Close() error {
	l.m.Lock()
	defer l.m.Unlock()
	l.closed = true
	for {
		select {
		case conn := <-l.incoming:
			conn.Close()
		default:
			break
		}
	}
	close(l.incoming)
	for _, conn := range l.conns {
		conn.Close()
	}
	l.conns = nil
	return nil
}

// Addr returns the listener's network address.
func (l *ListenFileListener) Addr() net.Addr {
	return &addr9p{"TODO"}
}

func (c *fileConn) Close() error {
	c.m.Lock()
	defer c.m.Unlock()
	c.closed = true
	c.inReader.CloseWithError(io.EOF)
	c.outWriter.CloseWithError(io.EOF)
	c.inWriter.CloseWithError(io.EOF)
	c.outReader.CloseWithError(io.EOF)
	return nil
}

func (c *fileConn) isClosed() bool {
	c.m.Lock()
	defer c.m.Unlock()
	return c.closed
}

func (c *fileConn) LocalAddr() net.Addr {
	return &addr9p{"TODO"}
}

func (c *fileConn) RemoteAddr() net.Addr {
	return &addr9p{"TODO"}
}

func (c *fileConn) SetDeadline(t time.Time) error {
	return fmt.Errorf("TODO")
}

func (c *fileConn) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("TODO")
}

func (c *fileConn) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("TODO")
}

func (c *fileConn) handleRead(count uint64) ([]byte, error) {
	buff := make([]byte, count)
	n, err := c.outReader.Read(buff)
	if err == io.EOF {
		err = nil
	}
	return buff[:n], err
}

func (c *fileConn) handleWrite(data []byte) (uint32, error) {
	n, err := c.inWriter.Write(data)
	return uint32(n), err
}

func (c *fileConn) Read(p []byte) (int, error) {
	//return c.inReader.Read(p)
	n, err := c.inReader.Read(p)
	if err == io.ErrClosedPipe {
		err = io.EOF
	}
	return n, err
}

func (c *fileConn) Write(p []byte) (int, error) {
	n, err := c.outWriter.Write(p)
	if err == io.ErrClosedPipe {
		err = io.EOF
	}
	return n, err
}

func (a *addr9p) Network() string {
	return "9p"
}

func (a *addr9p) String() string {
	return a.resource
}
