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
	nodeSettings commons.ManagedNodeSettings, peerEventChannel chan<- commons.PeerEvent,
	serviceEventChannel chan<- commons.ServiceEvent) {

	defer log.Info().Msgf("SubscribePeerEvents terminated for nodeId: %v", nodeSettings.NodeId)

	var stream lnrpc.Lightning_SubscribePeerEventsClient
	var err error
	var peerEvent *lnrpc.PeerEvent
	serviceStatus := commons.ServiceInactive
	subscriptionStream := commons.PeerEventStream

	importPeerEvents := commons.RunningServices[commons.LndService].HasCustomSetting(nodeSettings.NodeId, commons.ImportPeerEvents)
	if !importPeerEvents {
		log.Info().Msgf("Import of peer events is disabled for nodeId: %v", nodeSettings.NodeId)
		SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServiceDeleted, serviceStatus)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if stream == nil {
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServicePending, serviceStatus)
			stream, err = client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				log.Error().Err(err).Msgf("Obtaining stream (SubscribePeerEvents) from LND failed, will retry in %v seconds", streamErrorSleepSeconds)
				stream = nil
				time.Sleep(streamErrorSleepSeconds * time.Second)
				continue
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServiceActive, serviceStatus)
		}

		peerEvent, err = stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.ServicePending, serviceStatus)
			log.Error().Err(err).Msgf("Receiving peer events from the stream failed, will retry in %v seconds", streamErrorSleepSeconds)
			stream = nil
			time.Sleep(streamErrorSleepSeconds * time.Second)
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
	}
}
