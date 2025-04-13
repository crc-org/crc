package proto

import "fmt"

type TRemove struct {
	Header
	Fid uint32
}

func (remove *TRemove) String() string {
	return fmt.Sprintf("tremove: [%s, fid: %d]", &remove.Header, remove.Fid)
}

func (remove *TRemove) parse(buff []byte) ([]byte, error) {
	remove.Fid, buff = fromLittleE32(buff)
	return buff, nil
}

func (remove *TRemove) Compose() []byte {
	// size[4] Tremove tag[2] fid[4]
	length := 4 + 1 + 2 + 4
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = remove.Type
	buffer = buffer[1:]
	buffer = toLittleE16(remove.Tag, buffer)
	buffer = toLittleE32(remove.Fid, buffer)
	return buff
}

type RRemove struct {
	Header
}

func (remove *RRemove) String() string {
	return fmt.Sprintf("rremove: [%s]", &remove.Header)
}

func (remove *RRemove) parse(buff []byte) ([]byte, error) {
	return buff, nil
}

func (remove *RRemove) Compose() []byte {
	// size[4] Rwstat tag[2]
	length := 4 + 1 + 2
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = remove.Type
	buffer = buffer[1:]
	buffer = toLittleE16(remove.Tag, buffer)
	return buff
}
