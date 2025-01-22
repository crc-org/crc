package utils

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher is an utility that
type FileWatcher struct {
	w    *fsnotify.Watcher
	path string

	writeGracePeriod time.Duration
	timer            *time.Timer
}

func NewFileWatcher(path string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{w: watcher, path: path, writeGracePeriod: 200 * time.Millisecond}, nil
}

func (fw *FileWatcher) Start(changeHandler func()) error {
	// Ensure that the target that we're watching is not a symlink as we won't get any events when we're watching
	// a symlink.
	fileRealPath, err := filepath.EvalSymlinks(fw.path)
	if err != nil {
		return fmt.Errorf("adding watcher failed: %s", err)
	}

	// watch the directory instead of the individual file to ensure the notification still works when the file is modified
	// through moving/renaming rather than writing into it directly (like what most modern editor does by default).
	// ref: https://github.com/fsnotify/fsnotify/blob/a9bc2e01792f868516acf80817f7d7d7b3315409/README.md#watching-a-file-doesnt-work-well
	if err = fw.w.Add(filepath.Dir(fileRealPath)); err != nil {
		return fmt.Errorf("adding watcher failed: %s", err)
	}

	go func() {
		for {
			select {
			case _, ok := <-fw.w.Errors:
				if !ok {
					return // watcher is closed.
				}
			case event, ok := <-fw.w.Events:
				if !ok {
					return // watcher is closed.
				}

				if event.Name != fileRealPath {
					continue // we don't care about this file.
				}

				// Create may not always followed by Write e.g. when we replace the file with mv.
				if event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Write) {
					// as per the documentation, receiving Write does not mean that the write is finished.
					// we try our best here to ignore "unfinished" write by assuming that after [writeGracePeriod] of
					// inactivity the write has been finished.
					fw.debounce(changeHandler)
				}
			}
		}
	}()

	return nil
}

func (fw *FileWatcher) debounce(fn func()) {
	if fw.timer != nil {
		fw.timer.Stop()
	}

	fw.timer = time.AfterFunc(fw.writeGracePeriod, fn)
}

func (fw *FileWatcher) Stop() error {
	return fw.w.Close()
}
