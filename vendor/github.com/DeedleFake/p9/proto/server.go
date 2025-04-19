package proto

import (
	"io"
	"log"
	"net"
	"sync"
)

// Serve serves a server for the given Proto, listening for new
// connection on lis and handling them using the provided handler.
//
// Note that to avoid a data race, messages from a single client are
// handled entirely sequentially until an msize has been established,
// at which point they will be handled concurrently. An msize is
// established when a handler returns a Msizer.
func Serve(lis net.Listener, p Proto, connHandler ConnHandler) (err error) {
	for {
		c, err := lis.Accept()
		if err != nil {
			return err
		}

		go func() {
			defer c.Close()

			if h, ok := connHandler.(handleConn); ok {
				h.HandleConn(c)
			}
			if h, ok := connHandler.(handleDisconnect); ok {
				defer h.HandleDisconnect(c)
			}

			mh := connHandler.MessageHandler()
			if c, ok := mh.(io.Closer); ok {
				defer c.Close()
			}

			handleMessages(c, p, mh)
		}()
	}
}

// ListenAndServe is a convenience function that establishes a
// listener, via net.Listen, and then calls Serve.
func ListenAndServe(network, addr string, p Proto, connHandler ConnHandler) (rerr error) {
	lis, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	defer func() {
		err := lis.Close()
		if (err != nil) && (rerr == nil) {
			rerr = err
		}
	}()

	return Serve(lis, p, connHandler)
}

func handleMessages(c net.Conn, p Proto, handler MessageHandler) {
	var setter sync.Once

	var msize uint32
	mode := func(f func()) {
		f()
	}

	for {
		tmsg, tag, err := p.Receive(c, msize)
		if err != nil {
			if err == io.EOF {
				return
			}

			log.Printf("Error reading message: %v", err)
		}

		mode(func() {
			rmsg := handler.HandleMessage(tmsg)
			if rmsg, ok := rmsg.(Msizer); ok {
				if msize > 0 {
					log.Println("Warning: Attempted to set msize twice.")
				}

				setter.Do(func() {
					msize = rmsg.P9Msize()
					mode = func(f func()) {
						go f()
					}
				})
			}

			err := p.Send(c, tag, rmsg)
			if err != nil {
				log.Printf("Error writing message: %v", err)
			}
		})
	}
}

// ConnHandler initializes new MessageHandlers for incoming
// connections. Unlike HTTP, which is a connectionless protocol, 9P
// and related protocols require that each connection be handled as a
// unique client session with a stored state, hence this two-step
// process.
//
// If a ConnHandler provides a HandleConn(net.Conn) method, that
// method will be called when a new connection is made. Similarly, if
// it provides a HandleDisconnect(net.Conn) method, that method will
// be called when a connection is ended.
type ConnHandler interface {
	MessageHandler() MessageHandler
}

type handleConn interface {
	HandleConn(c net.Conn)
}

type handleDisconnect interface {
	HandleDisconnect(c net.Conn)
}

// ConnHandlerFunc allows a function to be used as a ConnHandler.
type ConnHandlerFunc func() MessageHandler

func (h ConnHandlerFunc) MessageHandler() MessageHandler {
	return h()
}

// MessageHandler handles messages for a single client connection.
//
// If a MessageHandler also implements io.Closer, then Close will be
// called when the connection ends. Its return value is ignored.
type MessageHandler interface {
	// HandleMessage is passed received messages from the client. Its
	// return value is then sent back to the client with the same tag.
	HandleMessage(any) any
}

// MessageHandlerFunc allows a function to be used as a MessageHandler.
type MessageHandlerFunc func(any) any

func (h MessageHandlerFunc) HandleMessage(msg any) any {
	return h(msg)
}

// Msizer is implemented by types that, when returned from a message
// handler, should modify the maximum message size that the server
// should from that point forward.
//
// Note that if this is returned more than once for a single
// connection, a warning will be printed to stderr and the later
// values will be ignored.
type Msizer interface {
	P9Msize() uint32
}
