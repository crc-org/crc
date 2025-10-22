package p9

import (
	"io"

	"github.com/DeedleFake/p9/proto"
)

// ReadDir decodes a series of directory entries from a reader. It
// reads until EOF, so it doesn't return io.EOF as a possible error.
//
// It is recommended that the reader passed to ReadDir have some form
// of buffering, as some servers will silently mishandle attempts to
// read pieces of a directory. Wrapping the reader with a bufio.Reader
// is often sufficient.
func ReadDir(r io.Reader) ([]DirEntry, error) {
	var entries []DirEntry
	for {
		var stat Stat
		err := proto.Read(r, &stat)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return entries, err
		}

		entries = append(entries, stat.DirEntry())
	}
}

// WriteDir writes a series of directory entries to w.
func WriteDir(w io.Writer, entries []DirEntry) error {
	for _, entry := range entries {
		err := proto.Write(w, entry.Stat())
		if err != nil {
			return err
		}
	}

	return nil
}
