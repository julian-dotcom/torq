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
	subscriptionStream := commons.PeerEventStream

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if stream == nil {
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
			stream, err = client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msgf("Obtaining stream (SubscribePeerEvents) from LND failed, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
				stream = nil
				time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
				continue
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
		}

		peerEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
			log.Error().Err(err).Msgf("Receiving peer events from the stream failed, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
			stream = nil
			time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
			continue
		}

		if eventChannel != nil {
			eventNodeId := commons.GetNodeIdByPublicKey(peerEvent.PubKey, nodeSettings.Chain, nodeSettings.Network)
			if eventNodeId != 0 {
				eventChannel <- commons.PeerEvent{
					EventData: commons.EventData{
						EventTime: time.Now().UTC(),
						NodeId:    nodeSettings.NodeId,
					},
					Type:        peerEvent.Type,
					EventNodeId: eventNodeId,
				}
			}
		}
	}
}
