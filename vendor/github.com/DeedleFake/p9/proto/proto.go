// Package proto provides a usage-inspecific wrapper around 9P's data
// serialization and communication scheme.
package proto

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"sync"

	"github.com/DeedleFake/p9/internal/util"
)

var (
	// ErrLargeMessage is returned by various functions if a message is
	// larger than the current maximum message size.
	ErrLargeMessage = errors.New("message larger than msize")

	// ErrClientClosed is returned by attempts to send to a closed
	// client.
	ErrClientClosed = errors.New("client closed")
)

const (
	// NoTag is a special tag that is used when a tag is unimportant.
	NoTag uint16 = 0xFFFF
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Proto represents a protocol. It maps between message type IDs and
// the Go types that those IDs correspond to.
type Proto struct {
	rmap map[uint8]reflect.Type
	smap map[reflect.Type]uint8
}

// NewProto builds a Proto from the given one-way mapping.
func NewProto(mapping map[uint8]reflect.Type) Proto {
	rmap := make(map[uint8]reflect.Type, len(mapping))
	smap := make(map[reflect.Type]uint8, len(mapping))
	for id, t := range mapping {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		rmap[id] = t
		smap[t] = id
	}

	return Proto{
		rmap: rmap,
		smap: smap,
	}
}

// TypeFromID returns the Go type that corresponds to the given ID. If
// the ID is not recognized, it returns nil.
func (p Proto) TypeFromID(id uint8) reflect.Type {
	return p.rmap[id]
}

// IDFromType returns the message type ID that corresponds to the
// given Go type, and a boolean indicating that the mapping is valid.
func (p Proto) IDFromType(t reflect.Type) (uint8, bool) {
	id, ok := p.smap[t]
	return id, ok
}

// Receive receives a message from r using the given maximum message
// size. It returns the message, the tag that the message was sent
// with, and an error, if any.
func (p Proto) Receive(r io.Reader, msize uint32) (msg interface{}, tag uint16, err error) {
	var size uint32
	err = Read(r, &size)
	if err != nil {
		return nil, NoTag, util.Errorf("receive: %w", err)
	}

	if (msize > 0) && (size > msize) {
		return nil, NoTag, util.Errorf("receive: %w", ErrLargeMessage)
	}

	lr := &util.LimitedReader{
		R: r,
		N: size,
		E: ErrLargeMessage,
	}

	var msgType uint8
	err = Read(lr, &msgType)
	if err != nil {
		return nil, NoTag, util.Errorf("receive: failed to read message type: %w", err)
	}

	t := p.TypeFromID(msgType)
	if t == nil {
		if err != nil {
			return nil, NoTag, err
		}

		return nil, NoTag, util.Errorf("receive: invalid message type: %v", msgType)
	}

	tag = NoTag
	err = Read(lr, &tag)
	if err != nil {
		return nil, tag, util.Errorf("receive: failed to read tag: %w", err)
	}

	m := reflect.New(t)
	err = Read(lr, m.Interface())
	if err != nil {
		return nil, tag, util.Errorf("receive %v: %w", m.Type().Elem(), err)
	}

	return m.Interface(), tag, err
}

// Send writes a message to w with the given tag. It returns any
// errors that occur.
//
// Encoded messages are buffered locally before sending to ensure that
// pieces of a message don't get mixed with another one.
func (p Proto) Send(w io.Writer, tag uint16, msg interface{}) (err error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		if err == nil {
			_, err = io.Copy(w, buf)
			if err != nil {
				err = util.Errorf("send: %w", err)
			}
		}

		buf.Reset()
		bufPool.Put(buf)
	}()

	msgType, ok := p.IDFromType(reflect.Indirect(reflect.ValueOf(msg)).Type())
	if !ok {
		return util.Errorf("send: invalid message type: %T", msg)
	}

	write := func(v interface{}) {
		if err != nil {
			return
		}

		err = Write(buf, v)
		if err != nil {
			err = util.Errorf("send %T: %w", msg, err)
		}
	}

	n, err := Size(msg)
	if err != nil {
		return util.Errorf("send %T: %w", msg, err)
	}

	write(4 + 1 + 2 + n)
	write(msgType)
	write(tag)
	write(msg)

	return err
}
