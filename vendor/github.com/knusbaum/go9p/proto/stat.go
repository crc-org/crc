package proto

import "fmt"

const (
	DMDIR    = uint32(1 << 31)
	DMAPPEND = uint32(1 << 30)
	DMEXCL   = uint32(1 << 29)
	DMTMP    = uint32(1 << 26)
)

type TStat struct {
	Header
	Fid uint32
}

func (stat *TStat) String() string {
	return fmt.Sprintf("tstat: [%s, fid: %d]", &stat.Header, stat.Fid)
}

func (stat *TStat) parse(buff []byte) ([]byte, error) {
	stat.Fid, buff = fromLittleE32(buff)
	return buff, nil
}

func (stat *TStat) Compose() []byte {
	// size[4] Twrite tag[2] fid[4]
	length := 4 + 1 + 2 + 4
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = stat.Type
	buffer = buffer[1:]
	buffer = toLittleE16(stat.Tag, buffer)
	buffer = toLittleE32(stat.Fid, buffer)

	return buff
}

type Stat struct {
	Type   uint16
	Dev    uint32
	Qid    Qid
	Mode   uint32
	Atime  uint32
	Mtime  uint32
	Length uint64
	Name   string
	Uid    string
	Gid    string
	Muid   string
}

func (stat *Stat) String() string {
	return fmt.Sprintf("stype: %d, dev: %d, qid: [%s], mode: %o, atime: %d, mtime: %d, length: %d, name: %s, uid: %s, gid: %s, muid: %s",
		stat.Type, stat.Dev, &stat.Qid, stat.Mode,
		stat.Atime, stat.Mtime, stat.Length, stat.Name, stat.Uid,
		stat.Gid, stat.Muid)
}

func ParseStats(buff []byte) ([]Stat, error) {
	stats := make([]Stat, 0)
	var err error
	for len(buff) > 0 {
		s := Stat{}
		buff, err = s.parse(buff)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (stat *Stat) parse(buff []byte) ([]byte, error) {
	_, buff = fromLittleE16(buff) // throw away length
	stat.Type, buff = fromLittleE16(buff)
	stat.Dev, buff = fromLittleE32(buff)
	buff, err := stat.Qid.parse(buff)
	if err != nil {
		return nil, err
	}
	stat.Mode, buff = fromLittleE32(buff)
	stat.Atime, buff = fromLittleE32(buff)
	stat.Mtime, buff = fromLittleE32(buff)
	stat.Length, buff = fromLittleE64(buff)
	stat.Name, buff = fromString(buff)
	stat.Uid, buff = fromString(buff)
	stat.Gid, buff = fromString(buff)
	stat.Muid, buff = fromString(buff)
	return buff, nil
}

func (stat *Stat) ComposeLength() uint16 {
	// size[2], type[2], dev[4], qid[13], mode[4], atime[4], mtime[4], length[8],
	// name[s], uid[s], gid[s], muid[s]
	return uint16(2 + 2 + 4 + 13 + 4 + 4 + 4 + 8 +
		(2 + len(stat.Name)) +
		(2 + len(stat.Uid)) +
		(2 + len(stat.Gid)) +
		(2 + len(stat.Muid)))
}

func (stat *Stat) Compose() []byte {
	length := stat.ComposeLength()
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE16(length-2, buffer)
	buffer = toLittleE16(stat.Type, buffer)
	buffer = toLittleE32(stat.Dev, buffer)
	qidbuff := stat.Qid.Compose()
	copy(buffer, qidbuff)
	buffer = buffer[len(qidbuff):]
	buffer = toLittleE32(stat.Mode, buffer)
	buffer = toLittleE32(stat.Atime, buffer)
	buffer = toLittleE32(stat.Mtime, buffer)
	buffer = toLittleE64(stat.Length, buffer)
	buffer = toString(stat.Name, buffer)
	buffer = toString(stat.Uid, buffer)
	buffer = toString(stat.Gid, buffer)
	buffer = toString(stat.Muid, buffer)
	return buff
}

type RStat struct {
	Header
	Stat
}

func (stat *RStat) String() string {
	return fmt.Sprintf("rstat: [%s, %s]",
		&stat.Header, &stat.Stat)
}

func (stat *RStat) parse(buff []byte) ([]byte, error) {
	_, buff = fromLittleE16(buff) // stat length
	buff, err := stat.Stat.parse(buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (stat *RStat) Compose() []byte {
	// size[4] Rstat tag[2] stat[n]
	statLength := stat.Stat.ComposeLength()
	length := 4 + 1 + 2 + 2 + statLength
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = stat.Header.Type
	buffer = buffer[1:]
	buffer = toLittleE16(stat.Tag, buffer)
	buffer = toLittleE16(statLength, buffer)
	copy(buffer, stat.Stat.Compose())

	return buff
}
