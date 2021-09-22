package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	terminal "golang.org/x/term"
)

var (
	logfile       *os.File
	logLevel      = defaultLogLevel()
	originalHooks = logrus.LevelHooks{}
	Memory        = newInMemoryHook(100)
)

func OpenLogFile(path string) (*os.File, error) {
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

func BackupLogFile() {
	if logfile == nil {
		return
	}
	os.Rename(logfile.Name(), fmt.Sprintf("%s_%s", logfile.Name(), time.Now().Format("20060102150405"))) // nolint
}

func InitLogrus(logFilePath string) {
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

func defaultLogLevel() string {
	defaultLevel := "info"
	envLogLevel := os.Getenv("CRC_LOG_LEVEL")
	if envLogLevel != "" {
		defaultLevel = envLogLevel
	}

	return defaultLevel
}

func AddLogLevelFlag(flagset *pflag.FlagSet) {
	flagset.StringVar(&logLevel, "log-level", defaultLogLevel(), "log level (e.g. \"debug | info | warn | error\")")
}

func IsDebug() bool {
	return logLevel == "debug"
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
