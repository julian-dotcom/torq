package broadcast

import (
	"context"

	"github.com/lncapital/torq/pkg/commons"
)

type BroadcastServer interface {
	SubscribeServiceEvent() <-chan commons.ServiceEvent
	CancelSubscriptionServiceEvent(<-chan commons.ServiceEvent)
	SubscribeHtlcEvent() <-chan commons.HtlcEvent
	CancelSubscriptionHtlcEvent(<-chan commons.HtlcEvent)
	SubscribeForwardEvent() <-chan commons.ForwardEvent
	CancelSubscriptionForwardEvent(<-chan commons.ForwardEvent)
	SubscribeChannelBalanceEvent() <-chan commons.ChannelBalanceEvent
	CancelSubscriptionChannelBalanceEvent(<-chan commons.ChannelBalanceEvent)
	SubscribeChannelEvent() <-chan commons.ChannelEvent
	CancelSubscriptionChannelEvent(<-chan commons.ChannelEvent)
	SubscribeNodeGraphEvent() <-chan commons.NodeGraphEvent
	CancelSubscriptionNodeGraphEvent(<-chan commons.NodeGraphEvent)
	SubscribeChannelGraphEvent() <-chan commons.ChannelGraphEvent
	CancelSubscriptionChannelGraphEvent(<-chan commons.ChannelGraphEvent)
	SubscribeInvoiceEvent() <-chan commons.InvoiceEvent
	CancelSubscriptionInvoiceEvent(<-chan commons.InvoiceEvent)
	SubscribePaymentEvent() <-chan commons.PaymentEvent
	CancelSubscriptionPaymentEvent(<-chan commons.PaymentEvent)
	SubscribeTransactionEvent() <-chan commons.TransactionEvent
	CancelSubscriptionTransactionEvent(<-chan commons.TransactionEvent)
	SubscribePeerEvent() <-chan commons.PeerEvent
	CancelSubscriptionPeerEvent(<-chan commons.PeerEvent)
	SubscribeBlockEvent() <-chan commons.BlockEvent
	CancelSubscriptionBlockEvent(<-chan commons.BlockEvent)
	SubscribeWebSocketResponse() <-chan interface{}
	CancelSubscriptionWebSocketResponse(<-chan interface{})
}

type broadcastServer struct {
	serviceEventSource         <-chan commons.ServiceEvent
	serviceEventListeners      []chan commons.ServiceEvent
	addServiceEventListener    chan chan commons.ServiceEvent
	removeServiceEventListener chan (<-chan commons.ServiceEvent)

	htlcEventSource         <-chan commons.HtlcEvent
	htlcEventListeners      []chan commons.HtlcEvent
	addHtlcEventListener    chan chan commons.HtlcEvent
	removeHtlcEventListener chan (<-chan commons.HtlcEvent)

	forwardEventSource         <-chan commons.ForwardEvent
	forwardEventListeners      []chan commons.ForwardEvent
	addForwardEventListener    chan chan commons.ForwardEvent
	removeForwardEventListener chan (<-chan commons.ForwardEvent)

	channelBalanceEventSource         <-chan commons.ChannelBalanceEvent
	channelBalanceEventListeners      []chan commons.ChannelBalanceEvent
	addChannelBalanceEventListener    chan chan commons.ChannelBalanceEvent
	removeChannelBalanceEventListener chan (<-chan commons.ChannelBalanceEvent)

	channelEventSource         <-chan commons.ChannelEvent
	channelEventListeners      []chan commons.ChannelEvent
	addChannelEventListener    chan chan commons.ChannelEvent
	removeChannelEventListener chan (<-chan commons.ChannelEvent)

	nodeGraphEventSource         <-chan commons.NodeGraphEvent
	nodeGraphEventListeners      []chan commons.NodeGraphEvent
	addNodeGraphEventListener    chan chan commons.NodeGraphEvent
	removeNodeGraphEventListener chan (<-chan commons.NodeGraphEvent)

	channelGraphEventSource         <-chan commons.ChannelGraphEvent
	channelGraphEventListeners      []chan commons.ChannelGraphEvent
	addChannelGraphEventListener    chan chan commons.ChannelGraphEvent
	removeChannelGraphEventListener chan (<-chan commons.ChannelGraphEvent)

	invoiceEventSource         <-chan commons.InvoiceEvent
	invoiceEventListeners      []chan commons.InvoiceEvent
	addInvoiceEventListener    chan chan commons.InvoiceEvent
	removeInvoiceEventListener chan (<-chan commons.InvoiceEvent)

	paymentEventSource         <-chan commons.PaymentEvent
	paymentEventListeners      []chan commons.PaymentEvent
	addPaymentEventListener    chan chan commons.PaymentEvent
	removePaymentEventListener chan (<-chan commons.PaymentEvent)

	transactionEventSource         <-chan commons.TransactionEvent
	transactionEventListeners      []chan commons.TransactionEvent
	addTransactionEventListener    chan chan commons.TransactionEvent
	removeTransactionEventListener chan (<-chan commons.TransactionEvent)

	peerEventSource         <-chan commons.PeerEvent
	peerEventListeners      []chan commons.PeerEvent
	addPeerEventListener    chan chan commons.PeerEvent
	removePeerEventListener chan (<-chan commons.PeerEvent)

	blockEventSource         <-chan commons.BlockEvent
	blockEventListeners      []chan commons.BlockEvent
	addBlockEventListener    chan chan commons.BlockEvent
	removeBlockEventListener chan (<-chan commons.BlockEvent)

	webSocketResponseSource         <-chan interface{}
	webSocketResponseListeners      []chan interface{}
	addWebSocketResponseListener    chan chan interface{}
	removeWebSocketResponseListener chan (<-chan interface{})
}

func (s *broadcastServer) SubscribeServiceEvent() <-chan commons.ServiceEvent {
	newListener := make(chan commons.ServiceEvent)
	s.addServiceEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionServiceEvent(channel <-chan commons.ServiceEvent) {
	s.removeServiceEventListener <- channel
}

func (s *broadcastServer) SubscribeHtlcEvent() <-chan commons.HtlcEvent {
	newListener := make(chan commons.HtlcEvent)
	s.addHtlcEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionHtlcEvent(channel <-chan commons.HtlcEvent) {
	s.removeHtlcEventListener <- channel
}

func (s *broadcastServer) SubscribeForwardEvent() <-chan commons.ForwardEvent {
	newListener := make(chan commons.ForwardEvent)
	s.addForwardEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionForwardEvent(channel <-chan commons.ForwardEvent) {
	s.removeForwardEventListener <- channel
}

func (s *broadcastServer) SubscribeChannelBalanceEvent() <-chan commons.ChannelBalanceEvent {
	newListener := make(chan commons.ChannelBalanceEvent)
	s.addChannelBalanceEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionChannelBalanceEvent(channel <-chan commons.ChannelBalanceEvent) {
	s.removeChannelBalanceEventListener <- channel
}

func (s *broadcastServer) SubscribeChannelEvent() <-chan commons.ChannelEvent {
	newListener := make(chan commons.ChannelEvent)
	s.addChannelEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionChannelEvent(channel <-chan commons.ChannelEvent) {
	s.removeChannelEventListener <- channel
}

func (s *broadcastServer) SubscribeNodeGraphEvent() <-chan commons.NodeGraphEvent {
	newListener := make(chan commons.NodeGraphEvent)
	s.addNodeGraphEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionNodeGraphEvent(channel <-chan commons.NodeGraphEvent) {
	s.removeNodeGraphEventListener <- channel
}

func (s *broadcastServer) SubscribeChannelGraphEvent() <-chan commons.ChannelGraphEvent {
	newListener := make(chan commons.ChannelGraphEvent)
	s.addChannelGraphEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionChannelGraphEvent(channel <-chan commons.ChannelGraphEvent) {
	s.removeChannelGraphEventListener <- channel
}

func (s *broadcastServer) SubscribeInvoiceEvent() <-chan commons.InvoiceEvent {
	newListener := make(chan commons.InvoiceEvent)
	s.addInvoiceEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionInvoiceEvent(channel <-chan commons.InvoiceEvent) {
	s.removeInvoiceEventListener <- channel
}

func (s *broadcastServer) SubscribePaymentEvent() <-chan commons.PaymentEvent {
	newListener := make(chan commons.PaymentEvent)
	s.addPaymentEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionPaymentEvent(channel <-chan commons.PaymentEvent) {
	s.removePaymentEventListener <- channel
}

func (s *broadcastServer) SubscribeTransactionEvent() <-chan commons.TransactionEvent {
	newListener := make(chan commons.TransactionEvent)
	s.addTransactionEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionTransactionEvent(channel <-chan commons.TransactionEvent) {
	s.removeTransactionEventListener <- channel
}

func (s *broadcastServer) SubscribePeerEvent() <-chan commons.PeerEvent {
	newListener := make(chan commons.PeerEvent)
	s.addPeerEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionPeerEvent(channel <-chan commons.PeerEvent) {
	s.removePeerEventListener <- channel
}

func (s *broadcastServer) SubscribeBlockEvent() <-chan commons.BlockEvent {
	newListener := make(chan commons.BlockEvent)
	s.addBlockEventListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionBlockEvent(channel <-chan commons.BlockEvent) {
	s.removeBlockEventListener <- channel
}

func (s *broadcastServer) SubscribeWebSocketResponse() <-chan interface{} {
	newListener := make(chan interface{})
	s.addWebSocketResponseListener <- newListener
	return newListener
}
func (s *broadcastServer) CancelSubscriptionWebSocketResponse(channel <-chan interface{}) {
	s.removeWebSocketResponseListener <- channel
}

func NewBroadcastServer(ctx context.Context,
	serviceEventSource <-chan commons.ServiceEvent,
	htlcEventSource <-chan commons.HtlcEvent,
	forwardEventSource <-chan commons.ForwardEvent,
	channelBalanceEventSource <-chan commons.ChannelBalanceEvent,
	channelEventSource <-chan commons.ChannelEvent,
	nodeGraphEventSource <-chan commons.NodeGraphEvent,
	channelGraphEventSource <-chan commons.ChannelGraphEvent,
	invoiceEventSource <-chan commons.InvoiceEvent,
	paymentEventSource <-chan commons.PaymentEvent,
	transactionEventSource <-chan commons.TransactionEvent,
	peerEventSource <-chan commons.PeerEvent,
	blockEventSource <-chan commons.BlockEvent,
	webSocketResponseSource <-chan interface{}) BroadcastServer {
	service := &broadcastServer{
		serviceEventSource:         serviceEventSource,
		serviceEventListeners:      make([]chan commons.ServiceEvent, 0),
		addServiceEventListener:    make(chan chan commons.ServiceEvent),
		removeServiceEventListener: make(chan (<-chan commons.ServiceEvent)),

		htlcEventSource:         htlcEventSource,
		htlcEventListeners:      make([]chan commons.HtlcEvent, 0),
		addHtlcEventListener:    make(chan chan commons.HtlcEvent),
		removeHtlcEventListener: make(chan (<-chan commons.HtlcEvent)),

		forwardEventSource:         forwardEventSource,
		forwardEventListeners:      make([]chan commons.ForwardEvent, 0),
		addForwardEventListener:    make(chan chan commons.ForwardEvent),
		removeForwardEventListener: make(chan (<-chan commons.ForwardEvent)),

		channelBalanceEventSource:         channelBalanceEventSource,
		channelBalanceEventListeners:      make([]chan commons.ChannelBalanceEvent, 0),
		addChannelBalanceEventListener:    make(chan chan commons.ChannelBalanceEvent),
		removeChannelBalanceEventListener: make(chan (<-chan commons.ChannelBalanceEvent)),

		channelEventSource:         channelEventSource,
		channelEventListeners:      make([]chan commons.ChannelEvent, 0),
		addChannelEventListener:    make(chan chan commons.ChannelEvent),
		removeChannelEventListener: make(chan (<-chan commons.ChannelEvent)),

		nodeGraphEventSource:         nodeGraphEventSource,
		nodeGraphEventListeners:      make([]chan commons.NodeGraphEvent, 0),
		addNodeGraphEventListener:    make(chan chan commons.NodeGraphEvent),
		removeNodeGraphEventListener: make(chan (<-chan commons.NodeGraphEvent)),

		channelGraphEventSource:         channelGraphEventSource,
		channelGraphEventListeners:      make([]chan commons.ChannelGraphEvent, 0),
		addChannelGraphEventListener:    make(chan chan commons.ChannelGraphEvent),
		removeChannelGraphEventListener: make(chan (<-chan commons.ChannelGraphEvent)),

		invoiceEventSource:         invoiceEventSource,
		invoiceEventListeners:      make([]chan commons.InvoiceEvent, 0),
		addInvoiceEventListener:    make(chan chan commons.InvoiceEvent),
		removeInvoiceEventListener: make(chan (<-chan commons.InvoiceEvent)),

		paymentEventSource:         paymentEventSource,
		paymentEventListeners:      make([]chan commons.PaymentEvent, 0),
		addPaymentEventListener:    make(chan chan commons.PaymentEvent),
		removePaymentEventListener: make(chan (<-chan commons.PaymentEvent)),

		transactionEventSource:         transactionEventSource,
		transactionEventListeners:      make([]chan commons.TransactionEvent, 0),
		addTransactionEventListener:    make(chan chan commons.TransactionEvent),
		removeTransactionEventListener: make(chan (<-chan commons.TransactionEvent)),

		peerEventSource:         peerEventSource,
		peerEventListeners:      make([]chan commons.PeerEvent, 0),
		addPeerEventListener:    make(chan chan commons.PeerEvent),
		removePeerEventListener: make(chan (<-chan commons.PeerEvent)),

		blockEventSource:         blockEventSource,
		blockEventListeners:      make([]chan commons.BlockEvent, 0),
		addBlockEventListener:    make(chan chan commons.BlockEvent),
		removeBlockEventListener: make(chan (<-chan commons.BlockEvent)),

		webSocketResponseSource:         webSocketResponseSource,
		webSocketResponseListeners:      make([]chan interface{}, 0),
		addWebSocketResponseListener:    make(chan chan interface{}),
		removeWebSocketResponseListener: make(chan (<-chan interface{})),
	}
	go service.serveServiceEvent(ctx)
	go service.serveHtlcEvent(ctx)
	go service.serveForwardEvent(ctx)
	go service.serveChannelBalanceEvent(ctx)
	go service.serveChannelEvent(ctx)
	go service.serveNodeGraphEvent(ctx)
	go service.serveChannelGraphEvent(ctx)
	go service.serveInvoiceEvent(ctx)
	go service.servePaymentEvent(ctx)
	go service.serveTransactionEvent(ctx)
	go service.servePeerEvent(ctx)
	go service.serveBlockEvent(ctx)
	go service.serveWebSocketResponse(ctx)
	return service
}

func (s *broadcastServer) serveServiceEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.serviceEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addServiceEventListener:
			s.serviceEventListeners = append(s.serviceEventListeners, newListener)
		case listenerToRemove := <-s.removeServiceEventListener:
			for i, ch := range s.serviceEventListeners {
				if ch == listenerToRemove {
					s.serviceEventListeners[i] = s.serviceEventListeners[len(s.serviceEventListeners)-1]
					s.serviceEventListeners = s.serviceEventListeners[:len(s.serviceEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.serviceEventSource:
			if !ok {
				return
			}
			for _, listener := range s.serviceEventListeners {
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

func (s *broadcastServer) serveHtlcEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.htlcEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addHtlcEventListener:
			s.htlcEventListeners = append(s.htlcEventListeners, newListener)
		case listenerToRemove := <-s.removeHtlcEventListener:
			for i, ch := range s.htlcEventListeners {
				if ch == listenerToRemove {
					s.htlcEventListeners[i] = s.htlcEventListeners[len(s.htlcEventListeners)-1]
					s.htlcEventListeners = s.htlcEventListeners[:len(s.htlcEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.htlcEventSource:
			if !ok {
				return
			}
			for _, listener := range s.htlcEventListeners {
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

func (s *broadcastServer) serveForwardEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.forwardEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addForwardEventListener:
			s.forwardEventListeners = append(s.forwardEventListeners, newListener)
		case listenerToRemove := <-s.removeForwardEventListener:
			for i, ch := range s.forwardEventListeners {
				if ch == listenerToRemove {
					s.forwardEventListeners[i] = s.forwardEventListeners[len(s.forwardEventListeners)-1]
					s.forwardEventListeners = s.forwardEventListeners[:len(s.forwardEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.forwardEventSource:
			if !ok {
				return
			}
			for _, listener := range s.forwardEventListeners {
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

func (s *broadcastServer) serveChannelBalanceEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.channelBalanceEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addChannelBalanceEventListener:
			s.channelBalanceEventListeners = append(s.channelBalanceEventListeners, newListener)
		case listenerToRemove := <-s.removeChannelBalanceEventListener:
			for i, ch := range s.channelBalanceEventListeners {
				if ch == listenerToRemove {
					s.channelBalanceEventListeners[i] = s.channelBalanceEventListeners[len(s.channelBalanceEventListeners)-1]
					s.channelBalanceEventListeners = s.channelBalanceEventListeners[:len(s.channelBalanceEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.channelBalanceEventSource:
			if !ok {
				return
			}
			for _, listener := range s.channelBalanceEventListeners {
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

func (s *broadcastServer) serveChannelEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.channelEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addChannelEventListener:
			s.channelEventListeners = append(s.channelEventListeners, newListener)
		case listenerToRemove := <-s.removeChannelEventListener:
			for i, ch := range s.channelEventListeners {
				if ch == listenerToRemove {
					s.channelEventListeners[i] = s.channelEventListeners[len(s.channelEventListeners)-1]
					s.channelEventListeners = s.channelEventListeners[:len(s.channelEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.channelEventSource:
			if !ok {
				return
			}
			for _, listener := range s.channelEventListeners {
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

func (s *broadcastServer) serveNodeGraphEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.nodeGraphEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addNodeGraphEventListener:
			s.nodeGraphEventListeners = append(s.nodeGraphEventListeners, newListener)
		case listenerToRemove := <-s.removeNodeGraphEventListener:
			for i, ch := range s.nodeGraphEventListeners {
				if ch == listenerToRemove {
					s.nodeGraphEventListeners[i] = s.nodeGraphEventListeners[len(s.nodeGraphEventListeners)-1]
					s.nodeGraphEventListeners = s.nodeGraphEventListeners[:len(s.nodeGraphEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.nodeGraphEventSource:
			if !ok {
				return
			}
			for _, listener := range s.nodeGraphEventListeners {
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

func (s *broadcastServer) serveChannelGraphEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.channelGraphEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addChannelGraphEventListener:
			s.channelGraphEventListeners = append(s.channelGraphEventListeners, newListener)
		case listenerToRemove := <-s.removeChannelGraphEventListener:
			for i, ch := range s.channelGraphEventListeners {
				if ch == listenerToRemove {
					s.channelGraphEventListeners[i] = s.channelGraphEventListeners[len(s.channelGraphEventListeners)-1]
					s.channelGraphEventListeners = s.channelGraphEventListeners[:len(s.channelGraphEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.channelGraphEventSource:
			if !ok {
				return
			}
			for _, listener := range s.channelGraphEventListeners {
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

func (s *broadcastServer) serveInvoiceEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.invoiceEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addInvoiceEventListener:
			s.invoiceEventListeners = append(s.invoiceEventListeners, newListener)
		case listenerToRemove := <-s.removeInvoiceEventListener:
			for i, ch := range s.invoiceEventListeners {
				if ch == listenerToRemove {
					s.invoiceEventListeners[i] = s.invoiceEventListeners[len(s.invoiceEventListeners)-1]
					s.invoiceEventListeners = s.invoiceEventListeners[:len(s.invoiceEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.invoiceEventSource:
			if !ok {
				return
			}
			for _, listener := range s.invoiceEventListeners {
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

func (s *broadcastServer) servePaymentEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.paymentEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addPaymentEventListener:
			s.paymentEventListeners = append(s.paymentEventListeners, newListener)
		case listenerToRemove := <-s.removePaymentEventListener:
			for i, ch := range s.paymentEventListeners {
				if ch == listenerToRemove {
					s.paymentEventListeners[i] = s.paymentEventListeners[len(s.paymentEventListeners)-1]
					s.paymentEventListeners = s.paymentEventListeners[:len(s.paymentEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.paymentEventSource:
			if !ok {
				return
			}
			for _, listener := range s.paymentEventListeners {
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

func (s *broadcastServer) serveTransactionEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.transactionEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addTransactionEventListener:
			s.transactionEventListeners = append(s.transactionEventListeners, newListener)
		case listenerToRemove := <-s.removeTransactionEventListener:
			for i, ch := range s.transactionEventListeners {
				if ch == listenerToRemove {
					s.transactionEventListeners[i] = s.transactionEventListeners[len(s.transactionEventListeners)-1]
					s.transactionEventListeners = s.transactionEventListeners[:len(s.transactionEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.transactionEventSource:
			if !ok {
				return
			}
			for _, listener := range s.transactionEventListeners {
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

func (s *broadcastServer) servePeerEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.peerEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addPeerEventListener:
			s.peerEventListeners = append(s.peerEventListeners, newListener)
		case listenerToRemove := <-s.removePeerEventListener:
			for i, ch := range s.peerEventListeners {
				if ch == listenerToRemove {
					s.peerEventListeners[i] = s.peerEventListeners[len(s.peerEventListeners)-1]
					s.peerEventListeners = s.peerEventListeners[:len(s.peerEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.peerEventSource:
			if !ok {
				return
			}
			for _, listener := range s.peerEventListeners {
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

func (s *broadcastServer) serveBlockEvent(ctx context.Context) {
	defer func() {
		for _, listener := range s.blockEventListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addBlockEventListener:
			s.blockEventListeners = append(s.blockEventListeners, newListener)
		case listenerToRemove := <-s.removeBlockEventListener:
			for i, ch := range s.blockEventListeners {
				if ch == listenerToRemove {
					s.blockEventListeners[i] = s.blockEventListeners[len(s.blockEventListeners)-1]
					s.blockEventListeners = s.blockEventListeners[:len(s.blockEventListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.blockEventSource:
			if !ok {
				return
			}
			for _, listener := range s.blockEventListeners {
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

func (s *broadcastServer) serveWebSocketResponse(ctx context.Context) {
	defer func() {
		for _, listener := range s.webSocketResponseListeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addWebSocketResponseListener:
			s.webSocketResponseListeners = append(s.webSocketResponseListeners, newListener)
		case listenerToRemove := <-s.removeWebSocketResponseListener:
			for i, ch := range s.webSocketResponseListeners {
				if ch == listenerToRemove {
					s.webSocketResponseListeners[i] = s.webSocketResponseListeners[len(s.webSocketResponseListeners)-1]
					s.webSocketResponseListeners = s.webSocketResponseListeners[:len(s.webSocketResponseListeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.webSocketResponseSource:
			if !ok {
				return
			}
			for _, listener := range s.webSocketResponseListeners {
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
