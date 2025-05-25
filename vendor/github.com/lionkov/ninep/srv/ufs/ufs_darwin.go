// Copyright 2012 The Ninep Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ufs

import (
	"syscall"
	"time"
)

func atime(stat *syscall.Stat_t) time.Time {
	return time.Unix(stat.Atimespec.Unix())
}
