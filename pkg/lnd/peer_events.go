package lnd

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/commons"
)

type peerEventsClient interface {
	SubscribePeerEvents(ctx context.Context, in *lnrpc.PeerEventSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error)
}

func SubscribePeerEvents(ctx context.Context, client peerEventsClient,
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}, serviceEventChannel chan commons.ServiceEvent) {

	var stream lnrpc.Lightning_SubscribePeerEventsClient
	var err error
	var peerEvent *lnrpc.PeerEvent
	serviceStatus := commons.Inactive

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if stream == nil {
			serviceStatus = sendStreamEvent(serviceEventChannel, nodeSettings.NodeId, commons.PeerEventStream, commons.Pending, serviceStatus)
			stream, err = client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msg("Obtaining stream (SubscribePeerEvents) from LND failed, will retry in 1 minute")
				stream = nil
				time.Sleep(1 * time.Minute)
				continue
			}
			serviceStatus = sendStreamEvent(serviceEventChannel, nodeSettings.NodeId, commons.PeerEventStream, commons.Active, serviceStatus)
		}

		peerEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = sendStreamEvent(serviceEventChannel, nodeSettings.NodeId, commons.PeerEventStream, commons.Pending, serviceStatus)
			log.Error().Err(err).Msg("Receiving peer events from the stream failed, will retry in 1 minute")
			stream = nil
			time.Sleep(1 * time.Minute)
			continue
		}

		if eventChannel != nil {
			eventChannel <- commons.PeerEvent{
				EventData: commons.EventData{
					EventTime: time.Now().UTC(),
					NodeId:    nodeSettings.NodeId,
				},
				Type:           peerEvent.Type,
				EventPublicKey: peerEvent.PubKey,
			}
		}
	}
}
