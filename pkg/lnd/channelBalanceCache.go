package lnd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

func ChannelBalanceCacheMaintenance(ctx context.Context, lndClient lnrpc.LightningClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer,
	serviceEventChannel chan commons.ServiceEvent) {

	defer log.Info().Msgf("ChannelBalanceCacheMaintenance terminated for nodeId: %v", nodeSettings.NodeId)

	serviceStatus := commons.Inactive
	bootStrapping := true
	subscriptionStream := commons.ChannelBalanceCacheStream
	lndSyncTicker := clock.New().Tick(commons.CHANNELBALANCE_TICKER_SECONDS * time.Second)
	mutex := &sync.RWMutex{}

	bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping,
		serviceStatus, serviceEventChannel, subscriptionStream, lndClient, db, mutex)

	go processServiceEvent(ctx, broadcaster)
	go processChannelEvent(ctx, broadcaster)
	go processChannelGraphEvent(ctx, broadcaster)
	go processForwardEvent(ctx, broadcaster)
	go processInvoiceEvent(ctx, broadcaster)
	go processPaymentEvent(ctx, broadcaster)
	//go processHtlcEvent(ctx, broadcaster)
	go processPeerEvent(ctx, broadcaster)
	go processWebSocketResponse(ctx, broadcaster)

	for {
		select {
		case <-ctx.Done():
			return
		case <-lndSyncTicker:
			bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping,
				serviceStatus, serviceEventChannel, subscriptionStream, lndClient, db, mutex)
		}
	}
}

func synchronizeDataFromLnd(nodeSettings commons.ManagedNodeSettings, bootStrapping bool, serviceStatus commons.Status,
	serviceEventChannel chan commons.ServiceEvent, subscriptionStream commons.SubscriptionStream,
	lndClient lnrpc.LightningClient, db *sqlx.DB, mutex *sync.RWMutex) (bool, commons.Status) {

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
	err := initializeChannelBalanceFromLnd(lndClient, nodeSettings.NodeId, db, mutex)
	if err != nil {
		serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
		log.Error().Err(err).Msgf("Failed to initialize channel balance cache. This is a critical issue! (nodeId: %v)", nodeSettings.NodeId)
		return bootStrapping, serviceStatus
	}
	if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) == commons.Active {
		bootStrapping = false
		serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
	}
	return bootStrapping, serviceStatus
}

func initializeChannelBalanceFromLnd(lndClient lnrpc.LightningClient, nodeId int, db *sqlx.DB, mutex *sync.RWMutex) error {
	if commons.RWMutexWriteLocked(mutex) {
		log.Error().Msgf("The lock initializeChannelBalanceFromLnd is already locked? This is a critical issue! (nodeId: %v)", nodeId)
		return errors.New(fmt.Sprintf("The lock initializeChannelBalanceFromLnd is already locked? This is a critical issue! (nodeId: %v)", nodeId))
	}
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()
	var channelStateSettingsList []commons.ManagedChannelStateSettings
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
		//localRoutingPolicy, err := channels.GetLocalRoutingPolicy(channelId, nodeId, db)
		//if err != nil {
		//	return errors.Wrapf(err, "Obtaining LocalRoutingPolicy from the database for channelId: %v", channelId)
		//}
		//channelStateSettings.LocalDisabled = localRoutingPolicy.Disabled
		//channelStateSettings.LocalFeeBaseMsat = localRoutingPolicy.FeeBaseMsat
		//channelStateSettings.LocalFeeRateMilliMsat = localRoutingPolicy.FeeRateMillMsat
		//channelStateSettings.LocalMinHtlcMsat = localRoutingPolicy.MinHtlcMsat
		//channelStateSettings.LocalMaxHtlcMsat = localRoutingPolicy.MaxHtlcMsat
		//channelStateSettings.LocalTimeLockDelta = localRoutingPolicy.TimeLockDelta
		//
		//remoteRoutingPolicy, err := channels.GetRemoteRoutingPolicy(channelId, nodeId, db)
		//if err != nil {
		//	return errors.Wrapf(err, "Obtaining RemoteRoutingPolicy from the database for channelId: %v", channelId)
		//}
		//channelStateSettings.RemoteDisabled = remoteRoutingPolicy.Disabled
		//channelStateSettings.RemoteFeeBaseMsat = remoteRoutingPolicy.FeeBaseMsat
		//channelStateSettings.RemoteFeeRateMilliMsat = remoteRoutingPolicy.FeeRateMillMsat
		//channelStateSettings.RemoteMinHtlcMsat = remoteRoutingPolicy.MinHtlcMsat
		//channelStateSettings.RemoteMaxHtlcMsat = remoteRoutingPolicy.MaxHtlcMsat
		//channelStateSettings.RemoteTimeLockDelta = remoteRoutingPolicy.TimeLockDelta

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
		channelStateSettingsList = append(channelStateSettingsList, channelStateSettings)
	}
	commons.SetChannelStates(nodeId, channelStateSettingsList)
	commons.SetChannelStateNodeStatus(nodeId, commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeId))
	return nil
}

func processServiceEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeServiceEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionServiceEvent(listener)
			return
		case serviceEvent := <-listener:
			if serviceEvent.NodeId == 0 || serviceEvent.Type != commons.LndService {
				continue
			}
			if serviceEvent.SubscriptionStream == nil {
				continue
			}
			if !serviceEvent.SubscriptionStream.IsChannelBalanceCache() {
				continue
			}
			commons.SetChannelStateNodeStatus(serviceEvent.NodeId, commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(serviceEvent.NodeId))
		}
	}
}

func processChannelEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeChannelEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionChannelEvent(listener)
			return
		case channelEvent := <-listener:
			if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
				continue
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
			commons.SetChannelStateChannelStatus(channelEvent.NodeId, channelEvent.ChannelId, status)
		}
	}
}

func processChannelGraphEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeChannelGraphEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionChannelGraphEvent(listener)
			return
		case channelGraphEvent := <-listener:
			if channelGraphEvent.NodeId == 0 || channelGraphEvent.ChannelId == nil || *channelGraphEvent.ChannelId == 0 ||
				channelGraphEvent.AnnouncingNodeId == nil || *channelGraphEvent.AnnouncingNodeId == 0 ||
				channelGraphEvent.ConnectingNodeId == nil || *channelGraphEvent.ConnectingNodeId == 0 {
				continue
			}
			local := *channelGraphEvent.AnnouncingNodeId == channelGraphEvent.NodeId
			commons.SetChannelStateRoutingPolicy(channelGraphEvent.NodeId, *channelGraphEvent.ChannelId, local,
				channelGraphEvent.Disabled, channelGraphEvent.TimeLockDelta, channelGraphEvent.MinHtlcMsat,
				channelGraphEvent.MaxHtlcMsat, channelGraphEvent.FeeBaseMsat, channelGraphEvent.FeeRateMilliMsat)
		}
	}
}

func processForwardEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeForwardEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionForwardEvent(listener)
			return
		case forwardEvent := <-listener:
			if forwardEvent.NodeId == 0 {
				continue
			}
			if forwardEvent.IncomingChannelId != nil {
				commons.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.IncomingChannelId, true, forwardEvent.AmountInMsat)
			}
			if forwardEvent.OutgoingChannelId != nil {
				commons.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.OutgoingChannelId, false, forwardEvent.AmountOutMsat)
			}
		}
	}
}

func processInvoiceEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeInvoiceEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionInvoiceEvent(listener)
			return
		case invoiceEvent := <-listener:
			if invoiceEvent.NodeId == 0 || invoiceEvent.State != lnrpc.Invoice_SETTLED {
				continue
			}
			commons.SetChannelStateBalanceUpdateMsat(invoiceEvent.NodeId, invoiceEvent.ChannelId, true, invoiceEvent.AmountPaidMsat)
		}
	}
}

func processPaymentEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribePaymentEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionPaymentEvent(listener)
			return
		case paymentEvent := <-listener:
			if paymentEvent.NodeId == 0 || paymentEvent.OutgoingChannelId == nil || *paymentEvent.OutgoingChannelId == 0 || paymentEvent.PaymentStatus != lnrpc.Payment_SUCCEEDED {
				continue
			}
			commons.SetChannelStateBalanceUpdate(paymentEvent.NodeId, *paymentEvent.OutgoingChannelId, false, paymentEvent.AmountPaid)
		}
	}
}

//func processHtlcEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
//	listener := broadcaster.SubscribeHtlcEvent()
//	for {
//		select {
//		case <-ctx.Done():
//			broadcaster.CancelSubscriptionHtlcEvent(listener)
//			return
//		case htlcEvent := <- listener:
//		    if htlcEvent.NodeId == 0 {
//			    continue
//		    }
//		    commons.SetChannelStateBalanceHtlcEvent(htlcEvent)
//	    }
//	}
//}

func processPeerEvent(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribePeerEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionPeerEvent(listener)
			return
		case peerEvent := <-listener:
			if peerEvent.NodeId == 0 || peerEvent.EventNodeId == 0 {
				continue
			}
			var status commons.Status
			switch peerEvent.Type {
			case lnrpc.PeerEvent_PEER_ONLINE:
				status = commons.Active
			case lnrpc.PeerEvent_PEER_OFFLINE:
				status = commons.Inactive
			}
			channelIds := commons.GetChannelIdsByNodeId(peerEvent.EventNodeId)
			for _, channelId := range channelIds {
				commons.SetChannelStateChannelStatus(peerEvent.NodeId, channelId, status)
			}
			// Channel open requires confirmations
			//} else if openChannelEvent, ok := event.(commons.OpenChannelResponse); ok {
			//	if openChannelEvent.Request.NodeId == 0 || openChannelEvent.Status != commons.Open {
			//		continue
			//	}
			//	commons.SetChannelStateChannelStatus(openChannelEvent.Request.NodeId, openChannelEvent.ChannelId, commons.Inactive)
		}
	}
}

func processWebSocketResponse(ctx context.Context, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeWebSocketResponse()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionWebSocketResponse(listener)
			return
		case event := <-listener:
			if closeChannelEvent, ok := event.(commons.CloseChannelResponse); ok {
				if closeChannelEvent.Request.NodeId == 0 {
					continue
				}
				commons.SetChannelStateChannelStatus(closeChannelEvent.Request.NodeId, closeChannelEvent.Request.ChannelId, commons.Deleted)
			} else if updateChannelEvent, ok := event.(commons.RoutingPolicyUpdateResponse); ok {
				if updateChannelEvent.Request.NodeId == 0 || updateChannelEvent.Request.ChannelId == 0 {
					continue
				}
				// Force Response because we don't care about balance accuracy
				currentStates := commons.GetChannelState(updateChannelEvent.Request.NodeId, updateChannelEvent.Request.ChannelId, true)
				timeLockDelta := currentStates.LocalTimeLockDelta
				if updateChannelEvent.Request.TimeLockDelta != nil {
					timeLockDelta = *updateChannelEvent.Request.TimeLockDelta
				}
				minHtlcMsat := currentStates.LocalMinHtlcMsat
				if updateChannelEvent.Request.MinHtlcMsat != nil {
					minHtlcMsat = *updateChannelEvent.Request.MinHtlcMsat
				}
				maxHtlcMsat := currentStates.LocalMaxHtlcMsat
				if updateChannelEvent.Request.MaxHtlcMsat != nil {
					maxHtlcMsat = *updateChannelEvent.Request.MaxHtlcMsat
				}
				feeBaseMsat := currentStates.LocalFeeBaseMsat
				if updateChannelEvent.Request.FeeBaseMsat != nil {
					feeBaseMsat = *updateChannelEvent.Request.FeeBaseMsat
				}
				feeRateMilliMsat := currentStates.LocalFeeRateMilliMsat
				if updateChannelEvent.Request.FeeRateMilliMsat != nil {
					feeRateMilliMsat = *updateChannelEvent.Request.FeeRateMilliMsat
				}
				commons.SetChannelStateRoutingPolicy(updateChannelEvent.Request.NodeId, updateChannelEvent.Request.ChannelId, true,
					currentStates.LocalDisabled, timeLockDelta, minHtlcMsat, maxHtlcMsat, feeBaseMsat, feeRateMilliMsat)
			}
		}
	}
}
