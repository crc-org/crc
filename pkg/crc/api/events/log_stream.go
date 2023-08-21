package events

import (
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/r3labs/sse/v2"
	"github.com/sirupsen/logrus"
)

type streamHook struct {
	server    *sse.Server
	formatter logrus.Formatter
	level     logrus.Level
}

type logsStream struct {
	hasInitialized bool
	server         *EventServer
}

func newSSEStreamHook(server *sse.Server) *streamHook {
	return &streamHook{
		server,
		&logrus.JSONFormatter{
			TimestampFormat:   "",
			DisableTimestamp:  false,
			DisableHTMLEscape: false,
			DataKey:           "",
			FieldMap:          nil,
			CallerPrettyfier:  nil,
			PrettyPrint:       false,
		},
		logging.DefaultLogLevel(),
	}
}

func newLogsStream(server *EventServer) EventStream {
	return &logsStream{
		hasInitialized: false,
		server:         server,
	}
}

func (s *streamHook) Levels() []logrus.Level {
	var levels []logrus.Level
	for _, level := range logrus.AllLevels {
		if level <= s.level {
			levels = append(levels, level)
		}
	}
	return levels
}

func (s *streamHook) Fire(entry *logrus.Entry) error {
	line, err := s.formatter.Format(entry)
	if err != nil {
		return err
	}

	s.server.Publish(LOGS, &sse.Event{Event: []byte(LOGS), Data: line})
	return nil
}

func (l *logsStream) AddSubscriber(_ *sse.Subscriber) {
	if !l.hasInitialized {
		logrus.AddHook(newSSEStreamHook(l.server.sseServer))
		l.hasInitialized = true
	}

}

func (l *logsStream) RemoveSubscriber(_ *sse.Subscriber) {
	// do nothing as we could not remove log listener
}
