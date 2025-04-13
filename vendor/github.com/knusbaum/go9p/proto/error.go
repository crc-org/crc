package proto

import (
	"fmt"
)

type RError struct {
	Header
	Ename string
}

func (error *RError) String() string {
	return fmt.Sprintf("rerror: [%s, ename: %s]",
		&error.Header, error.Ename)
}

func (error *RError) parse(buff []byte) ([]byte, error) {
	error.Ename, buff = fromString(buff)
	return buff, nil
}

func (error *RError) Compose() []byte {
	// size[4] Rerror tag[2] ename[s]
	length := 4 + 1 + 2 + (2 + len(error.Ename))
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = error.Type
	buffer = buffer[1:]
	buffer = toLittleE16(error.Tag, buffer)
	buffer = toString(error.Ename, buffer)

	return buff
}
