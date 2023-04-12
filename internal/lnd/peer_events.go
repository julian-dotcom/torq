package lnd

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
)

type peerEventsClient interface {
	SubscribePeerEvents(ctx context.Context, in *lnrpc.PeerEventSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error)
}

func SubscribePeerEvents(ctx context.Context,
	client peerEventsClient,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := core.LndServicePeerEventStream

	stream, err := client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			cache.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
			return
		}
		log.Error().Err(err).Msgf(
			"%v failure to obtain a stream from LND for nodeId: %v", serviceType.String(), nodeSettings.NodeId)
		cache.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
		return
	}

	cache.SetActiveLndServiceState(serviceType, nodeSettings.NodeId)

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
			return
		default:
		}

		peerEvent, err := stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				cache.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			log.Error().Err(err).Msgf(
				"Receiving channel events from the stream failed for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		eventNodeId := cache.GetNodeIdByPublicKey(peerEvent.PubKey, nodeSettings.Chain, nodeSettings.Network)
		if eventNodeId != 0 {
			err = setNodeConnectionHistory(db, peerEvent, eventNodeId, nodeSettings.NodeId)
			if err != nil {
				log.Error().Err(err).Msgf(
					"Adding node connection history entry failed for nodeId: %v (eventNodeId: %v)",
					nodeSettings.NodeId, eventNodeId)
			}

			ProcessPeerEvent(core.PeerEvent{
				EventData: core.EventData{
					EventTime: time.Now().UTC(),
					NodeId:    nodeSettings.NodeId,
				},
				Type:        peerEvent.Type,
				EventNodeId: eventNodeId,
			})
		}
	}
}

func setNodeConnectionHistory(db *sqlx.DB,
	peerEvent *lnrpc.PeerEvent,
	eventNodeId int,
	torqNodeId int) error {

	address, setting, connectionStatus, err := settings.GetNodeConnectionHistoryWithDetail(db, torqNodeId, eventNodeId)
	if err != nil {
		return errors.Wrap(err, "obtaining existing details like address, settings and status")
	}
	switch peerEvent.Type {
	case lnrpc.PeerEvent_PEER_ONLINE:
		if connectionStatus == nil || *connectionStatus != core.NodeConnectionStatusConnected {
			connected := core.NodeConnectionStatusConnected
			err = settings.AddNodeConnectionHistory(db, torqNodeId, eventNodeId, address, setting, &connected)
		}
	case lnrpc.PeerEvent_PEER_OFFLINE:
		if connectionStatus == nil || *connectionStatus != core.NodeConnectionStatusDisconnected {
			disconnected := core.NodeConnectionStatusDisconnected
			err = settings.AddNodeConnectionHistory(db, torqNodeId, eventNodeId, address, setting, &disconnected)
		}
	}
	if err != nil {
		return errors.Wrap(err, "adding connection history")
	}
	return nil
}
