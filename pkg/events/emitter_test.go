package events

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestEventData struct {
	Foo string
}

type eventListener struct {
	notifyCb func(e TestEventData)
}

func newEventListener(notifyCb func(e TestEventData)) *eventListener {
	return &eventListener{notifyCb: notifyCb}
}

func (listener *eventListener) Notify(e TestEventData) {
	listener.notifyCb(e)
}

func TestEmitter(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var data string

	event := NewEvent[TestEventData]()
	event.AddListener(newEventListener(func(e TestEventData) {
		data = e.Foo
		wg.Done()
	}))
	event.Fire(TestEventData{Foo: "bar"})

	wg.Wait()

	assert.Equal(t, "bar", data)
}
