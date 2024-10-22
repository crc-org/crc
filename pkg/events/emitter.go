package events

import (
	"sync"
)

type Event[T any] interface {
	AddListener(listener Notifiable[T])
	RemoveListener(listener Notifiable[T])
	Fire(data T)
}

type Notifiable[T any] interface {
	Notify(event T)
}

type event[T any] struct {
	listeners  map[Notifiable[T]]Notifiable[T]
	eventMutex sync.Mutex
}

func NewEvent[T any]() Event[T] {
	return &event[T]{
		listeners: make(map[Notifiable[T]]Notifiable[T]),
	}
}

func (e *event[T]) AddListener(listener Notifiable[T]) {
	e.eventMutex.Lock()
	defer e.eventMutex.Unlock()
	e.listeners[listener] = listener
}

func (e *event[T]) RemoveListener(listener Notifiable[T]) {
	e.eventMutex.Lock()
	defer e.eventMutex.Unlock()
	delete(e.listeners, listener)
}

func (e *event[T]) Fire(event T) {
	e.eventMutex.Lock()
	defer e.eventMutex.Unlock()
	for _, listener := range e.listeners {
		// shadowing for loop variable, need to remove after golang 1.22 migration
		listener := listener
		go listener.Notify(event)
	}
}
