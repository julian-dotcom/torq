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
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

func ChannelBalanceCacheMaintenance(ctx context.Context, lndClient lnrpc.LightningClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, broadcaster broadcast.BroadcastServer, eventChannel chan interface{},
	serviceEventChannel chan commons.ServiceEvent) {

	serviceStatus := commons.Inactive
	bootStrapping := true
	subscriptionStream := commons.ChannelBalanceCacheStream
	lndSyncTicker := clock.New().Tick(commons.CHANNELBALANCE_TICKER_SECONDS * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-lndSyncTicker:
			bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping,
				serviceStatus, serviceEventChannel, subscriptionStream, lndClient, db)
		case event := <-broadcaster.Subscribe():
			processBroadcastedEvent(event)
		}
	}
}

func synchronizeDataFromLnd(nodeSettings commons.ManagedNodeSettings, bootStrapping bool, serviceStatus commons.Status,
	serviceEventChannel chan commons.ServiceEvent, subscriptionStream commons.SubscriptionStream,
	lndClient lnrpc.LightningClient, db *sqlx.DB) (bool, commons.Status) {

	if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
		if !bootStrapping {
			bootStrapping = true
			log.Error().Msgf("Channel balance cache got out-of-sync because of a non-active LND stream. (nodeId: %v)", nodeSettings.NodeId)
		}
	}
	if bootStrapping {
		serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Initializing, serviceStatus)
	} else {
		serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
	}
	err := initializeChannelBalanceFromLnd(lndClient, nodeSettings.NodeId, db)
	if err == nil {
		if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) == commons.Active {
			bootStrapping = false
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
		}
	} else {
		serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
		log.Error().Err(err).Msgf("Failed to initialize channel balance cache. This is a critical issue! (nodeId: %v)", nodeSettings.NodeId)
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
	defer func() {
		commons.GetChannelStateLock(nodeId).Unlock()
	}()
	commons.ClearChannelStates(nodeId)
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

		pendingIncomingHtlcCount := 0
		pendingIncomingHtlcAmount := int64(0)
		pendingOutgoingHtlcCount := 0
		pendingOutgoingHtlcAmount := int64(0)
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
					pendingIncomingHtlcCount++
					pendingIncomingHtlcAmount += htlc.Amount
				} else {
					pendingOutgoingHtlcCount++
					pendingOutgoingHtlcAmount += htlc.Amount
				}
			}
		}
		channelStateSettings.PendingIncomingHtlcCount = pendingIncomingHtlcCount
		channelStateSettings.PendingIncomingHtlcAmount = pendingIncomingHtlcAmount
		channelStateSettings.PendingOutgoingHtlcCount = pendingOutgoingHtlcCount
		channelStateSettings.PendingOutgoingHtlcAmount = pendingOutgoingHtlcAmount
		commons.SetChannelState(nodeId, channelId, channelStateSettings)
	}
	return nil
}

func processBroadcastedEvent(event interface{}) {
	if serviceEvent, ok := event.(commons.ServiceEvent); ok {
		if serviceEvent.NodeId == 0 || serviceEvent.Type != commons.LndService {
			return
		}
		if !serviceEvent.SubscriptionStream.IsChannelBalanceCache() {
			return
		}
		commons.SetChannelStateNodeStatus(serviceEvent.NodeId, serviceEvent.Status)
	} else if channelEvent, ok := event.(commons.ChannelEvent); ok {
		if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
			return
		}

		var status commons.Status
		switch channelEvent.Type {
		case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
			status = commons.Active
		case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
			status = commons.Inactive
		case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
			status = commons.Deleted
		}
		commons.SetChannelStateChannelStatus(serviceEvent.NodeId, channelEvent.ChannelId, status)
	} else if channelGraphEvent, ok := event.(commons.ChannelGraphEvent); ok {
		if channelGraphEvent.NodeId == 0 || channelGraphEvent.ChannelId == nil || *channelGraphEvent.ChannelId == 0 ||
			channelGraphEvent.AnnouncingNodeId == nil || *channelGraphEvent.AnnouncingNodeId == 0 ||
			channelGraphEvent.ConnectingNodeId == nil || *channelGraphEvent.ConnectingNodeId == 0 {
			return
		}
		local := *channelGraphEvent.AnnouncingNodeId == channelGraphEvent.NodeId
		commons.SetChannelStateRoutingPolicy(channelGraphEvent.NodeId, *channelGraphEvent.ChannelId, local,
			channelGraphEvent.Disabled, channelGraphEvent.TimeLockDelta, channelGraphEvent.MinHtlc,
			channelGraphEvent.MaxHtlcMsat, channelGraphEvent.FeeBaseMsat, channelGraphEvent.FeeRateMilliMsat)
	} else if forwardEvent, ok := event.(commons.ForwardEvent); ok {
		if forwardEvent.NodeId == 0 {
			return
		}
		if forwardEvent.IncomingChannelId != nil {
			commons.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.IncomingChannelId, true, forwardEvent.AmountInMsat)
		}
		if forwardEvent.OutgoingChannelId != nil {
			commons.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.OutgoingChannelId, false, forwardEvent.AmountOutMsat)
		}
	} else if invoiceEvent, ok := event.(commons.InvoiceEvent); ok {
		if invoiceEvent.NodeId == 0 {
			return
		}
		commons.SetChannelStateBalanceUpdateMsat(invoiceEvent.NodeId, invoiceEvent.ChannelId, true, invoiceEvent.AmountPaidMsat)
	} else if paymentEvent, ok := event.(commons.PaymentEvent); ok {
		if paymentEvent.NodeId == 0 || paymentEvent.OutgoingChannelId == nil || *paymentEvent.OutgoingChannelId == 0 {
			return
		}
		commons.SetChannelStateBalanceUpdate(paymentEvent.NodeId, *paymentEvent.OutgoingChannelId, false, paymentEvent.AmountPaid)
	} else if htlcEvent, ok := event.(commons.HtlcEvent); ok {
		if htlcEvent.NodeId == 0 {
			return
		}
		commons.SetChannelStateBalanceHtlcEvent(htlcEvent)
	} else if peerEvent, ok := event.(commons.PeerEvent); ok {
		if peerEvent.NodeId == 0 || peerEvent.EventNodeId == 0 {
			return
		}
		var status commons.Status
		switch peerEvent.Type {
		case lnrpc.PeerEvent_PEER_ONLINE:
			status = commons.Active
		case lnrpc.PeerEvent_PEER_OFFLINE:
			status = commons.Inactive
		}
		channelIds := commons.GetChannelIdsByNodeId(peerEvent.EventNodeId)
		for channelId := range channelIds {
			commons.SetChannelStateChannelStatus(peerEvent.NodeId, channelId, status)
		}
	}
}
