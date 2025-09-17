package proto

import (
	"context"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/DeedleFake/p9/internal/debug"
)

// Client provides functionality for sending requests to and receiving
// responses from a server for a given protocol. It automatically
// handles message tags, properly blocking until a matching tag
// response has been received.
type Client struct {
	done   chan struct{}
	cancel func()

	p Proto
	c net.Conn

	nextTag   chan uint16
	sentMsg   chan clientMsg
	recvMsg   chan clientMsg
	cancelMsg chan uint16

	msize uint32
}

// NewClient initializes a client that communicates using c. The
// Client will close c when the Client is closed.
func NewClient(p Proto, c net.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		done:   make(chan struct{}),
		cancel: cancel,

		p: p,
		c: c,

		nextTag:   make(chan uint16),
		sentMsg:   make(chan clientMsg),
		recvMsg:   make(chan clientMsg),
		cancelMsg: make(chan uint16),

		msize: 1024,
	}
	go client.reader(ctx)
	go client.coord(ctx)

	return client
}

// Dial is a convenience function that dials and creates a client in
// the same step.
func Dial(p Proto, network, addr string) (*Client, error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	return NewClient(p, c), nil
}

// Close cleans up resources created by the client as well as closing
// the underlying connection.
func (c *Client) Close() error {
	c.cancel()
	return c.c.Close()
}

// reader reads messages from the connection, sending them to the
// coordinator to be sent to waiting Send calls.
func (c *Client) reader(ctx context.Context) {
	for {
		err := c.c.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			log.Printf("Failed to set conn deadline: %v", err)
			return
		}

		msg, tag, err := c.p.Receive(c.c, c.Msize())
		if err != nil {
			if (ctx.Err() != nil) || (err == io.EOF) {
				return
			}

			continue
		}

		select {
		case <-ctx.Done():
			return

		case c.recvMsg <- clientMsg{
			tag:  tag,
			recv: msg,
		}:
		}
	}
}

// coord coordinates between Send calls and the reader.
func (c *Client) coord(ctx context.Context) {
	defer close(c.done)

	var nextTag uint16
	tags := make(map[uint16]chan interface{})

	for {
		select {
		case <-ctx.Done():
			return

		case cm := <-c.sentMsg:
			tags[cm.tag] = cm.ret

		case cm := <-c.recvMsg:
			rcm, ok := tags[cm.tag]
			if !ok {
				continue
			}

			rcm <- cm.recv
			delete(tags, cm.tag)

		case tag := <-c.cancelMsg:
			delete(tags, tag)

		case c.nextTag <- nextTag:
			for {
				nextTag++
				if _, ok := tags[nextTag]; !ok {
					break
				}
			}
		}
	}
}

// Msize returns the maxiumum size of a message. This does not perform
// any communication with the server.
func (c *Client) Msize() uint32 {
	return atomic.LoadUint32(&c.msize)
}

// SetMsize sets the maximum size of a message. This does not perform
// any communication with the server.
func (c *Client) SetMsize(size uint32) {
	atomic.StoreUint32(&c.msize, size)
}

// Send sends a message to the server, blocking until a response has
// been received. It is safe to place multiple Send calls
// concurrently, and each will return when the response to that
// request has been received.
func (c *Client) Send(msg interface{}) (interface{}, error) {
	debug.Log("client -> %#v\n", msg)

	tag := NoTag
	if _, ok := msg.(P9NoTag); !ok {
		select {
		case <-c.done:
			return nil, ErrClientClosed
		case tag = <-c.nextTag:
		}
	}

	ret := make(chan interface{}, 1)
	select {
	case <-c.done:
		return nil, ErrClientClosed

	case c.sentMsg <- clientMsg{
		tag: tag,
		ret: ret,
	}:
	}

	err := c.p.Send(c.c, tag, msg)
	if err != nil {
		select {
		case <-c.done:
		case c.cancelMsg <- tag:
		}
		return nil, err
	}

	select {
	case <-c.done:
		return nil, ErrClientClosed
	case rsp := <-ret:
		debug.Log("client <- %#v\n", rsp)

		if err, ok := rsp.(error); ok {
			return nil, err
		}
		return rsp, nil
	}
}

// Sometimes I think that some type of tuples would be nice...
type clientMsg struct {
	tag  uint16
	recv interface{}
	ret  chan interface{}
}

// P9NoTag is implemented by any types that should not use tags for
// communicating. In 9P, for example, this is true of the Tversion
// message type, as it must be the first thing sent and no further
// communication can happen before an Rversion is sent in response.
type P9NoTag interface {
	P9NoTag()
}
