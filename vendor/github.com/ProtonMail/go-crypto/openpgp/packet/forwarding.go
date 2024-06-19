package packet

import "encoding/binary"

// ForwardingInstance represents a single forwarding instance (mapping IDs to a Proxy Param)
type ForwardingInstance struct {
	KeyVersion int
	ForwarderFingerprint []byte
	ForwardeeFingerprint []byte
	ProxyParameter       []byte
}

func (f *ForwardingInstance) GetForwarderKeyId() uint64 {
	return computeForwardingKeyId(f.ForwarderFingerprint, f.KeyVersion)
}

func (f *ForwardingInstance) GetForwardeeKeyId() uint64 {
	return computeForwardingKeyId(f.ForwardeeFingerprint, f.KeyVersion)
}

func (f *ForwardingInstance) getForwardeeKeyIdOrZero(originalKeyId uint64) uint64 {
	if originalKeyId == 0 {
		return 0
	}

	return f.GetForwardeeKeyId()
}

func computeForwardingKeyId(fingerprint []byte, version int) uint64 {
	switch version {
	case 4:
		return binary.BigEndian.Uint64(fingerprint[12:20])
	default:
		panic("invalid pgp key version")
	}
}