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

func SubscribePeerEvents(ctx context.Context,
	client peerEventsClient,
	nodeSettings commons.ManagedNodeSettings) {

	serviceType := commons.LndServiceHtlcEventStream

	stream, err := client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			commons.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
			return
		}
		log.Error().Err(err).Msgf(
			"Obtaining stream (%v) from LND failed for nodeId: %v", serviceType.String(), nodeSettings.NodeId)
		commons.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
		return
	}

	commons.SetActiveLndServiceState(serviceType, nodeSettings.NodeId)

	for {
		select {
		case <-ctx.Done():
			commons.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
			return
		default:
		}

		peerEvent, err := stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				commons.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			log.Error().Err(err).Msgf(
				"Receiving channel events from the stream failed for nodeId: %v", nodeSettings.NodeId)
			commons.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		eventNodeId := commons.GetNodeIdByPublicKey(peerEvent.PubKey, nodeSettings.Chain, nodeSettings.Network)
		if eventNodeId != 0 {
			ProcessPeerEvent(commons.PeerEvent{
				EventData: commons.EventData{
					EventTime: time.Now().UTC(),
					NodeId:    nodeSettings.NodeId,
				},
				Type:        peerEvent.Type,
				EventNodeId: eventNodeId,
			})
		}
	}
}
