package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// This is stdErrHook to send error to the stdErr.
type stdErrHook struct {
	stderr    io.Writer
	formatter logrus.Formatter
	level     logrus.Level
}

func newstdErrHook(level logrus.Level, formatter logrus.Formatter) *stdErrHook {
	return &stdErrHook{
		stderr:    os.Stderr,
		formatter: formatter,
		level:     level,
	}
}

func (h stdErrHook) Levels() []logrus.Level {
	var levels []logrus.Level
	for _, level := range logrus.AllLevels {
		if level <= h.level {
			// Only capture error and Fatal logs.
			if level == logrus.ErrorLevel || level == logrus.FatalLevel {
				levels = append(levels, level)
			}
		}
	}

	return levels
}

func (h *stdErrHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.stderr.Write(line)
	return err
}
