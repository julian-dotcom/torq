package broadcast

import (
	"context"
)

type BroadcastServer interface {
	Subscribe() <-chan interface{}
	CancelSubscription(<-chan interface{})
}

type broadcastServer struct {
	source         <-chan interface{}
	listeners      []chan interface{}
	addListener    chan chan interface{}
	removeListener chan (<-chan interface{})
}

func (s *broadcastServer) Subscribe() <-chan interface{} {
	newListener := make(chan interface{})
	s.addListener <- newListener
	return newListener
}

func (s *broadcastServer) CancelSubscription(channel <-chan interface{}) {
	s.removeListener <- channel
}

func NewBroadcastServer(ctx context.Context, source <-chan interface{}) BroadcastServer {
	service := &broadcastServer{
		source:         source,
		listeners:      make([]chan interface{}, 0),
		addListener:    make(chan chan interface{}),
		removeListener: make(chan (<-chan interface{})),
	}
	go service.serve(ctx)
	return service
}

func (s *broadcastServer) serve(ctx context.Context) {
	defer func() {
		for _, listener := range s.listeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addListener:
			s.listeners = append(s.listeners, newListener)
		case listenerToRemove := <-s.removeListener:
			for i, ch := range s.listeners {
				if ch == listenerToRemove {
					s.listeners[i] = s.listeners[len(s.listeners)-1]
					s.listeners = s.listeners[:len(s.listeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.source:
			if !ok {
				return
			}
			for _, listener := range s.listeners {
				if listener != nil {
					select {
					case listener <- val:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}
