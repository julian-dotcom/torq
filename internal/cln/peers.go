package cln

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/proto/cln"
)

const streamPeersTickerSeconds = 60

type client_ListPeers interface {
	ListPeers(ctx context.Context, in *cln.ListpeersRequest, opts ...grpc.CallOption) (*cln.ListpeersResponse, error)
}

func SubscribeAndStorePeers(ctx context.Context,
	client client_ListPeers,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServicePeersService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamPeersTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerPeers := ticker.C

	err := listAndProcessPeers(ctx, db, client, serviceType, nodeSettings, true)
	if err != nil {
		processError(ctx, serviceType, nodeSettings, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-tickerPeers:
			err = listAndProcessPeers(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func processError(ctx context.Context, serviceType services_helpers.ServiceType, nodeSettings cache.NodeSettingsCache, err error) {
	if errors.Is(ctx.Err(), context.Canceled) {
		cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
		return
	}
	cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
	log.Error().Err(err).Msgf("%v failed to process for nodeId: %v", serviceType.String(), nodeSettings.NodeId)
}

func listAndProcessPeers(ctx context.Context, db *sqlx.DB, client client_ListPeers,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	peers, err := client.ListPeers(ctx, &cln.ListpeersRequest{})
	if err != nil {
		return errors.Wrapf(err, "listing peers for nodeId: %v", nodeSettings.NodeId)
	}

	err = storePeers(db, peers.Peers, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing peers for nodeId: %v", nodeSettings.NodeId)
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of peers is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storePeers(db *sqlx.DB, peers []*cln.ListpeersPeers, nodeSettings cache.NodeSettingsCache) error {
	for _, peer := range peers {
		peerPublicKey := hex.EncodeToString(peer.Id)
		peerNodeId := cache.GetPeerNodeIdByPublicKey(peerPublicKey, nodeSettings.Chain, nodeSettings.Network)
		if peerNodeId == 0 {
			var err error
			peerNodeId, err = nodes.AddNodeWhenNew(db, nodes.Node{
				PublicKey: peerPublicKey,
				Chain:     nodeSettings.Chain,
				Network:   nodeSettings.Network,
			}, nil)
			if err != nil {
				return errors.Wrapf(err, "add new peer node for nodeId: %v", nodeSettings.NodeId)
			}
		}
		address, setting, connectionStatus, err := settings.GetNodeConnectionHistoryWithDetail(db, nodeSettings.NodeId, peerNodeId)
		if err != nil {
			return errors.Wrapf(err, "obtaining node connection history for peerNodeId: %v, nodeId: %v",
				peerNodeId, nodeSettings.NodeId)
		}
		if address == nil && peer.RemoteAddr != nil {
			remoteAddress := *peer.RemoteAddr
			address = &remoteAddress
		}
		//if address == nil && peer.Netaddr != nil {
		//}
		if peer.Connected {
			cache.SetConnectedPeerNode(peerNodeId, peerPublicKey, nodeSettings.Chain, nodeSettings.Network)
			if connectionStatus == nil || *connectionStatus != core.NodeConnectionStatusConnected {
				connected := core.NodeConnectionStatusConnected
				err = settings.AddNodeConnectionHistory(db, nodeSettings.NodeId, peerNodeId, address, setting, &connected)
				if err != nil {
					return errors.Wrapf(err, "add new node connection history for nodeId: %v", nodeSettings.NodeId)
				}
			}
		}
		if !peer.Connected {
			cache.RemoveConnectedPeerNode(peerNodeId, peerPublicKey, nodeSettings.Chain, nodeSettings.Network)
			if connectionStatus == nil || *connectionStatus != core.NodeConnectionStatusDisconnected {
				disconnected := core.NodeConnectionStatusDisconnected
				err = settings.AddNodeConnectionHistory(db, nodeSettings.NodeId, peerNodeId, address, setting, &disconnected)
				if err != nil {
					return errors.Wrapf(err, "add new node disconnection history for nodeId: %v", nodeSettings.NodeId)
				}
			}
		}
	}
	return nil
}
