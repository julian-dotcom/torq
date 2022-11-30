package lnd

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/pkg/commons"
)

func ChannelBalanceCacheMaintenance(ctx context.Context, lndClient lnrpc.LightningClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}, serviceEventChannel chan commons.ServiceEvent) {

	serviceStatus := commons.Inactive
	bootStrapping := true
	subscriptionStream := commons.ChannelBalanceCacheStream
	lndSyncTicker := clock.New().Tick(commons.CHANNELBALANCE_TICKER_SECONDS * time.Second)
	bootstrapTicker := clock.New().Tick(commons.CHANNELBALANCE_BOOTSTRAP_TICKER_SECONDS * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-lndSyncTicker:
			bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping,
				serviceStatus, serviceEventChannel, subscriptionStream, lndClient, db)
		case <-bootstrapTicker:
			if bootStrapping {
				bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping, serviceStatus,
					serviceEventChannel, subscriptionStream, lndClient, db)
			}
		}
	}
}

func synchronizeDataFromLnd(nodeSettings commons.ManagedNodeSettings, bootStrapping bool, serviceStatus commons.Status,
	serviceEventChannel chan commons.ServiceEvent, subscriptionStream commons.SubscriptionStream,
	lndClient lnrpc.LightningClient, db *sqlx.DB) (bool, commons.Status) {

	if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) == commons.Active {
		if bootStrapping {
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Initializing, serviceStatus)
		}
		err := initializeChannelBalanceFromLnd(lndClient, nodeSettings.NodeId, db)
		if err == nil {
			bootStrapping = false
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
		} else {
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
			log.Error().Err(err).Msgf("Failed to initialize channel balance cache. This is a critical issue! (nodeId: %v)", nodeSettings.NodeId)
		}
	} else {
		if !bootStrapping {
			bootStrapping = true
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Inactive, serviceStatus)
			log.Error().Msgf("Channel balance cache got out-of-sync because of a non-active LND stream. (nodeId: %v)", nodeSettings.NodeId)
		}
	}
	return bootStrapping, serviceStatus
}

func initializeChannelBalanceFromLnd(lndClient lnrpc.LightningClient, nodeId int, db *sqlx.DB) error {
	mutex := commons.GetChannelStateLock(nodeId)
	if commons.RWMutexWriteLocked(mutex) {
		log.Error().Msgf("The lock initializeChannelBalanceFromLnd is already locked? This is a critical issue! (nodeId: %v)", nodeId)
		return errors.New(fmt.Sprintf("The lock initializeChannelBalanceFromLnd is already locked? This is a critical issue! (nodeId: %v)", nodeId))
	}
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	mutex.Lock()
	defer mutex.Unlock()
	r, err := lndClient.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		return errors.Wrapf(err, "Obtaining channels from LND for nodeId: %v", nodeId)
	}
	for _, lndChannel := range r.Channels {
		channelId := commons.GetChannelIdByChannelPoint(lndChannel.ChannelPoint)
		remoteNodeId := commons.GetNodeIdByPublicKey(lndChannel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network)
		if channelId == 0 {
			return errors.Wrapf(err, "Obtaining channelId from channelPoint: %v", lndChannel.ChannelPoint)
		}
		channelStateSettings := commons.ManagedChannelStateSettings{
			NodeId:        nodeId,
			RemoteNodeId:  remoteNodeId,
			ChannelId:     channelId,
			LocalBalance:  lndChannel.LocalBalance,
			RemoteBalance: lndChannel.RemoteBalance,
			// STALE INFORMATION ONLY OBTAINED VIA LND REGULAR CHECKINS SO NOT MAINTAINED
			CommitFee:             lndChannel.CommitFee,
			CommitWeight:          lndChannel.CommitWeight,
			FeePerKw:              lndChannel.FeePerKw,
			NumUpdates:            lndChannel.NumUpdates,
			ChanStatusFlags:       lndChannel.ChanStatusFlags,
			CommitmentType:        lndChannel.CommitmentType,
			Lifetime:              lndChannel.Lifetime,
			TotalSatoshisSent:     lndChannel.TotalSatoshisSent,
			TotalSatoshisReceived: lndChannel.TotalSatoshisReceived,
		}
		localRoutingPolicy, err := channels.GetLocalRoutingPolicy(channelId, nodeId, db)
		if err != nil {
			return errors.Wrapf(err, "Obtaining LocalRoutingPolicy from the database for channelId: %v", channelId)
		}
		channelStateSettings.LocalDisabled = localRoutingPolicy.Disabled
		channelStateSettings.LocalFeeBaseMsat = localRoutingPolicy.FeeBaseMsat
		channelStateSettings.LocalFeeRateMilliMsat = localRoutingPolicy.FeeRateMillMsat
		channelStateSettings.LocalMinHtlc = localRoutingPolicy.MinHtlc
		channelStateSettings.LocalMaxHtlcMsat = localRoutingPolicy.MaxHtlcMsat
		channelStateSettings.LocalTimeLockDelta = localRoutingPolicy.TimeLockDelta

		remoteRoutingPolicy, err := channels.GetRemoteRoutingPolicy(channelId, nodeId, db)
		if err != nil {
			return errors.Wrapf(err, "Obtaining RemoteRoutingPolicy from the database for channelId: %v", channelId)
		}
		channelStateSettings.RemoteDisabled = remoteRoutingPolicy.Disabled
		channelStateSettings.RemoteFeeBaseMsat = remoteRoutingPolicy.FeeBaseMsat
		channelStateSettings.RemoteFeeRateMilliMsat = remoteRoutingPolicy.FeeRateMillMsat
		channelStateSettings.RemoteMinHtlc = remoteRoutingPolicy.MinHtlc
		channelStateSettings.RemoteMaxHtlcMsat = remoteRoutingPolicy.MaxHtlcMsat
		channelStateSettings.RemoteTimeLockDelta = remoteRoutingPolicy.TimeLockDelta

		pendingDecreasingHtlcCount := 0
		pendingDecreasingHtlcAmount := int64(0)
		pendingIncreasingHtlcCount := 0
		pendingIncreasingHtlcAmount := int64(0)
		if len(lndChannel.PendingHtlcs) > 0 {
			for _, pendingHtlc := range lndChannel.PendingHtlcs {
				htlc := commons.Htlc{
					Incoming:            pendingHtlc.Incoming,
					Amount:              pendingHtlc.Amount,
					HashLock:            pendingHtlc.HashLock,
					ExpirationHeight:    pendingHtlc.ExpirationHeight,
					HtlcIndex:           pendingHtlc.HtlcIndex,
					ForwardingChannel:   pendingHtlc.ForwardingChannel,
					ForwardingHtlcIndex: pendingHtlc.ForwardingHtlcIndex,
				}
				channelStateSettings.PendingHtlcs = append(channelStateSettings.PendingHtlcs, htlc)
				if htlc.ForwardingHtlcIndex == 0 {
					pendingDecreasingHtlcCount++
					pendingDecreasingHtlcAmount += htlc.Amount
				} else {
					pendingIncreasingHtlcCount++
					pendingIncreasingHtlcAmount += htlc.Amount
				}
			}
		}
		channelStateSettings.PendingDecreasingHtlcCount = pendingDecreasingHtlcCount
		channelStateSettings.PendingDecreasingHtlcAmount = pendingDecreasingHtlcAmount
		channelStateSettings.PendingIncreasingHtlcCount = pendingIncreasingHtlcCount
		channelStateSettings.PendingIncreasingHtlcAmount = pendingIncreasingHtlcAmount
		commons.SetChannelState(nodeId, channelId, channelStateSettings)
	}
	return nil
}
