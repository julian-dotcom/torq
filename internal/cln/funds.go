package cln

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/services_core"
	"github.com/lncapital/torq/proto/cln"
)

const streamFundsTickerSeconds = 10

type client_ListFunds interface {
	ListFunds(ctx context.Context,
		in *cln.ListfundsRequest,
		opts ...grpc.CallOption) (*cln.ListfundsResponse, error)
}

func SubscribeAndStoreFunds(ctx context.Context,
	client client_ListFunds,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_core.ClnServiceFundsService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamFundsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessFunds(ctx, db, client, serviceType, nodeSettings, true)
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
			err = listAndProcessFunds(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessFunds(ctx context.Context, db *sqlx.DB, client client_ListFunds,
	serviceType services_core.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	clnFunds, err := client.ListFunds(ctx, &cln.ListfundsRequest{})
	if err != nil {
		return errors.Wrapf(err, "listing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeChannelFunds(db, clnFunds.Channels, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of peers is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storeChannelFunds(db *sqlx.DB,
	clnChannelFunds []*cln.ListfundsChannels,
	nodeSettings cache.NodeSettingsCache) error {

	var channelStateSettingsList []cache.ChannelStateSettingsCache
	for _, clnChannel := range clnChannelFunds {
		channelId := 0
		if clnChannel.ShortChannelId != nil {
			channelId = cache.GetChannelIdByShortChannelId(clnChannel.ShortChannelId)
		}
		if channelId == 0 && len(clnChannel.FundingTxid) != 0 {
			fti := hex.EncodeToString(clnChannel.FundingTxid)
			foi := int(clnChannel.FundingOutput)
			channelId = cache.GetChannelIdByFundingTransaction(&fti, &foi)
		}
		if channelId == 0 {
			log.Info().Msgf("received funds for unknown channel for nodeId: %v", nodeSettings.NodeId)
			continue
		}
		if clnChannel.OurAmountMsat == nil || clnChannel.AmountMsat == nil {
			continue
		}
		remoteNodeId := cache.GetChannelPeerNodeIdByPublicKey(hex.EncodeToString(clnChannel.PeerId), nodeSettings.Chain, nodeSettings.Network)
		if channelId == 0 {
			return errors.New(fmt.Sprintf("obtaining remoteNodeId from peer public key: %v", hex.EncodeToString(clnChannel.PeerId)))
		}
		channelStateSettings := cache.ChannelStateSettingsCache{
			NodeId:        nodeSettings.NodeId,
			RemoteNodeId:  remoteNodeId,
			ChannelId:     channelId,
			LocalBalance:  int64(clnChannel.OurAmountMsat.Msat / 1000),
			RemoteBalance: int64(clnChannel.AmountMsat.Msat/1000 - clnChannel.OurAmountMsat.Msat/1000),
		}
		localRoutingPolicy, err := channels.GetLocalRoutingPolicy(db, channelId, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "obtaining LocalRoutingPolicy from the database for channelId: %v", channelId)
		}
		channelStateSettings.LocalDisabled = localRoutingPolicy.Disabled
		channelStateSettings.LocalFeeBaseMsat = localRoutingPolicy.FeeBaseMsat
		channelStateSettings.LocalFeeRateMilliMsat = localRoutingPolicy.FeeRateMillMsat
		channelStateSettings.LocalMinHtlcMsat = localRoutingPolicy.MinHtlcMsat
		channelStateSettings.LocalMaxHtlcMsat = localRoutingPolicy.MaxHtlcMsat
		channelStateSettings.LocalTimeLockDelta = localRoutingPolicy.TimeLockDelta

		remoteRoutingPolicy, err := channels.GetRemoteRoutingPolicy(db, channelId, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "obtaining RemoteRoutingPolicy from the database for channelId: %v", channelId)
		}
		channelStateSettings.RemoteDisabled = remoteRoutingPolicy.Disabled
		channelStateSettings.RemoteFeeBaseMsat = remoteRoutingPolicy.FeeBaseMsat
		channelStateSettings.RemoteFeeRateMilliMsat = remoteRoutingPolicy.FeeRateMillMsat
		channelStateSettings.RemoteMinHtlcMsat = remoteRoutingPolicy.MinHtlcMsat
		channelStateSettings.RemoteMaxHtlcMsat = remoteRoutingPolicy.MaxHtlcMsat
		channelStateSettings.RemoteTimeLockDelta = remoteRoutingPolicy.TimeLockDelta

		channelStateSettingsList = append(channelStateSettingsList, channelStateSettings)
	}
	cache.SetChannelStates(nodeSettings.NodeId, channelStateSettingsList)
	return nil
}
