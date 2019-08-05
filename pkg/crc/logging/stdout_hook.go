package logging

import (
	"io"
	"os"
	"runtime"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

// This is stdOutHook to send everything except error and Fatal to standard output.
type stdOutHook struct {
	stdout    io.Writer
	formatter logrus.Formatter
	level     logrus.Level
}

func newstdOutHook(level logrus.Level, formatter logrus.Formatter) *stdOutHook {
	// For windows to display colors we need to use the go-colorable writer
	if runtime.GOOS == "windows" {
		return &stdOutHook{
			stdout:    colorable.NewColorableStdout(),
			formatter: formatter,
			level:     level,
		}
	}
	return &stdOutHook{
		stdout:    os.Stdout,
		formatter: formatter,
		level:     level,
	}
}

func (h stdOutHook) Levels() []logrus.Level {
	var levels []logrus.Level
	for _, level := range logrus.AllLevels {
		if level <= h.level {
			// Ignore the Error and Fatal logs from stdout
			if level == logrus.ErrorLevel || level == logrus.FatalLevel {
				continue
			}
			levels = append(levels, level)
		}
	}

	return levels
}

func (h *stdOutHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.stdout.Write(line)
	return err
}
