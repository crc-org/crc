package p9

import (
	"math"
)

const (
	// Version is the 9P version implemented by this package, both for
	// server and client.
	Version = "9P2000"
)

const (
	// NoTag is a special tag that is used when the tag is unimportant.
	NoTag uint16 = math.MaxUint16

	// NoFID is a special FID that is used to signal a lack of an FID.
	NoFID uint32 = math.MaxUint32
)

// File open modes and flags. Note that not all flags are supported
// where you might expect them to be. If you get an error saying that
// a flag won't fit into uint8, the flag you're trying to use probably
// isn't supported there.
const (
	OREAD uint8 = iota
	OWRITE
	ORDWR
	OEXEC

	OTRUNC = 1 << (iota + 1)
	OCEXEC
	ORCLOSE
	ODIRECT
	ONONBLOCK
	OEXCL
	OLOCK
	OAPPEND
)

// QID represents a QID value.
type QID struct {
	Type    QIDType
	Version uint32
	Path    uint64
}

// QIDType represents an 8-bit QID type identifier.
type QIDType uint8

// Valid types of files.
const (
	QTFile QIDType = 0
	QTDir  QIDType = 1 << (8 - iota)
	QTAppend
	QTExcl
	QTMount
	QTAuth
	QTTmp
	QTSymlink
)

// FileMode converts the QIDType to a FileMode.
func (t QIDType) FileMode() FileMode {
	return FileMode(t) << 24
}

// Other constants.
const (
	IOHeaderSize = 24
)
