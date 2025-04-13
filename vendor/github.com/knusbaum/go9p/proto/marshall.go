package proto

import (
	"encoding/binary"
	"io"
)

func readBytes(r io.Reader, buff []byte) error {
	var read int
	var err error

	for read < len(buff) {
		currRead := 0
		currRead, err = r.Read(buff[read:])
		if err != nil {
			return err
		}
		read += currRead
	}
	return nil
}

func fromLittleE16(buff []byte) (uint16, []byte) {
	if len(buff) < 2 {
		return 0, nil
	}
	ret := binary.LittleEndian.Uint16(buff)
	return ret, buff[2:]
}

func fromLittleE32(buff []byte) (uint32, []byte) {
	if len(buff) < 4 {
		return 0, nil
	}
	ret := binary.LittleEndian.Uint32(buff)
	return ret, buff[4:]
}

func fromLittleE64(buff []byte) (uint64, []byte) {
	if len(buff) < 8 {
		return 0, nil
	}
	ret := binary.LittleEndian.Uint64(buff)
	return ret, buff[8:]
}

func fromString(buff []byte) (string, []byte) {
	var leng uint16
	leng, buff = fromLittleE16(buff)

	if len(buff) < int(leng) {
		return "", nil
	}
	ret := string(buff[:leng])
	return ret, buff[leng:]
}

func toLittleE16(i uint16, buff []byte) []byte {
	binary.LittleEndian.PutUint16(buff, i)
	return buff[2:]
}

func toLittleE32(i uint32, buff []byte) []byte {
	binary.LittleEndian.PutUint32(buff, i)
	return buff[4:]
}

func toLittleE64(i uint64, buff []byte) []byte {
	binary.LittleEndian.PutUint64(buff, i)
	return buff[8:]
}

func toString(s string, buff []byte) []byte {
	buff = toLittleE16(uint16(len(s)), buff)
	copy(buff, []byte(s))
	return buff[len(s):]
}

type ParseError struct {
	Err string
}

func (pe *ParseError) Error() string {
	return pe.Err
}
