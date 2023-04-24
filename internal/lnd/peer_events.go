package lnd

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/settings"
)

type peerEventsClient interface {
	SubscribePeerEvents(ctx context.Context, in *lnrpc.PeerEventSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribePeerEventsClient, error)
}

type listPeersClient interface {
	ListPeers(ctx context.Context, in *lnrpc.ListPeersRequest,
		opts ...grpc.CallOption) (*lnrpc.ListPeersResponse, error)
}

func SubscribePeerEvents(ctx context.Context,
	client peerEventsClient,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.LndServicePeerEventStream

	stream, err := client.SubscribePeerEvents(ctx, &lnrpc.PeerEventSubscription{})
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}
		log.Error().Err(err).Msgf(
			"%v failure to obtain a stream from LND for nodeId: %v", serviceType.String(), nodeSettings.NodeId)
		cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
		return
	}

	cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		default:
		}

		peerEvent, err := stream.Recv()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			log.Error().Err(err).Msgf(
				"Receiving channel events from the stream failed for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		eventNodeId := cache.GetPeerNodeIdByPublicKey(peerEvent.PubKey, nodeSettings.Chain, nodeSettings.Network)
		if eventNodeId != 0 {
			err = setNodeConnectionHistory(db, peerEvent.Type, eventNodeId, nodeSettings.NodeId)
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
	peerEventType lnrpc.PeerEvent_EventType,
	eventNodeId int,
	torqNodeId int) error {

	address, setting, connectionStatus, err := settings.GetNodeConnectionHistoryWithDetail(db, torqNodeId, eventNodeId)
	if err != nil {
		return errors.Wrap(err, "obtaining existing details like address, settings and status")
	}
	switch peerEventType {
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

func ImportPeerStatusFromLnd(ctx context.Context,
	client listPeersClient,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) error {

	resp, err := client.ListPeers(ctx, &lnrpc.ListPeersRequest{LatestError: true})
	if err != nil {
		return errors.Wrapf(err, "get list of peers from lnd for nodeId: %v", nodeSettings.NodeId)
	}

	processedPeerNodeIds := make(map[int]bool)
	for _, peer := range resp.Peers {
		peerNodeId := cache.GetPeerNodeIdByPublicKey(peer.PubKey, nodeSettings.Chain, nodeSettings.Network)
		if peerNodeId == 0 {
			peerNodeId, err = nodes.AddNodeWhenNew(db, nodes.Node{
				PublicKey: peer.PubKey,
				Chain:     nodeSettings.Chain,
				Network:   nodeSettings.Network,
			}, nil)
			if err != nil {
				return errors.Wrapf(err, "adding unknown peer (%v) for nodeId: %v", peer.PubKey, nodeSettings.NodeId)
			}
		}
		processedPeerNodeIds[peerNodeId] = true
		err = setNodeConnectionHistory(db, lnrpc.PeerEvent_PEER_ONLINE, peerNodeId, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "insert online peer status for nodeId: %v", nodeSettings.NodeId)
		}
	}

	var nodeIds []int
	err = db.Select(&nodeIds, `
			SELECT n.node_id
			FROM Node n
			LEFT JOIN (
				SELECT LAST(node_id, created_on) as node_id,
					   LAST(torq_node_id, created_on) as torq_node_id,
		       		   LAST(connection_status, created_on) as connection_status
				FROM node_connection_history
				GROUP BY node_id
			) nch on nch.node_id = n.node_id
			JOIN node_connection_details as ncd ON ncd.node_id = nch.torq_node_id
			WHERE nch.torq_node_id IS NOT NULL
				AND ncd.status_id NOT IN ($1, $2)
				AND n.network = $3
				AND nch.connection_status = $4;`,
		core.Deleted, core.Archived, nodeSettings.Network, core.NodeConnectionStatusConnected)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			nodeIds = []int{}
		}
		return errors.Wrapf(err, "obtaining existing peer status for nodeId: %v", nodeSettings.NodeId)
	}

	for _, peerNodeId := range nodeIds {
		if processedPeerNodeIds[peerNodeId] {
			continue
		}
		err = setNodeConnectionHistory(db, lnrpc.PeerEvent_PEER_OFFLINE, peerNodeId, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "insert offline peer status for nodeId: %v", nodeSettings.NodeId)
		}
	}

	return nil
}
