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
	"github.com/lncapital/torq/pkg/commons"
)

const channelbalanceTickerSeconds = 150

func ChannelBalanceCacheMaintenance(ctx context.Context,
	lndClient lnrpc.LightningClient,
	db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings) {

	serviceType := commons.LndServiceChannelBalanceCacheStream

	bootStrapping := true
	lndSyncTicker := clock.New().Tick(channelbalanceTickerSeconds * time.Second)
	fastTicker := clock.New().Tick(10 * time.Second)
	mutex := &sync.RWMutex{}
	channelBalanceStreamActive := false
	initiateSync := true
	var err error

	for {
		if initiateSync {
			bootStrapping, err = synchronizeDataFromLnd(nodeSettings, bootStrapping, serviceType, lndClient, db, mutex)
			if err != nil {
				log.Error().Err(err).Msgf("Channel balance synchronization failed for nodeId: %v", nodeSettings.NodeId)
				commons.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			initiateSync = false
		}
		select {
		case <-ctx.Done():
			commons.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-fastTicker:
			if commons.IsChannelBalanceCacheStreamActive(nodeSettings.NodeId) {
				if !channelBalanceStreamActive {
					initiateSync = true
					commons.SetChannelStateNodeStatus(nodeSettings.NodeId, commons.Active)
					channelBalanceStreamActive = true
				}
				// code stops here when all channel balance streams are active
				continue
			}
			// channel balance streams are not all active here
			if channelBalanceStreamActive {
				commons.SetChannelStateNodeStatus(nodeSettings.NodeId, commons.Inactive)
				channelBalanceStreamActive = false
			}
		case <-lndSyncTicker:
			bootStrapping, err = synchronizeDataFromLnd(nodeSettings, bootStrapping, serviceType, lndClient, db, mutex)
			if err != nil {
				log.Error().Err(err).Msgf("Channel balance synchronization failed for nodeId: %v", nodeSettings.NodeId)
				commons.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
				return
			}
		}
	}
}

func synchronizeDataFromLnd(nodeSettings commons.ManagedNodeSettings,
	bootStrapping bool,
	serviceType commons.ServiceType,
	lndClient lnrpc.LightningClient,
	db *sqlx.DB,
	mutex *sync.RWMutex) (bool, error) {

	if !commons.IsLndServiceActive(nodeSettings.NodeId) {
		if !bootStrapping {
			bootStrapping = true
			log.Error().Msgf("Channel balance cache got out-of-sync because of a non-active LND stream. (nodeId: %v)", nodeSettings.NodeId)
		}
	}
	if bootStrapping {
		commons.SetInitializingLndServiceState(serviceType, nodeSettings.NodeId)
	}
	err := initializeChannelBalanceFromLnd(lndClient, nodeSettings.NodeId, db, mutex)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to initialize channel balance cache. This is a critical issue! (nodeId: %v)", nodeSettings.NodeId)
		commons.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
		return bootStrapping, err
	}
	if commons.IsChannelBalanceCacheStreamActive(nodeSettings.NodeId) {
		bootStrapping = false
		commons.SetActiveLndServiceState(serviceType, nodeSettings.NodeId)
	}
	return bootStrapping, nil
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
	return nil
}

func ProcessChannelEvent(channelEvent commons.ChannelEvent) {
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
	commons.SetChannelStateChannelStatus(channelEvent.NodeId, channelEvent.ChannelId, status)
}

func ProcessChannelGraphEvent(channelGraphEvent commons.ChannelGraphEvent) {
	if channelGraphEvent.NodeId == 0 ||
		channelGraphEvent.ChannelId == nil || *channelGraphEvent.ChannelId == 0 ||
		channelGraphEvent.AnnouncingNodeId == nil || *channelGraphEvent.AnnouncingNodeId == 0 ||
		channelGraphEvent.ConnectingNodeId == nil || *channelGraphEvent.ConnectingNodeId == 0 {
		return
	}
	local := *channelGraphEvent.AnnouncingNodeId == channelGraphEvent.NodeId
	commons.SetChannelStateRoutingPolicy(channelGraphEvent.NodeId, *channelGraphEvent.ChannelId, local,
		channelGraphEvent.Disabled, channelGraphEvent.TimeLockDelta, channelGraphEvent.MinHtlcMsat,
		channelGraphEvent.MaxHtlcMsat, channelGraphEvent.FeeBaseMsat, channelGraphEvent.FeeRateMilliMsat)
}

func ProcessForwardEvent(forwardEvent commons.ForwardEvent) {
	if forwardEvent.NodeId == 0 {
		return
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

func ProcessInvoiceEvent(invoiceEvent commons.InvoiceEvent) {
	if invoiceEvent.NodeId == 0 || invoiceEvent.State != lnrpc.Invoice_SETTLED {
		return
	}
	commons.SetChannelStateBalanceUpdateMsat(invoiceEvent.NodeId, invoiceEvent.ChannelId, true,
		invoiceEvent.AmountPaidMsat, commons.BalanceUpdateInvoiceEvent)
}

func ProcessPaymentEvent(paymentEvent commons.PaymentEvent) {
	if paymentEvent.NodeId == 0 ||
		paymentEvent.OutgoingChannelId == nil || *paymentEvent.OutgoingChannelId == 0 ||
		paymentEvent.PaymentStatus != lnrpc.Payment_SUCCEEDED {
		return
	}
	commons.SetChannelStateBalanceUpdate(paymentEvent.NodeId, *paymentEvent.OutgoingChannelId, false,
		paymentEvent.AmountPaid, commons.BalanceUpdatePaymentEvent)
}

func ProcessPeerEvent(peerEvent commons.PeerEvent) {
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
	for _, channelId := range channelIds {
		commons.SetChannelStateChannelStatus(peerEvent.NodeId, channelId, status)
	}
}
