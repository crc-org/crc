package utils

import (
	"os"
	"time"
)

// FileWatcher is an utility that
type FileWatcher struct {
	path string

	closeCh      chan struct{}
	pollInterval time.Duration
}

func NewFileWatcher(path string) *FileWatcher {
	return &FileWatcher{
		path:         path,
		pollInterval: 5 * time.Second, // 5s is the default inode cache timeout in linux for most systems.
		closeCh:      make(chan struct{}),
	}
}

func (fw *FileWatcher) Start(changeHandler func()) {
	prevModTime := fw.fileModTime(fw.path)

	// use polling-based approach to detect file changes
	// we can't use fsnotify/fsnotify due to issues with symlink+socket. see #462.
	go func() {
		for {
			select {
			case _, ok := <-fw.closeCh:
				if !ok {
					return // watcher is closed.
				}
			case <-time.After(fw.pollInterval):
			}

			modTime := fw.fileModTime(fw.path)
			if modTime.IsZero() {
				continue // file does not exists
			}

			if !prevModTime.Equal(modTime) {
				changeHandler()
				prevModTime = modTime
			}
		}
	}()
}

func (fw *FileWatcher) fileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}

	return info.ModTime()
}

func (fw *FileWatcher) Stop() {
	close(fw.closeCh)
}
