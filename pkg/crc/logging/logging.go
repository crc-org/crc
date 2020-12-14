package logging

import (
	"os"

	"github.com/sirupsen/logrus"
	terminal "golang.org/x/term"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	lumberLogger  = &lumberjack.Logger{}
	LogLevel      string
	originalHooks = logrus.LevelHooks{}
)

func InitLumberLogger(path string) *lumberjack.Logger {
	if lumberLogger.Filename == "" {
		lumberLogger = &lumberjack.Logger{
			Filename:   path,
			MaxSize:    1, // megabytes
			MaxBackups: 1,
			MaxAge:     10, // days
		}
	}
	return lumberLogger
}

func CloseLumberLogger() {
	lumberLogger.Close()
}

func CloseLogging() {
	CloseLumberLogger()
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
}

func InitLogrus(logLevel, logFilePath string) {
	InitLumberLogger(logFilePath)

	// send logs to file
	logrus.SetOutput(lumberLogger)

	logrus.SetLevel(logrus.TraceLevel)

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}

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
