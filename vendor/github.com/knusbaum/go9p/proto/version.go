package proto

import (
	"fmt"
)

type TRVersion struct {
	Header
	Msize   uint32
	Version string
}

func (version *TRVersion) String() string {
	var c byte
	if version.Type == Tversion {
		c = 't'
	} else {
		c = 'r'
	}
	return fmt.Sprintf("%cversion: [%s, msize: %d, version: %s]",
		c, &version.Header, version.Msize, version.Version)
}

func (version *TRVersion) parse(buff []byte) ([]byte, error) {
	version.Msize, buff = fromLittleE32(buff)
	version.Version, buff = fromString(buff)
	return buff, nil
}

func (version *TRVersion) Compose() []byte {
	// size[4] Tversion tag[2] msize[4] version[s]
	length := 4 + 1 + 2 + 4 + (2 + len(version.Version))
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = version.Type
	buffer = buffer[1:]
	buffer = toLittleE16(version.Tag, buffer)
	buffer = toLittleE32(version.Msize, buffer)
	buffer = toString(version.Version, buffer)

	return buff
}
