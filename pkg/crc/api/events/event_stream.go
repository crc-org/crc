package events

import (
	"sync"

	"github.com/r3labs/sse/v2"
)

type EventStream interface {
	AddSubscriber(subscriber *sse.Subscriber)
	RemoveSubscriber(subscriber *sse.Subscriber)
}

type eventStream struct {
	subscribers map[*sse.Subscriber]interface{}
	producer    EventProducer
	publisher   EventPublisher
	streamMutex sync.Mutex
}

func newStream(producer EventProducer, publisher EventPublisher) EventStream {
	return &eventStream{
		subscribers: map[*sse.Subscriber]interface{}{},
		producer:    producer,
		publisher:   publisher,
	}
}

func (es *eventStream) AddSubscriber(subscriber *sse.Subscriber) {
	es.streamMutex.Lock()
	defer es.streamMutex.Unlock()

	es.subscribers[subscriber] = struct{}{}
	if len(es.subscribers) == 1 {
		es.producer.Start(es.publisher)
	}
}
func (es *eventStream) RemoveSubscriber(subscriber *sse.Subscriber) {
	es.streamMutex.Lock()
	defer es.streamMutex.Unlock()

	delete(es.subscribers, subscriber)
	if len(es.subscribers) == 0 {
		es.producer.Stop()
	}
}
