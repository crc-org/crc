package hosts

import (
	"errors"
	"runtime"
)

func (h *Hosts) saveHostsFile() error {
	if runtime.GOOS == "windows" {
		return h.File.SaveHostsFile()
	}

	if h.GoFileHandle == nil {
		return errors.New("cannot write to hosts file")
	}

	// Truncate the file to 0 bytes to ensure it's empty before writing.
	if err := h.GoFileHandle.Truncate(0); err != nil {
		return err
	}
	// Seek to the beginning of the file to overwrite the existing contents.
	if _, err := h.GoFileHandle.Seek(0, 0); err != nil {
		return err
	}
	dataBytes := []byte(h.File.RenderHostsFile())
	if _, err := h.GoFileHandle.Write(dataBytes); err != nil {
		return err
	}
	return h.GoFileHandle.Sync()
}
