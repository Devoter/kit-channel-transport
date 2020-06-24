package channel

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// HandlerFunc is a function that handles a request event by an endpoint
type HandlerFunc func(event *RequestEvent)

// MakeHandlerFunc returns a HandlerFunc
func MakeHandlerFunc(e endpoint.Endpoint) HandlerFunc {
	return func(event *RequestEvent) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		response, err := e(ctx, event.Body)
		if err != nil {
			event.ResponseChan <- ResponseEvent{Err: err}
			return
		}

		event.ResponseChan <- ResponseEvent{Message: Message{Body: response}}
	}
}
