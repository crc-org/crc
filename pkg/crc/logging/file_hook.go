package logging

import (
	"io"

	"github.com/sirupsen/logrus"
)

// This is file Hook to send everything in the log file.
type fileHook struct {
	file      io.Writer
	formatter logrus.Formatter
	level     logrus.Level
}

func newFileHook(file io.Writer, level logrus.Level, formatter logrus.Formatter) *fileHook {
	return &fileHook{
		file:      file,
		formatter: formatter,
		level:     level,
	}
}

func (h fileHook) Levels() []logrus.Level {
	var levels []logrus.Level
	for _, level := range logrus.AllLevels {
		if level <= h.level {
			levels = append(levels, level)
		}
	}

	return levels
}

func (h *fileHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.file.Write(line)
	return err
}
