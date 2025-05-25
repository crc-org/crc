// Copyright 2009 The Ninep Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srv

import "fmt"
import "github.com/lionkov/ninep"

// Respond to the request with Rerror message
func (req *Req) RespondError(err interface{}) {
	switch e := err.(type) {
	case *ninep.Error:
		ninep.PackRerror(req.Rc, e.Error(), uint32(e.Errornum), req.Conn.Dotu)
	case error:
		ninep.PackRerror(req.Rc, e.Error(), uint32(ninep.EIO), req.Conn.Dotu)
	default:
		ninep.PackRerror(req.Rc, fmt.Sprintf("%v", e), uint32(ninep.EIO), req.Conn.Dotu)
	}

	req.Respond()
}

// Respond to the request with Rversion message
func (req *Req) RespondRversion(msize uint32, version string) {
	err := ninep.PackRversion(req.Rc, msize, version)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rauth message
func (req *Req) RespondRauth(aqid *ninep.Qid) {
	err := ninep.PackRauth(req.Rc, aqid)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rflush message
func (req *Req) RespondRflush() {
	err := ninep.PackRflush(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rattach message
func (req *Req) RespondRattach(aqid *ninep.Qid) {
	err := ninep.PackRattach(req.Rc, aqid)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwalk message
func (req *Req) RespondRwalk(wqids []ninep.Qid) {
	err := ninep.PackRwalk(req.Rc, wqids)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Ropen message
func (req *Req) RespondRopen(qid *ninep.Qid, iounit uint32) {
	err := ninep.PackRopen(req.Rc, qid, iounit)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rcreate message
func (req *Req) RespondRcreate(qid *ninep.Qid, iounit uint32) {
	err := ninep.PackRcreate(req.Rc, qid, iounit)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rread message
func (req *Req) RespondRread(data []byte) {
	err := ninep.PackRread(req.Rc, data)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwrite message
func (req *Req) RespondRwrite(count uint32) {
	err := ninep.PackRwrite(req.Rc, count)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rclunk message
func (req *Req) RespondRclunk() {
	err := ninep.PackRclunk(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rremove message
func (req *Req) RespondRremove() {
	err := ninep.PackRremove(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rstat message
func (req *Req) RespondRstat(st *ninep.Dir) {
	err := ninep.PackRstat(req.Rc, st, req.Conn.Dotu)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwstat message
func (req *Req) RespondRwstat() {
	err := ninep.PackRwstat(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}
