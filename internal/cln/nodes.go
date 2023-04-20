package cln

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/internal/services_core"
	"github.com/lncapital/torq/proto/cln"
)

const streamNodesTickerSeconds = 15 * 60

type client_ListNodes interface {
	ListNodes(ctx context.Context,
		in *cln.ListnodesRequest,
		opts ...grpc.CallOption) (*cln.ListnodesResponse, error)
}

func SubscribeAndStoreNodes(ctx context.Context,
	client client_ListNodes,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_core.ClnServiceNodesService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamNodesTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessNodes(ctx, db, client, serviceType, nodeSettings, true)
	if err != nil {
		processError(ctx, serviceType, nodeSettings, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-tickerChannel:
			err = listAndProcessNodes(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessNodes(ctx context.Context, db *sqlx.DB, client client_ListNodes,
	serviceType services_core.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	for _, peerNodeId := range cache.GetPeerNodeIds(core.Bitcoin, nodeSettings.Network) {
		peerNodeSettings := cache.GetNodeSettingsByNodeId(peerNodeId)
		peerNodePk, err := hex.DecodeString(peerNodeSettings.PublicKey)
		if err != nil {
			return errors.Wrapf(err, "decoding peer public key for nodeId: %v", nodeSettings.NodeId)
		}
		clnNodes, err := client.ListNodes(ctx, &cln.ListnodesRequest{
			Id: peerNodePk,
		})
		if err != nil {
			return errors.Wrapf(err, "listing nodes for nodeId: %v", nodeSettings.NodeId)
		}

		err = storeNodes(db, clnNodes.Nodes, peerNodeId, nodeSettings)
		if err != nil {
			return errors.Wrapf(err, "storing source channels for nodeId: %v", nodeSettings.NodeId)
		}

		if bootStrapping {
			log.Info().Msgf("Initial import of nodes is done for nodeId: %v", nodeSettings.NodeId)
			cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
		}
	}
	return nil
}

func storeNodes(db *sqlx.DB,
	clnNodes []*cln.ListnodesNodes,
	eventNodeId int,
	nodeSettings cache.NodeSettingsCache) error {

	for _, clnNode := range clnNodes {
		eventTime := time.Now().UTC()
		if clnNode.LastTimestamp != nil {
			eventTime = time.Unix(int64(*clnNode.LastTimestamp), 0)
		}
		color := hex.EncodeToString(clnNode.Color)
		alias := ""
		if clnNode.Alias != nil {
			alias = *clnNode.Alias
		}
		// Create json byte object from node address map
		najb, err := json.Marshal(clnNode.Addresses)
		if err != nil {
			return errors.Wrap(err, "JSON Marshall node address map")
		}

		// Create json byte object from features list
		fjb, err := json.Marshal(clnNode.Features)
		if err != nil {
			return errors.Wrap(err, "JSON Marshal feature list")
		}

		nodeEvent := graph_events.NodeEventFromGraph{}
		err = db.Get(&nodeEvent, `
				SELECT *
				FROM node_event
				WHERE event_node_id=$1
				ORDER BY timestamp DESC
				LIMIT 1;`, eventNodeId)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return errors.Wrapf(err, "insertNodeEvent -> getPreviousNodeEvent.")
			}
		}

		// TODO FIXME ignore if previous update was from the same node so if event_node_id=node_id on previous record
		// and the current parameters are event_node_id!=node_id
		// TODO FIXME nodeAddresses or features can change order and still be identical
		if alias == nodeEvent.Alias &&
			color == nodeEvent.Color &&
			string(najb) == nodeEvent.NodeAddresses &&
			string(fjb) == nodeEvent.Features {

			return nil
		}

		_, err = db.Exec(`INSERT INTO node_event
    		(timestamp, event_node_id, alias, color, node_addresses, features, node_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7);`,
			eventTime, eventNodeId, alias, color, najb, fjb, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrap(err, "Executing SQL")
		}
		cache.SetNodeAlias(eventNodeId, alias)
	}
	return nil
}
