package cln

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
)

const streamChannelsTickerSeconds = 10

type client_ListChannels interface {
	ListChannels(ctx context.Context,
		in *cln.ListchannelsRequest,
		opts ...grpc.CallOption) (*cln.ListchannelsResponse, error)
}

func SubscribeAndStoreChannels(ctx context.Context,
	client client_ListChannels,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServiceChannelsService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamChannelsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessChannels(ctx, db, client, serviceType, nodeSettings, true)
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
			err = listAndProcessChannels(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessChannels(ctx context.Context, db *sqlx.DB, client client_ListChannels,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	currentChannels := cache.GetChannelSettingsByNodeId(nodeSettings.NodeId)
	var openChannelIds []int
	for _, currentChannel := range currentChannels {
		if currentChannel.Status < core.CooperativeClosed {
			openChannelIds = append(openChannelIds, currentChannel.ChannelId)
		}
	}
	processedChannelIds := make(map[int]bool, len(currentChannels))

	publicKey, err := hex.DecodeString(nodeSettings.PublicKey)
	if err != nil {
		return errors.Wrapf(err, "decoding public key for nodeId: %v", nodeSettings.NodeId)
	}
	clnChannels, err := client.ListChannels(ctx, &cln.ListchannelsRequest{
		Source: publicKey,
	})
	if err != nil {
		return errors.Wrapf(err, "listing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeChannels(db, clnChannels.Channels, nodeSettings, processedChannelIds)
	if err != nil {
		return errors.Wrapf(err, "storing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	clnChannels, err = client.ListChannels(ctx, &cln.ListchannelsRequest{
		Destination: publicKey,
	})
	if err != nil {
		return errors.Wrapf(err, "listing destination channels for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeChannels(db, clnChannels.Channels, nodeSettings, processedChannelIds)
	if err != nil {
		return errors.Wrapf(err, "storing destination channels for nodeId: %v", nodeSettings.NodeId)
	}

	for _, openChannelId := range openChannelIds {
		if !processedChannelIds[openChannelId] {
			log.Info().Msgf("Channel with channelId: %v got dropped from the list for nodeId: %v",
				openChannelId, nodeSettings.NodeId)
			channel, err := channels.GetChannel(db, openChannelId)
			if err != nil {
				return errors.Wrapf(err, "obtaining dropped channel with channelId: %v for nodeId: %v",
					openChannelId, nodeSettings.NodeId)
			}
			channel.Status = core.CooperativeClosed
			_, err = channels.AddChannelOrUpdateChannelStatus(db, nodeSettings, channel)
			if err != nil {
				return errors.Wrapf(err, "persisting dropped channel with channelId: %v for nodeId: %v",
					openChannelId, nodeSettings.NodeId)
			}

			peerNodeId := channel.FirstNodeId
			if peerNodeId == nodeSettings.NodeId {
				peerNodeId = channel.SecondNodeId
			}

			// This stops the graph from listening to node updates
			chans, err := channels.GetOpenChannelsForNodeId(db, nodeSettings.NodeId)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to verify if remote node still has open channels: %v", peerNodeId)
			}
			if len(chans) == 0 {
				peerPublicKey := cache.GetNodeSettingsByNodeId(peerNodeId).PublicKey
				cache.SetInactiveChannelPeerNode(peerNodeId, peerPublicKey, nodeSettings.Chain, nodeSettings.Network)
			}
		}
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of peers is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storeChannels(db *sqlx.DB,
	clnChannels []*cln.ListchannelsChannels,
	nodeSettings cache.NodeSettingsCache,
	processedChannelIds map[int]bool) error {

	processedShortChannelIds := make(map[string]bool)
	for _, clnChannel := range clnChannels {
		if clnChannel != nil {
			if processedShortChannelIds[clnChannel.ShortChannelId] {
				continue
			}
			sourcePublicKey := hex.EncodeToString(clnChannel.Source)
			destinationPublicKey := hex.EncodeToString(clnChannel.Destination)
			peerNodeId := cache.GetPeerNodeIdByPublicKey(sourcePublicKey, nodeSettings.Chain, nodeSettings.Network)
			peerPublicKey := sourcePublicKey
			if peerNodeId == nodeSettings.NodeId {
				peerNodeId = cache.GetPeerNodeIdByPublicKey(destinationPublicKey, nodeSettings.Chain, nodeSettings.Network)
				peerPublicKey = destinationPublicKey
			}
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
			channelId, err := processChannel(db, clnChannel, nodeSettings, peerNodeId, peerPublicKey)
			if err != nil {
				return errors.Wrapf(err, "process channel for nodeId: %v", nodeSettings.NodeId)
			}
			processedShortChannelIds[clnChannel.ShortChannelId] = true
			processedChannelIds[channelId] = true

			announcingNodeId := cache.GetPeerNodeIdByPublicKey(
				sourcePublicKey, nodeSettings.Chain, nodeSettings.Network)
			connectingNodeId := cache.GetPeerNodeIdByPublicKey(
				destinationPublicKey, nodeSettings.Chain, nodeSettings.Network)

			channelEvent := graph_events.ChannelEventFromGraph{}
			channelEvent.ChannelId = channelId
			channelEvent.NodeId = nodeSettings.NodeId
			channelEvent.AnnouncingNodeId = announcingNodeId
			channelEvent.ConnectingNodeId = connectingNodeId
			channelEvent.Outbound = announcingNodeId == nodeSettings.NodeId
			channelEvent.FeeRateMilliMsat = int64(clnChannel.FeePerMillionth)
			channelEvent.FeeBaseMsat = int64(clnChannel.BaseFeeMillisatoshi)
			channelEvent.Disabled = !clnChannel.Active
			minHtlcMsat := clnChannel.HtlcMinimumMsat
			if minHtlcMsat != nil {
				channelEvent.MinHtlcMsat = (*minHtlcMsat).Msat
			}
			maxHtlcMsat := clnChannel.HtlcMaximumMsat
			if maxHtlcMsat != nil {
				channelEvent.MaxHtlcMsat = (*maxHtlcMsat).Msat
			}
			channelEvent.TimeLockDelta = clnChannel.Delay
			err = insertRoutingPolicy(db, channelEvent, nodeSettings)
			if err != nil {
				return errors.Wrapf(err, "process routing policy for nodeId: %v", nodeSettings.NodeId)
			}
		}
	}
	return nil
}

func processChannel(db *sqlx.DB,
	clnChannel *cln.ListchannelsChannels,
	nodeSettings cache.NodeSettingsCache,
	peerNodeId int,
	peerPublicKey string) (int, error) {

	channelId := cache.GetChannelIdByShortChannelId(&clnChannel.ShortChannelId)
	var channel channels.Channel
	if channelId == 0 {
		// TODO FIXME CLN: ONCE the CLN gRPC is fixed then remove this code
		// Initial channel import should not be from the ListChannels gRPC!!!
		channel = channels.Channel{
			ShortChannelID: &clnChannel.ShortChannelId,
			FirstNodeId:    nodeSettings.NodeId,
			SecondNodeId:   peerNodeId,
			Status:         core.Open,
		}
	} else {
		channelSettings := cache.GetChannelSettingByChannelId(channelId)
		channel = channels.Channel{
			ChannelID:              channelSettings.ChannelId,
			ShortChannelID:         channelSettings.ShortChannelId,
			ClosingTransactionHash: channelSettings.ClosingTransactionHash,
			Capacity:               channelSettings.Capacity,
			Private:                channelSettings.Private,
			FirstNodeId:            channelSettings.FirstNodeId,
			SecondNodeId:           channelSettings.SecondNodeId,
			InitiatingNodeId:       channelSettings.InitiatingNodeId,
			AcceptingNodeId:        channelSettings.AcceptingNodeId,
			ClosingNodeId:          channelSettings.ClosingNodeId,
			Status:                 channelSettings.Status,
			FundingBlockHeight:     channelSettings.FundingBlockHeight,
			FundedOn:               channelSettings.FundedOn,
			ClosingBlockHeight:     channelSettings.ClosingBlockHeight,
			ClosedOn:               channelSettings.ClosedOn,
			Flags:                  channelSettings.Flags,
		}
	}
	if clnChannel.AmountMsat != nil {
		channel.Capacity = int64(clnChannel.AmountMsat.Msat / 1_000)
	}
	channel.Private = !clnChannel.Public
	channelId, err := channels.AddChannelOrUpdateChannelStatus(db, nodeSettings, channel)
	if err != nil {
		return 0, errors.Wrapf(err, "update channel data for channelId: %v, nodeId: %v",
			channelId, nodeSettings.NodeId)
	}
	cache.SetChannelPeerNode(peerNodeId, peerPublicKey, nodeSettings.Chain, nodeSettings.Network, core.Open)
	return channelId, nil
}

func insertRoutingPolicy(
	db *sqlx.DB,
	channelEvent graph_events.ChannelEventFromGraph,
	nodeSettings cache.NodeSettingsCache) error {

	existingChannelEvent := graph_events.ChannelEventFromGraph{}
	err := db.Get(&existingChannelEvent, `
				SELECT *
				FROM routing_policy
				WHERE channel_id=$1 AND announcing_node_id=$2 AND connecting_node_id=$3
				ORDER BY ts DESC
				LIMIT 1;`, channelEvent.ChannelId, channelEvent.AnnouncingNodeId, channelEvent.ConnectingNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrapf(err, "insertNodeEvent -> getPreviousChannelEvent.")
		}
	}

	// If one of our active torq nodes is announcing_node_id then the channel update was by our node
	// TODO FIXME ignore if previous update was from the same node so if announcing_node_id=node_id on previous record
	// and the current parameters are announcing_node_id!=node_id
	if existingChannelEvent.Disabled != channelEvent.Disabled ||
		existingChannelEvent.FeeBaseMsat != channelEvent.FeeBaseMsat ||
		existingChannelEvent.FeeRateMilliMsat != channelEvent.FeeRateMilliMsat ||
		existingChannelEvent.MaxHtlcMsat != channelEvent.MaxHtlcMsat ||
		existingChannelEvent.MinHtlcMsat != channelEvent.MinHtlcMsat ||
		existingChannelEvent.TimeLockDelta != channelEvent.TimeLockDelta {

		now := time.Now().UTC()
		_, err := db.Exec(`
		INSERT INTO routing_policy
			(ts,disabled,time_lock_delta,min_htlc,max_htlc_msat,fee_base_msat,fee_rate_mill_msat,
			 channel_id,announcing_node_id,connecting_node_id,node_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`, now,
			channelEvent.Disabled, channelEvent.TimeLockDelta, channelEvent.MinHtlcMsat,
			channelEvent.MaxHtlcMsat, channelEvent.FeeBaseMsat, channelEvent.FeeRateMilliMsat,
			channelEvent.ChannelId, channelEvent.AnnouncingNodeId, channelEvent.ConnectingNodeId, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "insertRoutingPolicy")
		}
	}
	return nil
}
