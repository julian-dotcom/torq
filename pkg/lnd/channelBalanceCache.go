package lnd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

const channelbalanceTickerSeconds = 150

func ChannelBalanceCacheMaintenance(ctx context.Context, lndClient lnrpc.LightningClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer) {

	defer log.Info().Msgf("ChannelBalanceCacheMaintenance terminated for nodeId: %v", nodeSettings.NodeId)

	serviceStatus := commons.ServiceInactive
	bootStrapping := true
	subscriptionStream := commons.ChannelBalanceCacheStream
	lndSyncTicker := clock.New().Tick(channelbalanceTickerSeconds * time.Second)
	mutex := &sync.RWMutex{}
	initiateSync := make(chan bool)

	bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping,
		serviceStatus, subscriptionStream, lndClient, db, mutex)

	go processServiceEvent(ctx, initiateSync, broadcaster, broadcaster.SubscribeServiceEvent())
	go processChannelEvent(ctx, broadcaster, broadcaster.SubscribeChannelEvent())
	go processChannelGraphEvent(ctx, broadcaster, broadcaster.SubscribeChannelGraphEvent())
	go processForwardEvent(ctx, broadcaster, broadcaster.SubscribeForwardEvent())
	go processInvoiceEvent(ctx, broadcaster, broadcaster.SubscribeInvoiceEvent())
	go processPaymentEvent(ctx, broadcaster, broadcaster.SubscribePaymentEvent())
	go processPeerEvent(ctx, broadcaster, broadcaster.SubscribePeerEvent())

	for {
		select {
		case <-ctx.Done():
			return
		case <-lndSyncTicker:
			bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping,
				serviceStatus, subscriptionStream, lndClient, db, mutex)
		case <-initiateSync:
			bootStrapping, serviceStatus = synchronizeDataFromLnd(nodeSettings, bootStrapping,
				serviceStatus, subscriptionStream, lndClient, db, mutex)
		}
	}
}

func synchronizeDataFromLnd(nodeSettings commons.ManagedNodeSettings,
	bootStrapping bool,
	serviceStatus commons.ServiceStatus,
	subscriptionStream commons.SubscriptionStream,
	lndClient lnrpc.LightningClient,
	db *sqlx.DB,
	mutex *sync.RWMutex) (bool, commons.ServiceStatus) {

	if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.ServiceActive {
		if !bootStrapping {
			bootStrapping = true
			log.Error().Msgf("Channel balance cache got out-of-sync because of a non-active LND stream. (nodeId: %v)", nodeSettings.NodeId)
		}
	}
	if bootStrapping {
		serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServiceInitializing)
	} else {
		serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServiceActive)
	}
	err := initializeChannelBalanceFromLnd(lndClient, nodeSettings.NodeId, db, mutex)
	if err != nil {
		serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServicePending)
		log.Error().Err(err).Msgf("Failed to initialize channel balance cache. This is a critical issue! (nodeId: %v)", nodeSettings.NodeId)
		return bootStrapping, serviceStatus
	}
	if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) == commons.ServiceActive {
		bootStrapping = false
		serviceStatus = SetStreamStatus(nodeSettings.NodeId, subscriptionStream, serviceStatus, commons.ServiceActive)
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
		localRoutingPolicy, err := channels.GetLocalRoutingPolicy(channelId, nodeId, db)
		if err != nil {
			return errors.Wrapf(err, "Obtaining LocalRoutingPolicy from the database for channelId: %v", channelId)
		}
		channelStateSettings.LocalDisabled = localRoutingPolicy.Disabled
		channelStateSettings.LocalFeeBaseMsat = localRoutingPolicy.FeeBaseMsat
		channelStateSettings.LocalFeeRateMilliMsat = localRoutingPolicy.FeeRateMillMsat
		channelStateSettings.LocalMinHtlcMsat = localRoutingPolicy.MinHtlcMsat
		channelStateSettings.LocalMaxHtlcMsat = localRoutingPolicy.MaxHtlcMsat
		channelStateSettings.LocalTimeLockDelta = localRoutingPolicy.TimeLockDelta

		remoteRoutingPolicy, err := channels.GetRemoteRoutingPolicy(channelId, nodeId, db)
		if err != nil {
			return errors.Wrapf(err, "Obtaining RemoteRoutingPolicy from the database for channelId: %v", channelId)
		}
		channelStateSettings.RemoteDisabled = remoteRoutingPolicy.Disabled
		channelStateSettings.RemoteFeeBaseMsat = remoteRoutingPolicy.FeeBaseMsat
		channelStateSettings.RemoteFeeRateMilliMsat = remoteRoutingPolicy.FeeRateMillMsat
		channelStateSettings.RemoteMinHtlcMsat = remoteRoutingPolicy.MinHtlcMsat
		channelStateSettings.RemoteMaxHtlcMsat = remoteRoutingPolicy.MaxHtlcMsat
		channelStateSettings.RemoteTimeLockDelta = remoteRoutingPolicy.TimeLockDelta

		pendingIncomingHtlcCount := 0
		pendingIncomingHtlcAmount := int64(0)
		pendingOutgoingHtlcCount := 0
		pendingOutgoingHtlcAmount := int64(0)
		if len(lndChannel.PendingHtlcs) != 0 {
			for _, pendingHtlc := range lndChannel.PendingHtlcs {
				if pendingHtlc.Incoming {
					channelStateSettings.RemoteBalance += pendingHtlc.Amount
				} else {
					channelStateSettings.LocalBalance += pendingHtlc.Amount
				}
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
			channelSettings := commons.GetChannelSettingByChannelId(channelId)
			if channelStateSettings.RemoteBalance+channelStateSettings.LocalBalance > channelSettings.Capacity {
				log.Error().Msgf("ChannelBalanceCacheMaintenance: RemoteBalance (%v) + LocalBalance (%v) > Capacity (%v) for channelId: %v",
					channelStateSettings.RemoteBalance, channelStateSettings.LocalBalance, channelSettings.Capacity,
					channelId)
			}
			tolerance := lndChannel.GetLocalConstraints().ChanReserveSat + lndChannel.GetLocalConstraints().DustLimitSat
			remoteTolerance := lndChannel.GetRemoteConstraints().ChanReserveSat + lndChannel.GetRemoteConstraints().DustLimitSat
			if tolerance < remoteTolerance {
				tolerance = remoteTolerance
			}
			if channelSettings.Capacity-(channelStateSettings.RemoteBalance+channelStateSettings.LocalBalance) > int64(tolerance) {
				log.Error().Msgf("ChannelBalanceCacheMaintenance: Capacity (%v) - ( RemoteBalance (%v) + LocalBalance (%v) ) > %v for channelId: %v",
					channelSettings.Capacity, channelStateSettings.RemoteBalance, channelStateSettings.LocalBalance,
					tolerance, channelId)
			}
		}
		channelStateSettings.PendingIncomingHtlcCount = pendingIncomingHtlcCount
		channelStateSettings.PendingIncomingHtlcAmount = pendingIncomingHtlcAmount
		channelStateSettings.PendingOutgoingHtlcCount = pendingOutgoingHtlcCount
		channelStateSettings.PendingOutgoingHtlcAmount = pendingOutgoingHtlcAmount
		channelStateSettingsList = append(channelStateSettingsList, channelStateSettings)
	}
	commons.SetChannelStates(nodeId, channelStateSettingsList)
	serviceStatus := commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeId)
	if serviceStatus < commons.ServiceBootRequested {
		commons.SetChannelStateNodeStatus(nodeId, commons.Status(serviceStatus))
	} else {
		commons.SetChannelStateNodeStatus(nodeId, commons.Inactive)
	}
	return nil
}

func processServiceEvent(ctx context.Context,
	initiateSync chan<- bool,
	broadcaster broadcast.BroadcastServer,
	listener <-chan commons.ServiceEvent) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionServiceEvent(listener)
			return
		}
	}()
	go func() {
		for serviceEvent := range listener {
			if serviceEvent.NodeId == 0 || serviceEvent.Type != commons.LndService {
				continue
			}
			if serviceEvent.SubscriptionStream == nil {
				if serviceEvent.Status == commons.ServiceActive {
					initiateSync <- true
				}
				continue
			}
			if !serviceEvent.SubscriptionStream.IsChannelBalanceCache() {
				continue
			}
			channelBalanceStreamStatus :=
				commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(serviceEvent.NodeId)
			if channelBalanceStreamStatus < commons.ServiceBootRequested {
				commons.SetChannelStateNodeStatus(serviceEvent.NodeId, commons.Status(channelBalanceStreamStatus))
			} else {
				commons.SetChannelStateNodeStatus(serviceEvent.NodeId, commons.Inactive)
			}
		}
	}()
	wg.Wait()
}

func processChannelEvent(ctx context.Context,
	broadcaster broadcast.BroadcastServer,
	listener <-chan commons.ChannelEvent) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionChannelEvent(listener)
			return
		}
	}()
	go func() {
		for channelEvent := range listener {
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
	}()
	wg.Wait()
}

func processChannelGraphEvent(ctx context.Context,
	broadcaster broadcast.BroadcastServer,
	listener <-chan commons.ChannelGraphEvent) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionChannelGraphEvent(listener)
			return
		}
	}()
	go func() {
		for channelGraphEvent := range listener {
			if channelGraphEvent.NodeId == 0 ||
				channelGraphEvent.ChannelId == nil || *channelGraphEvent.ChannelId == 0 ||
				channelGraphEvent.AnnouncingNodeId == nil || *channelGraphEvent.AnnouncingNodeId == 0 ||
				channelGraphEvent.ConnectingNodeId == nil || *channelGraphEvent.ConnectingNodeId == 0 {
				continue
			}
			local := *channelGraphEvent.AnnouncingNodeId == channelGraphEvent.NodeId
			commons.SetChannelStateRoutingPolicy(channelGraphEvent.NodeId, *channelGraphEvent.ChannelId, local,
				channelGraphEvent.Disabled, channelGraphEvent.TimeLockDelta, channelGraphEvent.MinHtlcMsat,
				channelGraphEvent.MaxHtlcMsat, channelGraphEvent.FeeBaseMsat, channelGraphEvent.FeeRateMilliMsat)
		}
	}()
	wg.Wait()
}

func processForwardEvent(ctx context.Context,
	broadcaster broadcast.BroadcastServer,
	listener <-chan commons.ForwardEvent) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionForwardEvent(listener)
			return
		}
	}()
	go func() {
		for forwardEvent := range listener {
			if forwardEvent.NodeId == 0 {
				continue
			}
			if forwardEvent.IncomingChannelId != nil {
				commons.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.IncomingChannelId, true,
					forwardEvent.AmountInMsat, commons.BalanceUpdateForwardEvent)
			}
			if forwardEvent.OutgoingChannelId != nil {
				commons.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.OutgoingChannelId, false,
					forwardEvent.AmountOutMsat, commons.BalanceUpdateForwardEvent)
			}
		}
	}()
	wg.Wait()
}

func processInvoiceEvent(ctx context.Context,
	broadcaster broadcast.BroadcastServer,
	listener <-chan commons.InvoiceEvent) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionInvoiceEvent(listener)
			return
		}
	}()
	go func() {
		for invoiceEvent := range listener {
			if invoiceEvent.NodeId == 0 || invoiceEvent.State != lnrpc.Invoice_SETTLED {
				continue
			}
			commons.SetChannelStateBalanceUpdateMsat(invoiceEvent.NodeId, invoiceEvent.ChannelId, true,
				invoiceEvent.AmountPaidMsat, commons.BalanceUpdateInvoiceEvent)
		}
	}()
	wg.Wait()
}

func processPaymentEvent(ctx context.Context,
	broadcaster broadcast.BroadcastServer,
	listener <-chan commons.PaymentEvent) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionPaymentEvent(listener)
			return
		}
	}()
	go func() {
		for paymentEvent := range listener {
			if paymentEvent.NodeId == 0 ||
				paymentEvent.OutgoingChannelId == nil || *paymentEvent.OutgoingChannelId == 0 ||
				paymentEvent.PaymentStatus != lnrpc.Payment_SUCCEEDED {
				continue
			}
			commons.SetChannelStateBalanceUpdate(paymentEvent.NodeId, *paymentEvent.OutgoingChannelId, false,
				paymentEvent.AmountPaid, commons.BalanceUpdatePaymentEvent)
		}
	}()
	wg.Wait()
}

func processPeerEvent(ctx context.Context,
	broadcaster broadcast.BroadcastServer,
	listener <-chan commons.PeerEvent) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionPeerEvent(listener)
			return
		}
	}()
	go func() {
		for peerEvent := range listener {
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
		}
	}()
	wg.Wait()
}
