package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/RedHatQE/gowinx/pkg/util"
	"github.com/sirupsen/logrus"
)

var (
	logfile       *os.File
	LogLevel      string
	originalHooks = logrus.LevelHooks{}
)

func OpenLogFile(basePath string, fileName string) (*os.File, error) {
	if err := util.EnsureBaseDirectoriesExist(basePath); err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(
		filepath.Join(basePath, fileName),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func CloseLogging() {
	logfile.Close()
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
}

func BackupLogFile() {
	if logfile == nil {
		return
	}
	os.Rename(logfile.Name(), fmt.Sprintf("%s_%s", logfile.Name(), time.Now().Format("20060102150405"))) // nolint
}

func InitLogrus(logLevel, basePath string, fileName string) {
	logrus.SetOutput(io.MultiWriter(os.Stdout))
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

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
