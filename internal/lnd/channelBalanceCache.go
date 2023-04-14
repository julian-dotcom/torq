package lnd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/core"
)

const channelbalanceTickerSeconds = 150

func ChannelBalanceCacheMaintenance(ctx context.Context,
	lndClient lnrpc.LightningClient,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := core.LndServiceChannelBalanceCacheStream

	bootStrapping := true
	lndSyncTicker := time.NewTicker(channelbalanceTickerSeconds * time.Second)
	defer lndSyncTicker.Stop()
	fastTicker := time.NewTicker(10 * time.Second)
	defer fastTicker.Stop()
	mutex := &sync.RWMutex{}
	channelBalanceStreamActive := false
	initiateSync := true
	var err error

	for {
		if initiateSync {
			bootStrapping, err = synchronizeDataFromLnd(nodeSettings, bootStrapping, serviceType, lndClient, db, mutex)
			if err != nil {
				log.Error().Err(err).Msgf("Channel balance synchronization failed for nodeId: %v", nodeSettings.NodeId)
				cache.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			initiateSync = false
		}
		select {
		case <-ctx.Done():
			cache.SetInactiveLndServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-fastTicker.C:
			if cache.IsChannelBalanceCacheStreamActive(nodeSettings.NodeId) {
				if !channelBalanceStreamActive {
					initiateSync = true
					cache.SetChannelStateNodeStatus(nodeSettings.NodeId, core.Active)
					channelBalanceStreamActive = true
				}
				// code stops here when all channel balance streams are active
				continue
			}
			// channel balance streams are not all active here
			if bootStrapping {
				cache.SetInitializingLndServiceState(serviceType, nodeSettings.NodeId)
			}
			if channelBalanceStreamActive {
				cache.SetChannelStateNodeStatus(nodeSettings.NodeId, core.Inactive)
				channelBalanceStreamActive = false
			}
		case <-lndSyncTicker.C:
			bootStrapping, err = synchronizeDataFromLnd(nodeSettings, bootStrapping, serviceType, lndClient, db, mutex)
			if err != nil {
				log.Error().Err(err).Msgf("Channel balance synchronization failed for nodeId: %v", nodeSettings.NodeId)
				cache.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
				return
			}
		}
	}
}

func synchronizeDataFromLnd(nodeSettings cache.NodeSettingsCache,
	bootStrapping bool,
	serviceType core.ServiceType,
	lndClient lnrpc.LightningClient,
	db *sqlx.DB,
	mutex *sync.RWMutex) (bool, error) {

	if !cache.IsLndServiceActive(nodeSettings.NodeId) {
		if !bootStrapping {
			bootStrapping = true
			log.Error().Msgf("Channel balance cache got out-of-sync because of a non-active LND stream. (nodeId: %v)", nodeSettings.NodeId)
		}
	}
	if bootStrapping {
		cache.SetInitializingLndServiceState(serviceType, nodeSettings.NodeId)
	}
	err := initializeChannelBalanceFromLnd(lndClient, nodeSettings.NodeId, db, mutex)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to initialize channel balance cache. This is a critical issue! (nodeId: %v)", nodeSettings.NodeId)
		cache.SetFailedLndServiceState(serviceType, nodeSettings.NodeId)
		return bootStrapping, err
	}
	if cache.IsChannelBalanceCacheStreamActive(nodeSettings.NodeId) {
		if bootStrapping {
			bootStrapping = false
			cache.SetActiveLndServiceState(serviceType, nodeSettings.NodeId)
		}
	}
	return bootStrapping, nil
}

func initializeChannelBalanceFromLnd(lndClient lnrpc.LightningClient, nodeId int, db *sqlx.DB, mutex *sync.RWMutex) error {
	if core.RWMutexWriteLocked(mutex) {
		log.Error().Msgf("The lock initializeChannelBalanceFromLnd is already locked? This is a critical issue! (nodeId: %v)", nodeId)
		return errors.New(fmt.Sprintf("The lock initializeChannelBalanceFromLnd is already locked? This is a critical issue! (nodeId: %v)", nodeId))
	}
	nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()
	var channelStateSettingsList []cache.ChannelStateSettingsCache
	r, err := lndClient.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		return errors.Wrapf(err, "Obtaining channels from LND for nodeId: %v", nodeId)
	}
	for _, lndChannel := range r.Channels {
		channelId := cache.GetChannelIdByChannelPoint(lndChannel.ChannelPoint)
		remoteNodeId := cache.GetNodeIdByPublicKey(lndChannel.RemotePubkey, nodeSettings.Chain, nodeSettings.Network)
		if channelId == 0 {
			return errors.Wrapf(err, "Obtaining channelId from channelPoint: %v", lndChannel.ChannelPoint)
		}
		channelStateSettings := cache.ChannelStateSettingsCache{
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
				htlc := cache.Htlc{
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

		channelStateSettings, err = verifyChannelCapacityMismatch(channelStateSettings, channelId, nodeId, lndChannel)
		if err != nil {
			return errors.Wrapf(err, "capacity mismatch found for cache initialization data.")
		}

		channelStateSettingsList = append(channelStateSettingsList, channelStateSettings)
	}
	cache.SetChannelStates(nodeId, channelStateSettingsList)
	return nil
}

func verifyChannelCapacityMismatch(channelStateSettings cache.ChannelStateSettingsCache,
	channelId int,
	nodeId int,
	lndChannel *lnrpc.Channel) (cache.ChannelStateSettingsCache, error) {

	channelSettings := cache.GetChannelSettingByChannelId(channelId)

	if channelStateSettings.RemoteBalance < 0 {
		log.Error().Msgf("ChannelBalanceCacheMaintenance: RemoteBalance (%v) < 0 for channelId: %v",
			channelStateSettings.RemoteBalance, channelId)
		logLndChannelDebugData(lndChannel)
		existingSettings := cache.GetChannelState(nodeId, channelId, true)
		if existingSettings == nil {
			return cache.ChannelStateSettingsCache{},
				errors.New(fmt.Sprintf("Capacity mismatch found and no fallback available for nodeId: %v", nodeId))
		}
		return *existingSettings, nil
	}

	if channelStateSettings.LocalBalance < 0 {
		log.Error().Msgf("ChannelBalanceCacheMaintenance: LocalBalance (%v) < 0 for channelId: %v",
			channelStateSettings.LocalBalance, channelId)
		logLndChannelDebugData(lndChannel)
		existingSettings := cache.GetChannelState(nodeId, channelId, true)
		if existingSettings == nil {
			return cache.ChannelStateSettingsCache{},
				errors.New(fmt.Sprintf("Capacity mismatch found and no fallback available for nodeId: %v", nodeId))
		}
		return *existingSettings, nil
	}
	localPlusRemote := channelStateSettings.RemoteBalance + channelStateSettings.LocalBalance
	if localPlusRemote > channelSettings.Capacity {
		log.Error().Msgf("ChannelBalanceCacheMaintenance: RemoteBalance (%v) + LocalBalance (%v) > Capacity (%v) for channelId: %v",
			channelStateSettings.RemoteBalance, channelStateSettings.LocalBalance, channelSettings.Capacity,
			channelId)
		logLndChannelDebugData(lndChannel)
		existingSettings := cache.GetChannelState(nodeId, channelId, true)
		if existingSettings == nil {
			return cache.ChannelStateSettingsCache{},
				errors.New(fmt.Sprintf("Capacity mismatch found and no fallback available for nodeId: %v", nodeId))
		}
		return *existingSettings, nil
	}

	tolerance := lndChannel.GetLocalConstraints().ChanReserveSat + lndChannel.GetLocalConstraints().DustLimitSat
	remoteTolerance := lndChannel.GetRemoteConstraints().ChanReserveSat + lndChannel.GetRemoteConstraints().DustLimitSat
	if tolerance < remoteTolerance {
		tolerance = remoteTolerance
	}
	tolerance = tolerance + uint64(core.Abs(lndChannel.CommitFee))
	if channelSettings.Capacity-localPlusRemote > int64(tolerance) {
		log.Error().Msgf("ChannelBalanceCacheMaintenance: Capacity (%v) - ( RemoteBalance (%v) + LocalBalance (%v) ) > %v for channelId: %v",
			channelSettings.Capacity, channelStateSettings.RemoteBalance, channelStateSettings.LocalBalance,
			tolerance, channelId)
		logLndChannelDebugData(lndChannel)
		existingSettings := cache.GetChannelState(nodeId, channelId, true)
		if existingSettings == nil {
			return cache.ChannelStateSettingsCache{},
				errors.New(fmt.Sprintf("Capacity mismatch found and no fallback available for nodeId: %v", nodeId))
		}
		return *existingSettings, nil
	}
	return channelStateSettings, nil
}

func logLndChannelDebugData(lndChannel *lnrpc.Channel) {
	marshalledLndChannel, err := json.Marshal(lndChannel)
	if err != nil {
		log.Error().Err(err).Msgf("ChannelBalanceCacheMaintenance: failed to marshal lnrpc data: %v", lndChannel)
	}
	if err == nil {
		log.Error().Msgf("ChannelBalanceCacheMaintenance: lnrpc channel data: %v", string(marshalledLndChannel))
	}
}

func ProcessChannelEvent(channelEvent core.ChannelEvent) {
	if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
		return
	}
	var status core.Status
	switch channelEvent.Type {
	case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
		status = core.Active
	case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
		status = core.Inactive
	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		status = core.Deleted
	}
	cache.SetChannelStateChannelStatus(channelEvent.NodeId, channelEvent.ChannelId, status)
}

func ProcessChannelGraphEvent(channelGraphEvent core.ChannelGraphEvent) {
	if channelGraphEvent.NodeId == 0 ||
		channelGraphEvent.ChannelId == nil || *channelGraphEvent.ChannelId == 0 ||
		channelGraphEvent.AnnouncingNodeId == nil || *channelGraphEvent.AnnouncingNodeId == 0 ||
		channelGraphEvent.ConnectingNodeId == nil || *channelGraphEvent.ConnectingNodeId == 0 {
		return
	}
	local := *channelGraphEvent.AnnouncingNodeId == channelGraphEvent.NodeId
	cache.SetChannelStateRoutingPolicy(channelGraphEvent.NodeId, *channelGraphEvent.ChannelId, local,
		channelGraphEvent.Disabled, channelGraphEvent.TimeLockDelta, channelGraphEvent.MinHtlcMsat,
		channelGraphEvent.MaxHtlcMsat, channelGraphEvent.FeeBaseMsat, channelGraphEvent.FeeRateMilliMsat)
}

func ProcessForwardEvent(forwardEvent core.ForwardEvent) {
	if forwardEvent.NodeId == 0 {
		return
	}
	if forwardEvent.IncomingChannelId != nil {
		cache.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.IncomingChannelId, true,
			forwardEvent.AmountInMsat, core.BalanceUpdateForwardEvent)
	}
	if forwardEvent.OutgoingChannelId != nil {
		cache.SetChannelStateBalanceUpdateMsat(forwardEvent.NodeId, *forwardEvent.OutgoingChannelId, false,
			forwardEvent.AmountOutMsat, core.BalanceUpdateForwardEvent)
	}
}

func ProcessInvoiceEvent(invoiceEvent core.InvoiceEvent) {
	if invoiceEvent.NodeId == 0 || invoiceEvent.State != lnrpc.Invoice_SETTLED {
		return
	}
	cache.SetChannelStateBalanceUpdateMsat(invoiceEvent.NodeId, invoiceEvent.ChannelId, true,
		invoiceEvent.AmountPaidMsat, core.BalanceUpdateInvoiceEvent)
}

func ProcessPaymentEvent(paymentEvent core.PaymentEvent) {
	if paymentEvent.NodeId == 0 ||
		paymentEvent.OutgoingChannelId == nil || *paymentEvent.OutgoingChannelId == 0 ||
		paymentEvent.PaymentStatus != lnrpc.Payment_SUCCEEDED {
		return
	}
	cache.SetChannelStateBalanceUpdate(paymentEvent.NodeId, *paymentEvent.OutgoingChannelId, false,
		paymentEvent.AmountPaid, core.BalanceUpdatePaymentEvent)
}

func ProcessPeerEvent(peerEvent core.PeerEvent) {
	if peerEvent.NodeId == 0 || peerEvent.EventNodeId == 0 {
		return
	}
	var status core.Status
	switch peerEvent.Type {
	case lnrpc.PeerEvent_PEER_ONLINE:
		status = core.Active
	case lnrpc.PeerEvent_PEER_OFFLINE:
		status = core.Inactive
	}
	channelIds := cache.GetChannelIdsByNodeId(peerEvent.EventNodeId)
	for _, channelId := range channelIds {
		cache.SetChannelStateChannelStatus(peerEvent.NodeId, channelId, status)
	}
}
