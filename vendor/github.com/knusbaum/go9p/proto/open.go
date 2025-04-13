package proto

import "fmt"

type Mode uint8

// Open mode file constants
const (
	Oread   Mode = 0
	Owrite  Mode = 1
	Ordwr   Mode = 2
	Oexec   Mode = 3
	None    Mode = 4
	Otrunc  Mode = 0x10
	Orclose Mode = 0x40
)

const (
	IOUnit = 16384
)

type TOpen struct {
	Header
	Fid uint32
	Mode
}

func (open *TOpen) String() string {
	return fmt.Sprintf("topen: [%s, fid: %d, mode: %d]",
		&open.Header, open.Fid, open.Mode)
}

func (open *TOpen) parse(buff []byte) ([]byte, error) {
	open.Fid, buff = fromLittleE32(buff)
	if len(buff) < 1 {
		return buff, fmt.Errorf("short fcall")
	}
	open.Mode = Mode(buff[0])
	return buff[1:], nil
}

func (open *TOpen) Compose() []byte {
	// size[4] Topen tag[2] fid[4] mode[1]
	length := 4 + 1 + 2 + 4 + 1
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = open.Type
	buffer = buffer[1:]
	buffer = toLittleE16(open.Tag, buffer)
	buffer = toLittleE32(open.Fid, buffer)
	buffer[0] = byte(open.Mode)
	buffer = buffer[1:]
	return buff
}

type ROpen struct {
	Header
	Qid    Qid
	Iounit uint32
}

func (open *ROpen) String() string {
	return fmt.Sprintf("ropen: [%s, qid: [%s], iounit: %d]",
		&open.Header, &open.Qid, open.Iounit)
}

func (open *ROpen) parse(buff []byte) ([]byte, error) {
	buff, err := open.Qid.parse(buff)
	if err != nil {
		return nil, err
	}
	open.Iounit, buff = fromLittleE32(buff)
	return buff, nil
}

func (open *ROpen) Compose() []byte {
	// size[4] Ropen tag[2] qid[13] iounit[4]
	length := 4 + 1 + 2 + 13 + 4
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = open.Type
	buffer = buffer[1:]
	buffer = toLittleE16(open.Tag, buffer)
	qidbuff := open.Qid.Compose()
	copy(buffer, qidbuff)
	buffer = buffer[len(qidbuff):]
	buffer = toLittleE32(open.Iounit, buffer)
	return buff
}
