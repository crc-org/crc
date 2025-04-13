package proto

import "fmt"

type TWstat struct {
	Header
	Fid  uint32
	Stat Stat
}

func (wstat *TWstat) String() string {
	return fmt.Sprintf("twstat: [%s, fid: %d, %s]",
		&wstat.Header, wstat.Fid, &wstat.Stat)
}

func (wstat *TWstat) parse(buff []byte) ([]byte, error) {
	wstat.Fid, buff = fromLittleE32(buff)
	_, buff = fromLittleE16(buff) // Throw away stat length.
	buff, err := wstat.Stat.parse(buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (wstat *TWstat) Compose() []byte {
	// size[4] Twstat tag[2] fid[4] stat[n]
	statLength := wstat.Stat.ComposeLength()
	length := 4 + 1 + 2 + 4 + 2 + statLength
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = wstat.Type
	buffer = buffer[1:]
	buffer = toLittleE16(wstat.Tag, buffer)
	buffer = toLittleE32(wstat.Fid, buffer)
	buffer = toLittleE16(statLength, buffer)
	copy(buffer, wstat.Stat.Compose())

	return buff
}

type RWstat struct {
	Header
}

func (wstat *RWstat) String() string {
	return fmt.Sprintf("rwstat: [%s]", &wstat.Header)
}

func (wstat *RWstat) parse(buff []byte) ([]byte, error) {
	return buff, nil
}

func (wstat *RWstat) Compose() []byte {
	// size[4] Rwstat tag[2]
	length := 4 + 1 + 2
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = wstat.Type
	buffer = buffer[1:]
	buffer = toLittleE16(wstat.Tag, buffer)
	return buff
}
