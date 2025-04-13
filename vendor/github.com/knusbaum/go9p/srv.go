// +build !plan9

package go9p

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"9fans.net/go/plan9/client"
	"github.com/fhs/mux9p"
)

type readWriteCloser interface {
	io.Reader
	io.Writer
	Close()
}

type pipe struct {
	sync.Mutex
	in   <-chan []byte
	inbs []byte
	out  chan<- []byte
}

func (p *pipe) Read(b []byte) (n int, err error) {
	if len(p.inbs) > 0 {
		copied := copy(b, p.inbs)
		p.inbs = p.inbs[copied:]
		return copied, nil
	}

	bs := <-p.in
	if bs == nil {
		return 0, fmt.Errorf("pipe closed.")
	}
	copied := copy(b, bs)
	if copied < len(bs) {
		p.inbs = bs[copied:]
	}
	return copied, nil
}

func (p *pipe) Write(b []byte) (n int, err error) {
	p.out <- b
	return len(b), nil
}

func (p *pipe) Close() {
	close(p.out)
}

func newPipePair() (*pipe, *pipe) {
	c1 := make(chan []byte, 100)
	c2 := make(chan []byte, 100)
	p1 := &pipe{
		in:  c1,
		out: c2,
	}
	p2 := &pipe{
		in:  c2,
		out: c1,
	}
	return p1, p2
}

func postfd(name string) (readWriteCloser, *os.File, error) {
	ns := client.Namespace()
	if err := os.MkdirAll(ns, 0700); err != nil {
		return nil, nil, err
	}
	name = filepath.Join(ns, name)

	f1, f2 := newPipePair()
	go func() {
		err := mux9p.Listen("unix", name, f2, &mux9p.Config{})
		if err != nil {
			log.Printf("Error: %v", err)
		}
		f2.Close()
	}()
	return f1, nil, nil
}
