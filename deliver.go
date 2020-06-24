package channel

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// DeliverFunc is a function that is used to publish events to a Bus instance.
type DeliverFunc func(context.Context, *Bus, string, *Message, chan ResponseEvent) (interface{}, error)

// MakeClientEndpoint returns a client endpoint which associated with a key.
func MakeClientEndpoint(bus *Bus, key string, deliver DeliverFunc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ch := make(chan ResponseEvent)
		msg := Message{Body: request}

		resp, err := deliver(ctx, bus, key, &msg, ch)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

// SyncDeliver is a delivery function that publishes a request and awaits a response.
func SyncDeliver(ctx context.Context, bus *Bus, key string, msg *Message, channel chan ResponseEvent) (interface{}, error) {
	bus.Publish(key, msg.Body, channel)

	for {
		select {
		case resp := <-channel:
			return &resp.Body, resp.Err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// AsyncDeliver is a delivery function that only publishes a request and doesn't wait for a response.
func AsyncDeliver(ctx context.Context, bus *Bus, key string, msg *Message) (interface{}, error) {
	bus.Publish(key, msg.Body, nil)

	return nil, nil
}
