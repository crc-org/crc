package proto

import "fmt"

type TWrite struct {
	Header
	Fid    uint32
	Offset uint64
	Count  uint32
	Data   []byte
}

func (write *TWrite) String() string {
	return fmt.Sprintf("twrite: [%s, fid: %d, offset: %d, count: %d]",
		&write.Header, write.Fid, write.Offset, write.Count)
}

func (write *TWrite) parse(buff []byte) ([]byte, error) {
	write.Fid, buff = fromLittleE32(buff)
	write.Offset, buff = fromLittleE64(buff)
	write.Count, buff = fromLittleE32(buff)
	write.Data = make([]byte, write.Count)
	copy(write.Data, buff[:write.Count])
	return buff[write.Count:], nil
}

func (write *TWrite) Compose() []byte {
	// size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	length := 4 + 1 + 2 + 4 + 8 + 4 + write.Count
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = write.Type
	buffer = buffer[1:]
	buffer = toLittleE16(write.Tag, buffer)
	buffer = toLittleE32(write.Fid, buffer)
	buffer = toLittleE64(write.Offset, buffer)
	buffer = toLittleE32(write.Count, buffer)
	copy(buffer, write.Data)
	return buff
}

type RWrite struct {
	Header
	Count uint32
}

func (write *RWrite) String() string {
	return fmt.Sprintf("rwrite: [%s, count: %d]", &write.Header, write.Count)
}

func (write *RWrite) parse(buff []byte) ([]byte, error) {
	write.Count, buff = fromLittleE32(buff)
	return buff, nil
}

func (write *RWrite) Compose() []byte {
	// size[4] Rwrite tag[2] count[4]
	length := 4 + 1 + 2 + 4
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = write.Type
	buffer = buffer[1:]
	buffer = toLittleE16(write.Tag, buffer)
	buffer = toLittleE32(write.Count, buffer)
	return buff
}
