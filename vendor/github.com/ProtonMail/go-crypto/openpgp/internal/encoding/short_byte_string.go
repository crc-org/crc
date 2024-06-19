package encoding

import (
	"io"
)

type ShortByteString struct {
	length uint8
	data   []byte
}

func NewShortByteString(data []byte) *ShortByteString {
	byteLength := uint8(len(data))

	return &ShortByteString{byteLength, data}
}

func (byteString *ShortByteString) Bytes() []byte {
	return byteString.data
}

func (byteString *ShortByteString) BitLength() uint16 {
	return uint16(byteString.length) * 8
}

func (byteString *ShortByteString) EncodedBytes() []byte {
	encodedLength := [1]byte{
		uint8(byteString.length),
	}
	return append(encodedLength[:], byteString.data...)
}

func (byteString *ShortByteString) EncodedLength() uint16 {
	return uint16(byteString.length) + 1
}

func (byteString *ShortByteString) ReadFrom(r io.Reader) (int64, error) {
	var lengthBytes [1]byte
	if n, err := io.ReadFull(r, lengthBytes[:]); err != nil {
		return int64(n), err
	}

	byteString.length = uint8(lengthBytes[0])

	byteString.data = make([]byte, byteString.length)
	if n, err := io.ReadFull(r, byteString.data); err != nil {
		return int64(n + 1), err
	}
	return int64(byteString.length + 1), nil
}
