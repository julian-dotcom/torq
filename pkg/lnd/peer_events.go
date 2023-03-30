package lnd

import (
	"context"
	"time"

	"github.com/andres-erbsen/clock"
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
	nodeSettings commons.ManagedNodeSettings) {

	defer log.Info().Msgf("SubscribePeerEvents terminated for nodeId: %v", nodeSettings.NodeId)

	var stream lnrpc.Lightning_SubscribePeerEventsClient
	var err error
	var peerEvent *lnrpc.PeerEvent
	serviceStatus := commons.ServiceInactive
	serviceType := commons.LndServicePeerEventStream

	var delay bool

	for {
		if delay {
			ticker := clock.New().Tick(streamErrorSleepSeconds * time.Second)
			select {
			case <-ctx.Done():
				return
			case <-ticker:
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		if stream == nil {
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, serviceType, serviceStatus, commons.ServicePending)
			stream, err = client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msgf("Obtaining stream (SubscribePeerEvents) from LND failed, will retry in %v seconds", streamErrorSleepSeconds)
				stream = nil
				delay = true
				continue
			}
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, serviceType, serviceStatus, commons.ServiceActive)
		}

		peerEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, serviceType, serviceStatus, commons.ServicePending)
			log.Error().Err(err).Msgf("Receiving peer events from the stream failed, will retry in %v seconds", streamErrorSleepSeconds)
			stream = nil
			delay = true
			continue
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
		delay = false
	}
}
