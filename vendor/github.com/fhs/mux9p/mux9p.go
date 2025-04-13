// Package mux9p implements a multiplexer for a 9P service.
//
// It is a port of Plan 9 Port's 9pserve program
// (https://9fans.github.io/plan9port/man/man4/9pserve.html)
// and can be used instead of 9pserve in a 9P server written in Go.
//
package mux9p

// Life cycle of a 9P message:
//
// 1. clientIO/processTx:
//		read from client into msg.tx
//		write global tags and fids to msg.tx
//		set msg.outc to a response channel
//		send msg to Config.outc
// 2. writeToServer:
//		receive from Config.outc
//		write msg.tx to server
// 3. readFromServer:
//		read from server into msg.rx
//		write client's tag to msg.rx
//		send msg to msg.outc
// 4. clientIO/processRx:
//		receive msg from the response channel
//		write msg.rx to client

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"

	"9fans.net/go/plan9"
	"github.com/fhs/mux9p/internal/p9p"
)

// Config contains options for the 9P multiplexer.
type Config struct {
	// No authentication. Respond to Tauth messages with an error.
	NoAuth bool

	// Logs are written here. It's set to the standard logger if nil.
	Logger *log.Logger

	// Sets the verbosity of the log.
	// Logs are not written if it's 0.
	// It can be overridden with environment variable verbose9pserve.
	LogLevel int

	// Options for the listener.
	ListenConfig net.ListenConfig

	msize     uint32 // 9P message size
	versioned bool   // Do not initialize the connection with a Tversion

	outc chan *msg

	fidtab  []*fid // global fids
	freefid []*fid
	msgtab  []*msg // msg indexed by global tag
	freemsg []*msg
	mu      sync.Mutex
}

type fid struct {
	fid  uint32 // global fid
	cfid uint32 // client's fid
}

type msg struct {
	ctag     uint16     // client's tag
	tag      uint16     // unique tag over all clients
	isopenfd bool       // Topenfd message
	tx, rx   *p9p.Fcall // transmit/receive 9P message
	outc     chan *msg  // server handler sends response to this channel
	fid      *fid       // Tattach, Twalk, etc.
	newfid   *fid       // Twalk Newfid
	afid     *fid       // Tauth Fid
	oldm     *msg       // msg corresponding to Tflush Oldtag
}

type client struct {
	tag  map[uint16]*msg // client tag → global tag
	fid  map[uint32]*fid // client fid → global fid
	outc chan *msg       // msg queued for write to 9P client
	cfg  *Config
}

// Listen creates a listener at the given network and address,
// accepts 9P clients from it and mutiplexes them into 9P server srv.
func Listen(network, address string, srv io.ReadWriter, cfg *Config) error {
	return ListenContext(context.Background(), network, address, srv, cfg)
}

// ListenContext is equivalent to Listen but with a context.
func ListenContext(ctx context.Context, network, address string, srv io.ReadWriter, cfg *Config) error {
	ln, err := cfg.ListenConfig.Listen(ctx, network, address)
	if err != nil && network == "unix" && isAddrInUse(err) {
		var d net.Dialer
		if _, err1 := d.DialContext(ctx, network, address); !isConnRefused(err1) {
			return err // Listen error
		}
		// Dead socket, so remove it.
		err = os.Remove(address)
		if err != nil {
			return err
		}
		ln, err = cfg.ListenConfig.Listen(ctx, network, address)
	}
	if err != nil {
		return err
	}
	defer ln.Close()

	return Do(ln, srv, cfg)
}

// Do accepts 9P clients from listener ln and mutiplexes them into 9P server srv.
func Do(ln net.Listener, srv io.ReadWriter, cfg *Config) error {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.Logger == nil {
		cfg.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	if x := os.Getenv("verbose9pserve"); x != "" {
		n, err := strconv.Atoi(x)
		if err == nil {
			cfg.LogLevel = n
			fmt.Fprintf(os.Stderr, "verbose9pserve %s => %d\n", x, cfg.LogLevel)
		}
	}
	cfg.msize = 8092
	return cfg.mainproc(srv, ln)
}

func (cfg *Config) mainproc(srv io.ReadWriter, ln net.Listener) error {
	cfg.log("9pserve running\n")

	if !cfg.versioned {
		f := &plan9.Fcall{
			Type:    plan9.Tversion,
			Version: "9P2000",
			Msize:   cfg.msize,
			Tag:     plan9.NOTAG,
		}
		vbuf, err := f.Bytes()
		if err != nil {
			cfg.log("Fcall conversion to bytes failed: %v", err)
			return err
		}
		cfg.log2("init: * <- %v\n", f)
		_, err = srv.Write(vbuf)
		if err != nil {
			cfg.log("error writing Tversion: %v", err)
			return err
		}
		f, err = plan9.ReadFcall(srv)
		if err != nil {
			cfg.log("ReadFcall failed: %v", err)
			return err
		}
		if f.Msize < cfg.msize {
			cfg.msize = f.Msize
		}
		cfg.log2("init: * -> %v\n", f)
	}

	cfg.outc = make(chan *msg)
	go cfg.readFromServer(srv)
	go cfg.writeToServer(srv)

	return cfg.listenthread(ln)
}

func (cfg *Config) listenthread(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			cfg.log("listen: %v\n", err)
			return err
		}
		cfg.log("incoming call on %v\n", conn.LocalAddr())
		go cfg.clientIO(conn)
	}
}

func (cfg *Config) clientIO(conn net.Conn) {
	readc := make(chan *p9p.Fcall)
	go func() {
		for {
			f, err := p9p.ReadFcall(conn)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("ReadFcall failed: %v", err)
				break
			}
			cfg.log2("fd#%v -> %v\n", conn.RemoteAddr(), f)
			readc <- f
		}
		close(readc)
	}()

	c := &client{
		tag:  make(map[uint16]*msg),
		fid:  make(map[uint32]*fid),
		outc: make(chan *msg, 1),
		cfg:  cfg,
	}

	for reading, writing := true, true; reading && writing; {
		select {
		case f := <-readc:
			if f == nil { // EOF or error
				reading = false
				readc = nil
				// Ask readFromServer to close c.outc.
				cfg.outc <- &msg{tx: nil, rx: nil, outc: c.outc}
				continue
			}
			m := cfg.msgnew()
			m.tx = f
			m.ctag = m.tx.Tag
			m.outc = c.outc
			cfg.log2("fd#%v: new msg %p\n", conn.RemoteAddr(), m)
			if _, ok := c.tag[m.tx.Tag]; ok {
				c.send9pError(m, "duplicate tag")
				continue
			}
			c.tag[m.tx.Tag] = m

			_, allowOpenfd := conn.(*net.UnixConn)
			c.processTx(m, allowOpenfd)

		case m := <-c.outc:
			if m == nil {
				writing = false
				c.outc = nil
				continue
			}
			c.processRx(conn, m)
		}
	}
	c.cleanup()
}

func (c *client) send9pError(m *msg, ename string) {
	m.rx.Type = plan9.Rerror
	m.rx.Ename = ename
	m.rx.Tag = m.tx.Tag
	c.outc <- m
}

func (c *client) processTx(m *msg, allowOpenfd bool) {
	var ok bool
	cfg := c.cfg

	switch m.tx.Type {
	default:
		cfg.log("unknown fcall type %v", m.tx.Type)

	case plan9.Tversion:
		m.rx = &p9p.Fcall{}
		m.rx.Tag = m.tx.Tag
		m.rx.Msize = m.tx.Msize
		if m.rx.Msize > cfg.msize {
			m.rx.Msize = cfg.msize
		}
		m.rx.Version = "9P2000"
		m.rx.Type = plan9.Rversion
		c.outc <- m
		return

	case plan9.Tflush:
		m.oldm, ok = c.tag[m.tx.Oldtag]
		if !ok {
			m.rx = &p9p.Fcall{}
			m.rx.Tag = m.tx.Tag
			m.rx.Type = plan9.Rflush
			c.outc <- m
			return
		}

	case plan9.Tattach:
		m.afid = nil
		if m.tx.Afid != plan9.NOFID {
			m.afid, ok = c.fid[m.tx.Afid]
			if !ok {
				c.send9pError(m, "unknown fid")
				return
			}
		}
		m.fid = cfg.fidnew(m.tx.Fid)
		if _, ok := c.fid[m.tx.Fid]; ok {
			c.send9pError(m, "duplicate fid")
			return
		}
		c.fid[m.tx.Fid] = m.fid

	case plan9.Twalk:
		m.fid, ok = c.fid[m.tx.Fid]
		if !ok {
			c.send9pError(m, "unknown fid")
			return
		}
		if m.tx.Newfid == m.tx.Fid {
			m.newfid = m.fid
		} else {
			m.newfid = cfg.fidnew(m.tx.Newfid)
			if _, ok := c.fid[m.tx.Newfid]; ok {
				c.send9pError(m, "duplicate fid")
				return
			}
			c.fid[m.tx.Newfid] = m.newfid
		}

	case plan9.Tauth:
		if cfg.NoAuth {
			c.send9pError(m, "authentication rejected")
			return
		}
		m.afid = cfg.fidnew(m.tx.Afid)
		if _, ok := c.fid[m.tx.Afid]; ok {
			c.send9pError(m, "duplicate fid")
			return
		}
		c.fid[m.tx.Afid] = m.afid

	case p9p.Topenfd:
		if !allowOpenfd {
			c.send9pError(m, "only supported on unix socket")
			return
		}
		if m.tx.Mode&^(plan9.OTRUNC|3) != 0 {
			c.send9pError(m, "bad openfd mode")
			return
		}
		m.isopenfd = true
		m.tx.Type = plan9.Topen
		m.fid, ok = c.fid[m.tx.Fid]
		if !ok {
			c.send9pError(m, "unknown fid")
			return
		}

	case plan9.Tcreate:
		if m.tx.Perm&(plan9.DMSYMLINK|plan9.DMDEVICE|plan9.DMNAMEDPIPE|plan9.DMSOCKET) != 0 {
			c.send9pError(m, "unsupported file type")
			return
		}
		fallthrough
	case plan9.Topen, plan9.Tclunk, plan9.Tread, plan9.Twrite, plan9.Tremove, plan9.Tstat, plan9.Twstat:
		m.fid, ok = c.fid[m.tx.Fid]
		if !ok {
			c.send9pError(m, "unknown fid")
			return
		}
	}

	// have everything - translate and send
	m.ctag = m.tx.Tag
	m.tx.Tag = m.tag
	if m.fid != nil {
		m.tx.Fid = m.fid.fid
	}
	if m.newfid != nil {
		m.tx.Newfid = m.newfid.fid
	}
	if m.afid != nil {
		m.tx.Afid = m.afid.fid
	}
	if m.oldm != nil {
		m.tx.Oldtag = m.oldm.tag
	}
	cfg.outc <- m
}

func (c *client) cleanup() {
	cfg := c.cfg

	internalc := make(chan *msg)

	// flush all outstanding messages
	for _, om := range c.tag {
		m := cfg.msgnew()
		m.tx = &p9p.Fcall{}
		m.tx.Type = plan9.Tflush
		m.tx.Tag = m.tag
		m.tx.Oldtag = om.tag
		m.oldm = om
		m.outc = internalc
		cfg.outc <- m
		mm := <-internalc
		assert(mm == m)
		cfg.msgput(m) // got from msgnew
		if c.deleteTag(om.ctag, om) {
			cfg.msgput(om)
		}
	}

	// clunk all outstanding fids
	for _, f := range c.fid {
		m := cfg.msgnew()
		m.tx = &p9p.Fcall{}
		m.tx.Type = plan9.Tclunk
		m.tx.Tag = m.tag
		m.tx.Fid = f.fid
		m.fid = f
		m.outc = internalc
		cfg.outc <- m
		mm := <-internalc
		assert(mm == m)
		cfg.msgput(m) // got from msgnew
		if c.deleteFid(m.fid.cfid, f) {
			cfg.fidput(f)
		}
	}

	assert(len(c.tag) == 0)
	assert(len(c.fid) == 0)
}

// writeToPipe copies data from fid to pipe.
func (cfg *Config) writeToPipe(fid uint32, pipe io.Writer) {
	ch := make(chan *msg)
	tot := uint64(0)
	for {
		cfg.log("tread...")
		m := cfg.msgnew()
		m.tx = &p9p.Fcall{}
		m.tx.Type = plan9.Tread
		m.tx.Count = cfg.msize - plan9.IOHDRSZ
		m.tx.Fid = fid
		m.tx.Tag = m.tag
		m.tx.Offset = tot
		m.outc = ch
		cfg.outc <- m
		m = <-ch
		if m.rx.Type == plan9.Rerror {
			cfg.log("read error: %s\n", m.rx.Ename)
			break
		}
		if m.rx.Count == 0 {
			break
		}
		tot += uint64(m.rx.Count)
		if _, err := pipe.Write(m.rx.Data[:m.rx.Count]); err != nil {
			cfg.log("pipe write error: %v\n", err)
			break
		}
		cfg.msgput(m)
	}
}

func (cfg *Config) readFromPipe(fid uint32, pipe io.Reader) {
	var buf [1024]byte

	ch := make(chan *msg)
	tot := uint64(0)
	for {
		cfg.log("twrite...")
		n := len(buf)
		if n > int(cfg.msize) {
			n = int(cfg.msize)
		}
		cfg.log("openfd reading %v bytes...", n)
		n, err := pipe.Read(buf[:n])
		if err == io.EOF {
			break
		}
		if err != nil {
			cfg.log("pipe read error: %v\n", err)
		}
		cfg.log("openfd read: %v", string(buf[:n]))
		m := cfg.msgnew()
		m.tx = &p9p.Fcall{}
		m.tx.Type = plan9.Twrite
		m.tx.Fid = fid
		m.tx.Data = buf[:]
		m.tx.Count = uint32(n)
		m.tx.Tag = m.tag
		m.tx.Offset = tot
		m.outc = ch
		ch <- m
		m = <-ch
		if m.rx.Type == plan9.Rerror {
			cfg.log("write error: %s\n", m.rx.Ename)
		}
		tot += uint64(n)
		cfg.msgput(m)
	}
}

func (cfg *Config) pipeIO(mode uint8, fdfid *fid, pipe *os.File) {
	var wg sync.WaitGroup

	mo := mode & 3
	if mo == plan9.OREAD || mo == plan9.ORDWR {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg.writeToPipe(fdfid.fid, pipe)
		}()
	}
	if mo == plan9.OWRITE || mo == plan9.ORDWR {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg.readFromPipe(fdfid.fid, pipe)
		}()
	}

	wg.Wait()

	cfg.log("eof on %d fid %d\n", pipe.Fd(), fdfid.fid)
	pipe.Close()

	cfg.log("eof on %d fid %d\n", pipe.Fd(), fdfid.fid)
	ch := make(chan *msg)
	m := cfg.msgnew()
	m.tx = &p9p.Fcall{}
	m.tx.Type = plan9.Tclunk
	m.tx.Tag = m.tag
	m.tx.Fid = fdfid.fid
	m.fid = fdfid
	m.outc = ch
	cfg.outc <- m
	m = <-ch
	cfg.msgput(m)
	cfg.fidput(m.fid)
}

func (c *client) xopenfd(conn net.Conn, m *msg) error {
	cfg := c.cfg

	p0, p1, err := p9p.Pipe()
	if err != nil {
		return err
	}
	cfg.log("xopen pipe %d %d...", p0.Fd(), p1.Fd())

	// now we're committed.

	go cfg.pipeIO(m.tx.Mode, m.fid, p0)

	// rewrite as Ropenfd
	m.rx.Type = p9p.Ropenfd
	m.rx.Unixfd = uint32(p1.Fd())
	return nil
}

func (c *client) processRx(conn net.Conn, m *msg) {
	cfg := c.cfg

	badType := m.tx.Type+1 != m.rx.Type
	if !badType && m.isopenfd {
		if err := c.xopenfd(conn, m); err != nil {
			c.send9pError(m, err.Error())
			return
		}
		// Client considers the fid to be clunked.
		// Delete fid from client but keep in the server.
		c.deleteFid(m.fid.cfid, m.fid)
	}
	switch m.tx.Type {
	case plan9.Tflush:
		om := m.oldm
		if om != nil {
			if c.deleteTag(om.ctag, om) {
				cfg.msgput(om)
			}
		}

	case plan9.Tclunk, plan9.Tremove:
		if m.fid != nil {
			if c.deleteFid(m.fid.cfid, m.fid) {
				cfg.fidput(m.fid)
			}
		}

	case plan9.Tauth:
		if badType && m.afid != nil {
			cfg.log("auth error\n")
			if c.deleteFid(m.afid.cfid, m.afid) {
				cfg.fidput(m.afid)
			}
		}

	case plan9.Tattach:
		if badType && m.fid != nil {
			if c.deleteFid(m.fid.cfid, m.fid) {
				cfg.fidput(m.fid)
			}
		}

	case plan9.Twalk:
		if badType || len(m.rx.Wqid) < len(m.tx.Wname) {
			if m.tx.Fid != m.tx.Newfid && m.newfid != nil {
				if c.deleteFid(m.newfid.cfid, m.newfid) {
					cfg.fidput(m.newfid)
				}
			}
		}

	case plan9.Tread:
	case plan9.Tstat:
	case plan9.Topen:
	case plan9.Tcreate:
	}
	cfg.log2("fd#%v <- %v\n", conn.RemoteAddr(), m.rx)
	rpkt, err := m.rx.Bytes()
	if err != nil {
		log.Fatalf("failed to convert Fcall to bytes: %v\n", err)
	}
	if _, err := conn.Write(rpkt); err != nil {
		cfg.log("write error: %v\n", err)
	}
	if m.rx.Type == p9p.Ropenfd {
		if err := p9p.SendFD(conn.(*net.UnixConn), uintptr(m.rx.Unixfd)); err != nil {
			cfg.log("sendfd failed: %v\n", err)
		}
	}
	if c.deleteTag(m.ctag, m) {
		cfg.msgput(m)
	}
}

func (cfg *Config) writeToServer(srv io.ReadWriter) {
	for {
		m := <-cfg.outc
		if m == nil { // all clients have disconnected
			break
		}
		if m.tx == nil {
			// The client for this message has closed the connection.
			close(m.outc)
			continue
		}
		cfg.log2("* <- %v\n", m.tx)
		tpkt, err := m.tx.Bytes()
		if err != nil {
			log.Fatalf("failed to convert Fcall to bytes: %v\n", err)
		}
		if _, err := srv.Write(tpkt); err != nil {
			log.Fatalf("output error: %s\n", err)
		}
	}
}

func (cfg *Config) readFromServer(srv io.ReadWriter) {
	cfg.log("input thread\n")

	for {
		f, err := p9p.ReadFcall(srv)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("ReadFcall failed: %v", err)
		}
		m := cfg.msgget(int(f.Tag))
		if m == nil {
			log.Printf("unexpected 9P response tag %v\n", f.Tag)
			continue
		}
		m.rx = f
		cfg.log2("* -> %v\n", m.rx)
		m.rx.Tag = m.ctag
		m.outc <- m
	}
}

func (cfg *Config) fidnew(cfid uint32) *fid {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if len(cfg.freefid) > 0 {
		n := len(cfg.freefid) - 1
		f := cfg.freefid[n]
		cfg.freefid = cfg.freefid[:n]

		// clear everything except global fid and cfid
		*f = fid{
			fid:  f.fid,
			cfid: cfid,
		}
		return f
	}

	f := &fid{
		fid:  uint32(len(cfg.fidtab)),
		cfid: cfid,
	}
	cfg.fidtab = append(cfg.fidtab, f)
	return f
}

func (cfg *Config) fidput(f *fid) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.freefid = append(cfg.freefid, f)
}

func (cfg *Config) msgnew() *msg {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if len(cfg.freemsg) > 0 {
		n := len(cfg.freemsg) - 1
		m := cfg.freemsg[n]
		cfg.freemsg = cfg.freemsg[:n]

		// clear everything except the tag
		*m = msg{
			tag: m.tag,
		}
		return m
	}
	m := &msg{
		tag: uint16(len(cfg.msgtab)),
	}
	cfg.msgtab = append(cfg.msgtab, m)
	cfg.log2("msgnew %p tag %d\n", m, m.tag)
	return m
}

func (cfg *Config) msgput(m *msg) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.log2("msgput %p tag %d/%d\n", m, m.tag, m.ctag)
	cfg.freemsg = append(cfg.freemsg, m)
}

func (cfg *Config) msgget(n int) *msg {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if n < 0 || n >= len(cfg.msgtab) {
		return nil
	}
	m := cfg.msgtab[n]
	return m
}

func (cfg *Config) log(format string, a ...interface{}) {
	if cfg.LogLevel > 0 {
		cfg.Logger.Printf(format, a...)
	}
}

func (cfg *Config) log2(format string, a ...interface{}) {
	if cfg.LogLevel > 1 {
		cfg.Logger.Printf(format, a...)
	}
}

func assert(b bool) {
	if !b {
		panic("assert failed")
	}
}

func (c *client) deleteTag(tag uint16, m *msg) bool {
	if m1, ok := c.tag[tag]; ok {
		if m1 != m {
			c.cfg.log("deleteTag %d got %p want %p\n", tag, m1, m)
		}
		delete(c.tag, tag)
		return true
	}
	return false
}

func (c *client) deleteFid(fid uint32, f *fid) bool {
	if f1, ok := c.fid[fid]; ok {
		if f1 != f {
			c.cfg.log("deleteFid %d got %p want %p\n", fid, f1, f)
		}
		delete(c.fid, fid)
		return true
	}
	return false
}

func isAddrInUse(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		if err, ok := err.Err.(*os.SyscallError); ok {
			return err.Err == syscall.EADDRINUSE
		}
	}
	return false
}

func isConnRefused(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		if err, ok := err.Err.(*os.SyscallError); ok {
			return err.Err == syscall.ECONNREFUSED
		}
	}
	return false
}
