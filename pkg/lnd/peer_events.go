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
	nodeSettings commons.ManagedNodeSettings, peerEventChannel chan<- commons.PeerEvent) {

	defer log.Info().Msgf("SubscribePeerEvents terminated for nodeId: %v", nodeSettings.NodeId)

	var stream lnrpc.Lightning_SubscribePeerEventsClient
	var err error
	var peerEvent *lnrpc.PeerEvent
	serviceStatus := commons.ServiceInactive
	subscriptionStream := commons.PeerEventStream

	importPeerEvents := commons.RunningServices[commons.LndService].HasCustomSetting(nodeSettings.NodeId, commons.ImportPeerEvents)
	if !importPeerEvents {
		log.Info().Msgf("Import of peer events is disabled for nodeId: %v", nodeSettings.NodeId)
		SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServiceDeleted)
		return
	}

	var delay bool

	for {
		if delay {
			ticker := clock.New().Tick(streamErrorSleepSeconds * time.Second)
			select {
			case <-ctx.Done():
				return
			case <-ticker:
			}
		} else {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		if stream == nil {
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServicePending)
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
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServiceActive)
		}

		peerEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServicePending)
			log.Error().Err(err).Msgf("Receiving peer events from the stream failed, will retry in %v seconds", streamErrorSleepSeconds)
			stream = nil
			delay = true
			continue
		}

		if peerEventChannel != nil {
			eventNodeId := commons.GetNodeIdByPublicKey(peerEvent.PubKey, nodeSettings.Chain, nodeSettings.Network)
			if eventNodeId != 0 {
				peerEventChannel <- commons.PeerEvent{
					EventData: commons.EventData{
						EventTime: time.Now().UTC(),
						NodeId:    nodeSettings.NodeId,
					},
					Type:        peerEvent.Type,
					EventNodeId: eventNodeId,
				}
			}
		}
		delay = false
	}
}
