package fs

import (
	"github.com/knusbaum/go9p/proto"
)

// DynamicFile is a File implementation that will serve dynamic content.
// Every time a client opens the DynamicFile, content is generated for
// the opening fid with the function genContent, which is passed to NewDynamicFile.
// Subsequent reads on the fid will return byte ranges from that content
// generated when the fid was opened. Closing the fid releases the content.
type DynamicFile struct {
	BaseFile
	fidContent map[uint64][]byte
	genContent func() []byte
}

// NewDynamicFile creates a new DynamicFile that will use getContent to
// generate the file's content for each fid that opens it.
func NewDynamicFile(s *proto.Stat, genContent func() []byte) *DynamicFile {
	return &DynamicFile{
		BaseFile:   BaseFile{fStat: *s},
		fidContent: make(map[uint64][]byte),
		genContent: genContent,
	}
}

func (f *DynamicFile) Open(fid uint64, omode proto.Mode) error {
	f.Lock()
	defer f.Unlock()
	f.fidContent[fid] = f.genContent()
	return nil
}

func (f *DynamicFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	f.RLock()
	defer f.RUnlock()

	data := f.fidContent[fid]

	flen := uint64(len(data))
	if offset >= flen {
		return []byte{}, nil
	}
	if offset+count > flen {
		count = flen - offset
	}
	return data[offset : offset+count], nil
}

func (f *DynamicFile) Close(fid uint64) error {
	delete(f.fidContent, fid)
	return nil
}

// WrappedFile takes an existing File and adds optional hooks for the
// Open, Read, Write, and Close functions.
// OpenF, ReadF, WriteF, and CloseF, if set, are called rather than
// the File's Open, Read, Write, and Close functions.
type WrappedFile struct {
	File
	OpenF  func(fid uint64, omode proto.Mode) error
	ReadF  func(fid uint64, offset uint64, count uint64) ([]byte, error)
	WriteF func(fid uint64, offset uint64, data []byte) (uint32, error)
	CloseF func(fid uint64) error
}

func (f *WrappedFile) Open(fid uint64, omode proto.Mode) error {
	if f.OpenF != nil {
		return f.OpenF(fid, omode)
	}
	return f.File.Open(fid, omode)
}

func (f *WrappedFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	if f.ReadF != nil {
		return f.ReadF(fid, offset, count)
	}
	return f.File.Read(fid, offset, count)
}

func (f *WrappedFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	if f.WriteF != nil {
		return f.WriteF(fid, offset, data)
	}
	return f.File.Write(fid, offset, data)
}

func (f *WrappedFile) Close(fid uint64) error {
	if f.CloseF != nil {
		return f.CloseF(fid)
	}
	return f.File.Close(fid)
}
