// This file is based on 9fans.net/go, which has this license:
//
// Copyright (c) 2009 Google Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Subject to the terms and conditions of this License, Google hereby
// grants to You a perpetual, worldwide, non-exclusive, no-charge,
// royalty-free, irrevocable (except as stated in this section) patent
// license to make, have made, use, offer to sell, sell, import, and
// otherwise transfer this implementation of Go, where such license
// applies only to those patent claims licensable by Google that are
// necessarily infringed by use of this implementation of Go. If You
// institute patent litigation against any entity (including a
// cross-claim or counterclaim in a lawsuit) alleging that this
// implementation of Go or a Contribution incorporated within this
// implementation of Go constitutes direct or contributory patent
// infringement, then any patent licenses granted to You under this
// License for this implementation of Go shall terminate as of the date
// such litigation is filed.

package p9p

import (
	"fmt"
	"io"

	"9fans.net/go/plan9"
)

// The openfd 9P message type is from plan9port (see openfd(9)).
const (
	Topenfd = 98 + iota
	Ropenfd
)

// Fcall adds support for Topenfd and Ropenfd to 9fans.net/go/plan9.Fcall.
type Fcall struct {
	plan9.Fcall
	Unixfd uint32 // Ropenfd
}

// String returns a string representation.
func (f *Fcall) String() string {
	switch f.Type {
	case Topenfd:
		return fmt.Sprintf("Topenfd tag %v fid %v mode %o", f.Tag, f.Fid, f.Mode)
	case Ropenfd:
		return fmt.Sprintf("Ropenfd tag %v qid %v iounit %v unixfd %v", f.Tag, f.Qid, f.Iounit, f.Unixfd)
	}
	return f.Fcall.String()
}

// Bytes marshals the Fcall.
func (f *Fcall) Bytes() ([]byte, error) {
	b := pbit32(nil, 0) // length: fill in later
	b = pbit8(b, f.Type)
	b = pbit16(b, f.Tag)
	switch f.Type {
	default:
		return f.Fcall.Bytes()

	case Topenfd:
		b = pbit32(b, f.Fid)
		b = pbit8(b, f.Mode)

	case Ropenfd:
		b = pqid(b, f.Qid)
		b = pbit32(b, f.Iounit)
		b = pbit32(b, f.Unixfd)
		// Caller is responsible for sending f.Unixfd using SendFD.
	}

	pbit32(b[0:0], uint32(len(b)))
	return b, nil
}

// ReadFcall reads a Fcall from reader r.
func ReadFcall(r io.Reader) (*Fcall, error) {
	// 128 bytes should be enough for most messages
	buf := make([]byte, 128)
	_, err := io.ReadFull(r, buf[0:4])
	if err != nil {
		return nil, err
	}

	// read 4-byte header, make room for remainder
	n, _ := gbit32(buf)
	if n < 4 {
		return nil, plan9.ProtocolError("invalid length")
	}
	if int(n) <= len(buf) {
		buf = buf[0:n]
	} else {
		buf = make([]byte, n)
		pbit32(buf[0:0], n)
	}

	// read remainder and unpack
	_, err = io.ReadFull(r, buf[4:])
	if err != nil {
		return nil, err
	}
	fc, err := UnmarshalFcall(buf)
	if err != nil {
		return nil, err
	}
	return fc, nil
}

// UnmarshalFcall unmarshals a Fcall.
func UnmarshalFcall(b []byte) (f *Fcall, err error) {
	ob := b[:]
	n, b := gbit32(b)
	if len(b) != int(n)-4 {
		return nil, plan9.ProtocolError("malformed Fcall")
	}

	f = new(Fcall)
	f.Type, b = gbit8(b)
	f.Tag, b = gbit16(b)

	switch f.Type {
	default:
		fc, err := plan9.UnmarshalFcall(ob)
		if err != nil {
			return nil, err
		}
		f.Fcall = *fc
		return f, nil

	case Topenfd:
		f.Fid, b = gbit32(b)
		f.Mode, b = gbit8(b)

	case Ropenfd:
		f.Qid, b = gqid(b)
		f.Iounit, b = gbit32(b)
		f.Unixfd, b = gbit32(b)
		// Caller is responsible for receiving the real unixfd using ReceiveFD.
	}

	if len(b) != 0 {
		return nil, plan9.ProtocolError("malformed Fcall")
	}

	return f, nil
}

func gbit8(b []byte) (uint8, []byte) {
	return uint8(b[0]), b[1:]
}

func gbit16(b []byte) (uint16, []byte) {
	return uint16(b[0]) | uint16(b[1])<<8, b[2:]
}

func gbit32(b []byte) (uint32, []byte) {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24, b[4:]
}

func gbit64(b []byte) (uint64, []byte) {
	lo, b := gbit32(b)
	hi, b := gbit32(b)
	return uint64(hi)<<32 | uint64(lo), b
}

func pbit8(b []byte, x uint8) []byte {
	n := len(b)
	if n+1 > cap(b) {
		nb := make([]byte, n, 100+2*cap(b))
		copy(nb, b)
		b = nb
	}
	b = b[0 : n+1]
	b[n] = x
	return b
}

func pbit16(b []byte, x uint16) []byte {
	n := len(b)
	if n+2 > cap(b) {
		nb := make([]byte, n, 100+2*cap(b))
		copy(nb, b)
		b = nb
	}
	b = b[0 : n+2]
	b[n] = byte(x)
	b[n+1] = byte(x >> 8)
	return b
}

func pbit32(b []byte, x uint32) []byte {
	n := len(b)
	if n+4 > cap(b) {
		nb := make([]byte, n, 100+2*cap(b))
		copy(nb, b)
		b = nb
	}
	b = b[0 : n+4]
	b[n] = byte(x)
	b[n+1] = byte(x >> 8)
	b[n+2] = byte(x >> 16)
	b[n+3] = byte(x >> 24)
	return b
}

func pbit64(b []byte, x uint64) []byte {
	b = pbit32(b, uint32(x))
	b = pbit32(b, uint32(x>>32))
	return b
}

func gqid(b []byte) (plan9.Qid, []byte) {
	var q plan9.Qid
	q.Type, b = gbit8(b)
	q.Vers, b = gbit32(b)
	q.Path, b = gbit64(b)
	return q, b
}

func pqid(b []byte, q plan9.Qid) []byte {
	b = pbit8(b, q.Type)
	b = pbit32(b, q.Vers)
	b = pbit64(b, q.Path)
	return b
}
