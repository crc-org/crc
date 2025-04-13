package proto

import "fmt"

type TClunk struct {
	Header
	Fid uint32
}

func (clunk *TClunk) String() string {
	return fmt.Sprintf("tclunk: [%s, fid: %d]", &clunk.Header, clunk.Fid)
}

func (clunk *TClunk) parse(buff []byte) ([]byte, error) {
	clunk.Fid, buff = fromLittleE32(buff)
	return buff, nil
}

func (clunk *TClunk) Compose() []byte {
	// size[4] Tclunk tag[2] fid[4]
	length := 4 + 1 + 2 + 4
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = clunk.Type
	buffer = buffer[1:]
	buffer = toLittleE16(clunk.Tag, buffer)
	buffer = toLittleE32(clunk.Fid, buffer)
	return buff
}

type RClunk struct {
	Header
}

func (clunk *RClunk) String() string {
	return fmt.Sprintf("rclunk: [%s]", &clunk.Header)
}

func (clunk *RClunk) parse(buff []byte) ([]byte, error) {
	return buff, nil
}

func (clunk *RClunk) Compose() []byte {
	// size[4] Rclunk tag[2]
	length := 4 + 1 + 2
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = clunk.Type
	buffer = buffer[1:]
	buffer = toLittleE16(clunk.Tag, buffer)
	return buff
}
