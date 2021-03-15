package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/sirupsen/logrus"
	terminal "golang.org/x/term"
)

var (
	logfile       *os.File
	LogLevel      string
	originalHooks = logrus.LevelHooks{}
	Memory        = newInMemoryHook(100)
)

func OpenLogFile(path string) (*os.File, error) {
	err := constants.EnsureBaseDirectoriesExist()
	if err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func CloseLogging() {
	logfile.Close()
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
}

func BackupLogFile() error {
	if logfile == nil {
		return nil
	}
	return os.Rename(logfile.Name(), fmt.Sprintf("%s_%s", logfile.Name(), time.Now().Format("20060102150405")))
}

func InitLogrus(logLevel, logFilePath string) {
	var err error
	logfile, err = OpenLogFile(logFilePath)
	if err != nil {
		logrus.Fatal("Unable to open log file: ", err)
	}
	// send logs to file
	logrus.SetOutput(logfile)

	logrus.SetLevel(logrus.TraceLevel)

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}

	logrus.AddHook(Memory)

	// Add hook to send error/fatal to stderr
	logrus.AddHook(newstdErrHook(level, &logrus.TextFormatter{
		ForceColors:            terminal.IsTerminal(int(os.Stderr.Fd())),
		DisableTimestamp:       true,
		DisableLevelTruncation: false,
	}))

	for k, v := range logrus.StandardLogger().Hooks {
		originalHooks[k] = v
	}
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(s string, args ...interface{}) {
	logrus.Infof(s, args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warnf(s string, args ...interface{}) {
	logrus.Warnf(s, args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Fatalf(s string, args ...interface{}) {
	logrus.Fatalf(s, args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Errorf(s string, args ...interface{}) {
	logrus.Errorf(s, args...)
}

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Debugf(s string, args ...interface{}) {
	logrus.Debugf(s, args...)
}
