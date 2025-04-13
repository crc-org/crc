// Package go9p contains contains an interface definition for a 9p2000 server, `Srv`.
// along with a few functions that will serve the 9p2000 protocol using a `Srv`.
//
// Most people wanting to implement a 9p filesystem should start in the subpackage
// github.com/knusbaum/go9p/fs, which contains tools for constructing a file system
// which can be served using the functions in this package.
//
// The subpackage github.com/knusbaum/go9p/proto contains the protocol implementation.
// It is used by the other packages to send and receive 9p2000 messages. It may be
// useful to someone who wants to investigate 9p2000 at the protocol level.
package go9p

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"

	"github.com/knusbaum/go9p/proto"
)

// If Verbose is true, incoming and outgoing 9p messages will be printed to stderr.
var Verbose bool

func verboseLog(msg string, args ...interface{}) {
	if Verbose {
		log.Printf(msg, args...)
	}
}

// The Srv interface is used to handle 9p2000 messages.
// Each function handles a specific type of message, and
// should return a response. If some expected error occurs,
// for example a TOpen message for a file with the wrong
// permissions, a proto.TError message should be returned
// rather than a go error. Returning a go error indicates that
// something has gone wrong with the server, and when used with
// Serve and PostSrv, will cause the connection to be terminated
// or the file descriptor to be closed.
type Srv interface {
	NewConn() Conn
	Version(Conn, *proto.TRVersion) (proto.FCall, error)
	Auth(Conn, *proto.TAuth) (proto.FCall, error)
	Attach(Conn, *proto.TAttach) (proto.FCall, error)
	Walk(Conn, *proto.TWalk) (proto.FCall, error)
	Open(Conn, *proto.TOpen) (proto.FCall, error)
	Create(Conn, *proto.TCreate) (proto.FCall, error)
	Read(Conn, *proto.TRead) (proto.FCall, error)
	Write(Conn, *proto.TWrite) (proto.FCall, error)
	Clunk(Conn, *proto.TClunk) (proto.FCall, error)
	Remove(Conn, *proto.TRemove) (proto.FCall, error)
	Stat(Conn, *proto.TStat) (proto.FCall, error)
	Wstat(Conn, *proto.TWstat) (proto.FCall, error)
}

// Conn represents an individual connection to a 9p server.
// In the case of a server listening on a network, there
// may be many clients connected to a given server at once.
type Conn interface {
	TagContext(uint16) context.Context
	DropContext(uint16)
}

func handleConnection(nc net.Conn, srv Srv) {
	defer nc.Close()
	read := bufio.NewReader(nc)
	err := handleIOAsync(read, nc, srv)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

// handleIO seems to be about 10x faster than handleIOAsync
// in my experiments. It would be nice to be able to keep some
// performance without making the reading, handling, and
// writing of calls synchronous.
func handleIO(r io.Reader, w io.Writer, srv Srv) error {
	conn := srv.NewConn()
	for {
		call, err := proto.ParseCall(r)
		if err != nil {
			return err
		}
		verboseLog("=in=> %s\n", call)
		resp, err := handleCall(call, srv, conn)
		if err != nil {
			return err
		}

		if resp == nil {
			// This case happens when an active tag is
			// flushed.
			continue
		}
		verboseLog("<=out= %s\n", resp)
		_, err = w.Write(resp.Compose())
		if err != nil {
			return err
		}
	}
	return nil
}

func handleIOAsync(r io.Reader, w io.Writer, srv Srv) error {
	incoming := make(chan proto.FCall, 100)
	outgoing := make(chan proto.FCall, 100)

	conn := srv.NewConn()

	// Write the outgoing
	var outgoingWG sync.WaitGroup
	defer func() { outgoingWG.Wait() }()
	outgoingWG.Add(1)
	go func() {
		outgoingWG.Done()
		for call := range outgoing {
			verboseLog("<=out= %s\n", call)
			_, err := w.Write(call.Compose())
			if err != nil {
				log.Printf("Protocol error: %v\n", err)
			}
		}
	}()

	var workerWG sync.WaitGroup
	defer func() { workerWG.Wait(); close(outgoing) }()
	for i := 0; i < 100; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			for call := range incoming {
				resp, err := handleCall(call, srv, conn)
				if err != nil {
					log.Printf("Protocol error: %v\n", err)
					//return err
					return
				}
				if resp == nil {
					// This case happens when an active tag is
					// flushed.
					continue
				}
				outgoing <- resp
			}
		}()
	}

	// Read incoming
	defer close(incoming)
	for {
		call, err := proto.ParseCall(r)
		verboseLog("=in=> %s\n", call)
		if err != nil {
			log.Printf("Protocol error: %v\n", err)
			return err
		}
		select {
		case incoming <- call:
		default:
			panic("FAILED TO QUEUE INCOMING!")
		}
	}
	return nil
}

func handleCall(call proto.FCall, srv Srv, conn Conn) (proto.FCall, error) {
	ctx := conn.TagContext(call.GetTag())
	var (
		ret proto.FCall
		err error
	)
	switch call.(type) {
	case *proto.TRVersion:
		ret, err = srv.Version(conn, call.(*proto.TRVersion))
	case *proto.TAuth:
		ret, err = srv.Auth(conn, call.(*proto.TAuth))
	case *proto.TAttach:
		ret, err = srv.Attach(conn, call.(*proto.TAttach))
	case *proto.TFlush:
		flush := call.(*proto.TFlush)
		//conn.DropContext(flush.Oldtag)
		ret, err = &proto.RFlush{proto.Header{proto.Rflush, flush.Tag}}, nil
	case *proto.TWalk:
		ret, err = srv.Walk(conn, call.(*proto.TWalk))
	case *proto.TOpen:
		ret, err = srv.Open(conn, call.(*proto.TOpen))
	case *proto.TCreate:
		ret, err = srv.Create(conn, call.(*proto.TCreate))
	case *proto.TRead:
		ret, err = srv.Read(conn, call.(*proto.TRead))
	case *proto.TWrite:
		ret, err = srv.Write(conn, call.(*proto.TWrite))
	case *proto.TClunk:
		ret, err = srv.Clunk(conn, call.(*proto.TClunk))
	case *proto.TRemove:
		ret, err = srv.Remove(conn, call.(*proto.TRemove))
	case *proto.TStat:
		ret, err = srv.Stat(conn, call.(*proto.TStat))
	case *proto.TWstat:
		ret, err = srv.Wstat(conn, call.(*proto.TWstat))
	default:
		return nil, fmt.Errorf("Invalid call: %s", reflect.TypeOf(call))
	}

	if ctx.Err() != nil {
		return nil, nil
	}
	conn.DropContext(call.GetTag())
	return ret, err
}

// ServeReadWriter accepts an io.Reader an io.Writer, and an Srv.
// It reads 9p2000 messages from r, handles them with srv, and
// writes the responses to w.
func ServeReadWriter(r io.Reader, w io.Writer, srv Srv) error {
	return handleIOAsync(r, w, srv)
}

// Serve serves srv on the given address, addr.
func Serve(addr string, srv Srv) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		a, err := l.Accept()
		if err != nil {
			return err
		}
		go func(nc net.Conn, srv Srv) {
			defer nc.Close()
			read := bufio.NewReader(nc)
			err := ServeReadWriter(read, nc, srv)
			if err != nil {
				log.Printf("%v\n", err)
			}
		}(a, srv)
	}
}

// PostSrv serves srv, from a file descriptor named name.
// The fd is posted and can subsequently be mounted. On Unix, the
// descriptor is posted under in the current namespace, which is
// determined by 9fans.net/go/plan9/client Namespace. On Plan9 it
// is posted in the usual place, /srv.
func PostSrv(name string, srv Srv) error {
	f, handle, err := postfd(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if handle != nil {
		defer handle.Close()
	}
	err = ServeReadWriter(f, f, srv)
	return err
}
