package proto

import "fmt"

type TFlush struct {
	Header
	Oldtag uint16
}

func (flush *TFlush) String() string {
	return fmt.Sprintf("tflush: [%s, oldtag: %d]",
		&flush.Header, flush.Oldtag)
}

func (flush *TFlush) parse(buff []byte) ([]byte, error) {
	flush.Oldtag, buff = fromLittleE16(buff)
	return buff, nil
}

func (flush *TFlush) Compose() []byte {
	// size[4] Tflush tag[2] oldtag[2]
	length := 4 + 1 + 2 + 2
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = flush.Type
	buffer = buffer[1:]
	buffer = toLittleE16(flush.Tag, buffer)
	buffer = toLittleE16(flush.Oldtag, buffer)
	return buff
}

type RFlush struct {
	Header
}

func (flush *RFlush) String() string {
	return fmt.Sprintf("rflush: [%s]", &flush.Header)
}

func (flush *RFlush) parse(buff []byte) ([]byte, error) {
	return buff, nil
}

func (flush *RFlush) Compose() []byte {
	// size[4] Rflush tag[2]
	length := 4 + 1 + 2
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = flush.Type
	buffer = buffer[1:]
	buffer = toLittleE16(flush.Tag, buffer)
	return buff
}
