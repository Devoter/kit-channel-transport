package channel

import (
	"sync"
)

// Message is a structure which contains a request or response body.
type Message struct {
	Body interface{}
}

// ResponseEvent is a structure which contains a response data body or/and error.
type ResponseEvent struct {
	Message
	Err error
}

// RequestEvent is a structure which contains a request data bode and optionally a response channel.
type RequestEvent struct {
	Message
	ResponseChan chan<- ResponseEvent
}

// Bus is an event bus.
type Bus struct {
	mx       sync.RWMutex
	channels map[string][]chan<- RequestEvent
}

// NewBus returns an initialized Bus instance.
func NewBus() *Bus {
	return &Bus{
		channels: make(map[string][]chan<- RequestEvent),
	}
}

// Subscribe associates a requests channel to the specified key string.
func (b *Bus) Subscribe(key string, ch chan<- RequestEvent) {
	defer b.mx.Unlock()

	b.mx.Lock()

	channels, ok := b.channels[key]
	if !ok {
		channels = []chan<- RequestEvent{}
	}

	for _, c := range channels {
		if c == ch {
			return
		}
	}

	b.channels[key] = append(channels, ch)
}

// Unsubscribe removes an associated requests channel from the channels list.
func (b *Bus) Unsubscribe(key string, ch chan<- RequestEvent) {
	defer b.mx.Unlock()

	b.mx.Lock()

	channels, ok := b.channels[key]
	if !ok {
		return
	}

	newChannels := make([]chan<- RequestEvent, len(channels)-1)
	found := 0

	for i, c := range channels {
		if c == ch {
			found = 1
		} else {
			newChannels[i+found] = c
		}
	}

	if found == 0 {
		return
	}

	if len(newChannels) == 0 {
		delete(b.channels, key)
	} else {
		b.channels[key] = newChannels
	}
}

// Publish sends a request event to associated channels.
func (b *Bus) Publish(key string, request interface{}, responseChan chan<- ResponseEvent) {
	channels, ok := b.channels[key]
	if !ok {
		return
	}

	for _, c := range channels {
		c <- RequestEvent{Message: Message{Body: request}, ResponseChan: responseChan}
	}
}
