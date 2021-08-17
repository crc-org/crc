package tap

import "encoding/binary"

type protocol interface {
	Buf() []byte
	Write(buf []byte, size int)
	Read(buf []byte) int
}

type hyperkitProtocol struct {
}

func (s *hyperkitProtocol) Buf() []byte {
	return make([]byte, 2)
}

func (s *hyperkitProtocol) Write(buf []byte, size int) {
	binary.LittleEndian.PutUint16(buf, uint16(size))
}

func (s *hyperkitProtocol) Read(buf []byte) int {
	return int(binary.LittleEndian.Uint16(buf[0:2]))
}

type qemuProtocol struct {
}

func (s *qemuProtocol) Buf() []byte {
	return make([]byte, 4)
}

func (s *qemuProtocol) Write(buf []byte, size int) {
	binary.BigEndian.PutUint32(buf, uint32(size))
}

func (s *qemuProtocol) Read(buf []byte) int {
	return int(binary.BigEndian.Uint32(buf[0:4]))
}
