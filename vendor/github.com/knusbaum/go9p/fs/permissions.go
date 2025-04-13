package fs

import (
	"github.com/knusbaum/go9p/proto"
)

const (
	ugo_user  = iota
	ugo_group = iota
	ugo_other = iota
)

func userInGroup(user string, group string) bool {
	// For now groups and users are equivalent.
	return user == group
}

func userRelation(user string, f FSNode) uint8 {
	st := f.Stat()
	if user == st.Uid {
		return ugo_user
	}
	if userInGroup(user, st.Gid) {
		return ugo_group
	}
	return ugo_other
}

func omodePermits(perm uint8, omode proto.Mode) bool {
	switch omode {
	case proto.Oread:
		return perm&0x4 != 0
		break
	case proto.Owrite:
		return perm&0x2 != 0
		break
	case proto.Ordwr:
		return (perm&0x2 != 0) && (perm&0x4 != 0)
		break
	case proto.Oexec:
		return perm&0x01 != 0
		break
	case proto.None:
		return false
		break
	default:
		return false
		break
	}
	return false
}

func openPermission(f FSNode, user string, omode proto.Mode) bool {
	switch userRelation(user, f) {
	case ugo_user:
		return omodePermits(uint8(f.Stat().Mode>>6)&0x07, omode)
		break
	case ugo_group:
		return omodePermits(uint8(f.Stat().Mode>>3)&0x07, omode)
		break
	case ugo_other:
		return omodePermits(uint8(f.Stat().Mode)&0x07, omode)
		break
	default:
		return false
	}
	return false
}
