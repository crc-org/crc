package proto

import "fmt"

type TRead struct {
	Header
	Fid    uint32
	Offset uint64
	Count  uint32
}

func (read *TRead) String() string {
	return fmt.Sprintf("tread: [%s, fid: %d, offset: %d, count: %d]",
		&read.Header, read.Fid, read.Offset, read.Count)
}

func (read *TRead) parse(buff []byte) ([]byte, error) {
	read.Fid, buff = fromLittleE32(buff)
	read.Offset, buff = fromLittleE64(buff)
	read.Count, buff = fromLittleE32(buff)
	return buff, nil
}

func (read *TRead) Compose() []byte {
	// size[4] Twrite tag[2] fid[4] offset[8] count[4]
	length := 4 + 1 + 2 + 4 + 8 + 4
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = read.Type
	buffer = buffer[1:]
	buffer = toLittleE16(read.Tag, buffer)
	buffer = toLittleE32(read.Fid, buffer)
	buffer = toLittleE64(read.Offset, buffer)
	buffer = toLittleE32(read.Count, buffer)
	return buff
}

type RRead struct {
	Header
	Count uint32
	Data  []byte
}

func (read *RRead) String() string {
	return fmt.Sprintf("rread: [%s, count: %d]", &read.Header, read.Count)
}

func (read *RRead) parse(buff []byte) ([]byte, error) {
	read.Count, buff = fromLittleE32(buff)
	read.Data = make([]byte, read.Count)
	copy(read.Data, buff[:read.Count])
	return buff[read.Count:], nil
}

func (read *RRead) Compose() []byte {
	// size[4] Rread tag[2] count[4] data[count]
	length := 4 + 1 + 2 + 4 + read.Count
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = read.Type
	buffer = buffer[1:]
	buffer = toLittleE16(read.Tag, buffer)
	buffer = toLittleE32(read.Count, buffer)
	copy(buffer, read.Data)
	return buff
}
