package proto

import (
	"fmt"
)

type TAttach struct {
	Header
	Fid   uint32
	Afid  uint32
	Uname string
	Aname string
}

func (attach *TAttach) String() string {
	return fmt.Sprintf("tattach: [%s, fid: %d, afid: %d, uname: %s, aname: %s]",
		&attach.Header, attach.Fid, attach.Afid, attach.Uname, attach.Aname)
}

func (attach *TAttach) parse(buff []byte) ([]byte, error) {
	attach.Fid, buff = fromLittleE32(buff)
	attach.Afid, buff = fromLittleE32(buff)
	attach.Uname, buff = fromString(buff)
	attach.Aname, buff = fromString(buff)
	return buff, nil
}

func (attach *TAttach) Compose() []byte {
	// size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
	length := 4 + 1 + 2 + 4 + 4 +
		(2 + len(attach.Uname)) + (2 + len(attach.Aname))
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = attach.Type
	buffer = buffer[1:]
	buffer = toLittleE16(attach.Tag, buffer)
	buffer = toLittleE32(attach.Fid, buffer)
	buffer = toLittleE32(attach.Afid, buffer)
	buffer = toString(attach.Uname, buffer)
	buffer = toString(attach.Aname, buffer)
	return buff
}

type RAttach struct {
	Header
	Qid Qid
}

func (attach *RAttach) String() string {
	return fmt.Sprintf("rattach: [%s, qid: [%s]]",
		&attach.Header, &attach.Qid)
}

func (attach *RAttach) parse(buff []byte) ([]byte, error) {
	buff, err := attach.Qid.parse(buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (attach *RAttach) Compose() []byte {
	// size[4] Rattach tag[2] qid[13]
	length := 4 + 1 + 2 + 13
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = attach.Type
	buffer = buffer[1:]
	buffer = toLittleE16(attach.Tag, buffer)
	qidbuff := attach.Qid.Compose()
	copy(buffer, qidbuff)
	return buff
}
