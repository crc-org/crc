package proto

import (
	"encoding/binary"
	"io"
	"reflect"
	"time"
	"unsafe"

	"github.com/DeedleFake/p9/internal/util"
)

// Size returns the size of a message after encoding.
func Size(v any) (uint32, error) {
	e := &encoder{}
	e.mode = e.size

	e.encode(reflect.ValueOf(v))
	return e.n, e.err
}

// Write encodes and writes a message to w. It does not perform any
// buffering. It is the caller's responsibility to ensure that
// encoding does not interleave with other usages of w.
func Write(w io.Writer, v any) error {
	e := &encoder{w: w}
	e.mode = e.write

	e.encode(reflect.ValueOf(v))
	return e.err
}

// Read decodes a message from r into v.
func Read(r io.Reader, v any) error {
	d := &decoder{r: r}
	d.decode(reflect.ValueOf(v))
	return d.err
}

type encoder struct {
	w   io.Writer
	n   uint32
	err error

	mode func(any)
}

func (e *encoder) size(v any) {
	e.n += uint32(binary.Size(v))
}

func (e *encoder) write(v any) {
	if e.err != nil {
		return
	}

	e.err = binary.Write(e.w, binary.LittleEndian, v)
}

func (e *encoder) encode(v reflect.Value) {
	if e.err != nil {
		return
	}

	switch v := v.Interface().(type) {
	case Encoder:
		buf, err := v.P9Encode()
		e.err = err
		e.mode(buf)
		return
	}

	v = reflect.Indirect(v)

	switch v := v.Interface().(type) {
	case time.Time:
		e.mode(uint32(v.Unix()))
		return
	}

	switch v.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uintptr:
		e.mode(v.Interface())

	case reflect.Array, reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			e.mode(uint32(v.Len()))
		default:
			e.mode(uint16(v.Len()))
		}

		for i := 0; i < v.Len(); i++ {
			e.encode(v.Index(i))
		}

	case reflect.String:
		e.mode(uint16(v.Len()))
		e.mode([]byte(v.String()))

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			e.encode(v.Field(i))
		}

	default:
		e.err = util.Errorf("invalid type: %T", v)
	}
}

type decoder struct {
	r   io.Reader
	err error
}

func (d *decoder) read(v any) {
	if d.err != nil {
		return
	}

	d.err = binary.Read(d.r, binary.LittleEndian, v)
}

func (d *decoder) decode(v reflect.Value) {
	if d.err != nil {
		return
	}

	v = reflect.Indirect(v)

	switch v := v.Addr().Interface().(type) {
	case Decoder:
		d.err = v.P9Decode(d.r)
		return
	case *time.Time:
		var unix uint32
		d.read(&unix)
		*v = time.Unix(int64(unix), 0)
		return
	}

	switch v.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uintptr:
		d.read(v.Addr().Interface())

	case reflect.Slice:
		var length uint32
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			d.read(&length)
		default:
			var t uint16
			d.read(&t)
			length = uint32(t)
		}

		if int(length) > v.Cap() {
			v.Set(reflect.MakeSlice(v.Type(), int(length), int(length)))
		}
		v.Set(v.Slice(0, int(length)))

		for i := 0; i < v.Len(); i++ {
			d.decode(v.Index(i))
		}

	case reflect.String:
		var length uint16
		d.read(&length)

		buf := make([]byte, int(length))
		d.read(buf)

		v.SetString(unsafe.String(unsafe.SliceData(buf), len(buf)))

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			d.decode(v.Field(i))
		}

	default:
		d.err = util.Errorf("invalid type: %T", v)
	}
}

// Encoder is implemented by types that want to encode themselves in
// a customized, non-standard way.
type Encoder interface {
	P9Encode() ([]byte, error)
}

// Decoder is implemented by types that want to decode themselves in
// a customized, non-standard way.
type Decoder interface {
	P9Decode(r io.Reader) error
}
