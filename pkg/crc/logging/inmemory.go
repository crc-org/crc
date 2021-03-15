package logging

import (
	"container/ring"

	"github.com/sirupsen/logrus"
)

// This hook keeps in memory n messages from error to info level
type inMemoryHook struct {
	messages *ring.Ring
}

func newInMemoryHook(size int) *inMemoryHook {
	return &inMemoryHook{
		messages: ring.New(size),
	}
}

func (h *inMemoryHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel}
}

func (h *inMemoryHook) Fire(entry *logrus.Entry) error {
	h.messages.Value = entry.Message
	h.messages = h.messages.Next()
	return nil
}

func (h *inMemoryHook) Messages() []string {
	var ret []string
	h.messages.Do(func(elem interface{}) {
		if str, ok := elem.(string); ok {
			ret = append(ret, str)
		}
	})
	return ret
}
