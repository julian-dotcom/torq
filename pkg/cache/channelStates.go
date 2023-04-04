package cache

import (
	"context"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/pkg/core"
)

const toleratedSubscriptionDowntimeSeconds = 15

var ChannelStatesCacheChannel = make(chan ChannelStateCache)    //nolint:gochecknoglobals
var ChannelBalanceChanges = make(chan core.ChannelBalanceEvent) //nolint:gochecknoglobals

type ChannelStateCacheOperationType uint

const (
	// readChannelState please provide NodeId, ChannelId and StateOut
	readChannelState ChannelStateCacheOperationType = iota
	// readAllChannelStates please provide NodeId and StatesOut
	readAllChannelStates
	// readAllChannelStateChannelIds please provide NodeId and ChannelIdsOut
	readAllChannelStateChannelIds
	// readSharedChannelStateChannelIds please provide NodeId and ChannelIdsOut
	readSharedChannelStateChannelIds
	// readChannelBalanceState please provide NodeId, ChannelId, HtlcInclude and BalanceStateOut
	readChannelBalanceState
	// readAllChannelBalanceStates please provide NodeId, StateInclude, HtlcInclude and BalanceStatesOut
	readAllChannelBalanceStates
	writeInitialChannelState
	// writeInitialChannelStates This requires the lock being active for writing! Please provide the complete information set
	writeInitialChannelStates
	writeChannelStateNodeStatus
	writeChannelStateChannelStatus
	writeChannelStateRoutingPolicy
	writeChannelStateUpdateBalance
	writeChannelStateUpdateHtlcEvent
	removeChannelStatesFromCache
	removeChannelStateFromCache
)

type ChannelBalanceStateHtlcInclude uint

const (
	// pendingHtlcsLocalBalanceAdjustedDownwards:
	//   LocalBalance = ConfirmedLocalBalance - PendingDecreasingForwardHTLCsAmount - PendingPaymentHTLCsAmount
	pendingHtlcsLocalBalanceAdjustedDownwards ChannelBalanceStateHtlcInclude = iota
	// pendingHtlcsRemoteBalanceAdjustedDownwards:
	//   RemoteBalance = ConfirmedRemoteBalance - PendingIncreasingForwardHTLCsAmount - PendingInvoiceHTLCsAmount
	pendingHtlcsRemoteBalanceAdjustedDownwards
	// pendingHtlcsLocalAndRemoteBalanceAdjustedDownwards:
	//   LocalBalance = ConfirmedLocalBalance - PendingDecreasingForwardHTLCsAmount - PendingPaymentHTLCsAmount
	//   RemoteBalance = ConfirmedRemoteBalance - PendingIncreasingForwardHTLCsAmount - PendingInvoiceHTLCsAmount
	pendingHtlcsLocalAndRemoteBalanceAdjustedDownwards
	// pendingHtlcsLocalBalanceAdjustedUpwards:
	//   LocalBalance = ConfirmedLocalBalance + PendingIncreasingForwardHTLCsAmount + PendingInvoiceHTLCsAmount
	pendingHtlcsLocalBalanceAdjustedUpwards
	//   RemoteBalance = ConfirmedRemoteBalance + PendingDecreasingForwardHTLCsAmount + PendingPaymentHTLCsAmount
	pendingHtlcsRemoteBalanceAdjustedUpwards
	// pendingHtlcsLocalAndRemoteBalanceAdjustedUpwards:
	//   LocalBalance = ConfirmedLocalBalance + PendingIncreasingForwardHTLCsAmount + PendingInvoiceHTLCsAmount
	//   RemoteBalance = ConfirmedRemoteBalance + PendingDecreasingForwardHTLCsAmount + PendingPaymentHTLCsAmount
	pendingHtlcsLocalAndRemoteBalanceAdjustedUpwards
)

type ChannelStateInclude uint

const (
	allLocalAndRemoteActiveChannels ChannelStateInclude = iota
	allChannels
)

type Htlc struct {
	Incoming         bool
	Amount           int64
	HashLock         []byte
	ExpirationHeight uint32
	// Index identifying the htlc on the channel.
	HtlcIndex uint64
	// If this HTLC is involved in a forwarding operation, this field indicates
	// the forwarding channel. For an outgoing htlc, it is the incoming channel.
	// For an incoming htlc, it is the outgoing channel. When the htlc
	// originates from this node or this node is the final destination,
	// forwarding_channel will be zero. The forwarding channel will also be zero
	// for htlcs that need to be forwarded but don't have a forwarding decision
	// persisted yet.
	ForwardingChannel uint64
	// Index identifying the htlc on the forwarding channel.
	ForwardingHtlcIndex uint64
}

type ChannelStateCache struct {
	Type                     ChannelStateCacheOperationType
	NodeId                   int
	RemoteNodeId             int
	ChannelId                int
	Status                   core.Status
	Local                    bool
	Balance                  int64
	Disabled                 bool
	FeeBaseMsat              int64
	FeeRateMilliMsat         int64
	MinHtlcMsat              uint64
	MaxHtlcMsat              uint64
	TimeLockDelta            uint32
	Amount                   int64
	ForceResponse            bool
	BalanceUpdateEventOrigin core.BalanceUpdateEventOrigin
	HtlcEvent                core.HtlcEvent
	ChannelStateSetting      ChannelStateSettingsCache
	ChannelStateSettings     []ChannelStateSettingsCache
	HtlcInclude              ChannelBalanceStateHtlcInclude
	ChannelIdsOut            chan<- []int
	StateInclude             ChannelStateInclude
	StateOut                 chan<- *ChannelStateSettingsCache
	StatesOut                chan<- []ChannelStateSettingsCache
	BalanceStateOut          chan<- *ChannelBalanceStateSettingsCache
	BalanceStatesOut         chan<- []ChannelBalanceStateSettingsCache
}

type ChannelStateSettingsCache struct {
	NodeId       int `json:"nodeId"`
	RemoteNodeId int `json:"remoteNodeId"`
	ChannelId    int `json:"channelId"`

	LocalBalance          int64  `json:"localBalance"`
	LocalDisabled         bool   `json:"localDisabled"`
	LocalFeeBaseMsat      int64  `json:"localFeeBaseMsat"`
	LocalFeeRateMilliMsat int64  `json:"localFeeRateMilliMsat"`
	LocalMinHtlcMsat      uint64 `json:"localMinHtlcMsat"`
	LocalMaxHtlcMsat      uint64 `json:"localMaxHtlcMsat"`
	LocalTimeLockDelta    uint32 `json:"localTimeLockDelta"`

	RemoteBalance          int64  `json:"remoteBalance"`
	RemoteDisabled         bool   `json:"remoteDisabled"`
	RemoteFeeBaseMsat      int64  `json:"remoteFeeBaseMsat"`
	RemoteFeeRateMilliMsat int64  `json:"remoteFeeRateMilliMsat"`
	RemoteMinHtlcMsat      uint64 `json:"remoteMinHtlcMsat"`
	RemoteMaxHtlcMsat      uint64 `json:"remoteMaxHtlcMsat"`
	RemoteTimeLockDelta    uint32 `json:"remoteTimeLockDelta"`

	UnsettledBalance int64 `json:"unsettledBalance"`

	PendingHtlcs []Htlc `json:"pendingHtlcs"`
	// INCREASING LOCAL BALANCE HTLCs
	PendingIncomingHtlcCount  int   `json:"pendingIncomingHtlcCount"`
	PendingIncomingHtlcAmount int64 `json:"pendingIncomingHtlcAmount"`
	// DECREASING LOCAL BALANCE HTLCs
	PendingOutgoingHtlcCount  int   `json:"pendingOutgoingHtlcCount"`
	PendingOutgoingHtlcAmount int64 `json:"pendingOutgoingHtlcAmount"`

	// STALE INFORMATION ONLY OBTAINED VIA LND REGULAR CHECKINS SO NOT MAINTAINED
	CommitFee             int64                `json:"commitFee"`
	CommitWeight          int64                `json:"commitWeight"`
	FeePerKw              int64                `json:"feePerKw"`
	NumUpdates            uint64               `json:"numUpdates"`
	ChanStatusFlags       string               `json:"chanStatusFlags"`
	CommitmentType        lnrpc.CommitmentType `json:"commitmentType"`
	Lifetime              int64                `json:"lifetime"`
	TotalSatoshisReceived int64                `json:"totalSatoshisReceived"`
	TotalSatoshisSent     int64                `json:"totalSatoshisSent"`
	PeerChannelCount      int                  `json:"peerChannelCount"`
	PeerChannelCapacity   int64                `json:"peerChannelCapacity"`
	PeerLocalBalance      int64                `json:"peerRemoteBalance"`
}

type ChannelBalanceStateSettingsCache struct {
	NodeId                     int                            `json:"nodeId"`
	RemoteNodeId               int                            `json:"remoteNodeId"`
	ChannelId                  int                            `json:"channelId"`
	HtlcInclude                ChannelBalanceStateHtlcInclude `json:"htlcInclude"`
	LocalBalance               int64                          `json:"localBalance"`
	LocalBalancePerMilleRatio  int                            `json:"localBalancePerMilleRatio"`
	RemoteBalance              int64                          `json:"remoteBalance"`
	RemoteBalancePerMilleRatio int                            `json:"remoteBalancePerMilleRatio"`
}

// ChannelStatesCacheHandler parameter Context is for test cases...
func ChannelStatesCacheHandler(ch <-chan ChannelStateCache, ctx context.Context) {
	channelStateSettingsByChannelIdCache := make(map[nodeId]map[channelId]ChannelStateSettingsCache, 0)
	channelStateSettingsStatusCache := make(map[nodeId]core.Status, 0)
	channelStateSettingsDeactivationTimeCache := make(map[nodeId]time.Time, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case channelStateCache := <-ch:
			handleChannelStateSettingsOperation(channelStateCache,
				channelStateSettingsStatusCache, channelStateSettingsByChannelIdCache,
				channelStateSettingsDeactivationTimeCache)
		}
	}
}

func handleChannelStateSettingsOperation(channelStateCache ChannelStateCache,
	channelStateSettingsStatusCache map[nodeId]core.Status,
	channelStateSettingsByChannelIdCache map[nodeId]map[channelId]ChannelStateSettingsCache,
	channelStateSettingsDeactivationTimeCache map[nodeId]time.Time) {
	switch channelStateCache.Type {
	case readChannelState:
		if channelStateCache.ChannelId == 0 || channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", channelStateCache.ChannelId, channelStateCache.NodeId)
			channelStateCache.StateOut <- nil
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
			if !exists {
				channelStateCache.StateOut <- nil
				break
			}
			settings, exists := settingsByChannel[channelId(channelStateCache.ChannelId)]
			if !exists {
				channelStateCache.StateOut <- nil
				break
			}
			channelStateCache.StateOut <- &settings
			break
		}
		channelStateCache.StateOut <- nil
	case readAllChannelStates:
		if channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", channelStateCache.NodeId)
			channelStateCache.StatesOut <- nil
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
			if !exists {
				channelStateCache.StatesOut <- nil
				break
			}
			var channelStates []ChannelStateSettingsCache
			for _, channelState := range settingsByChannel {
				channelStates = append(channelStates, channelState)
			}
			channelStateCache.StatesOut <- channelStates
			break
		}
		channelStateCache.StatesOut <- nil
	case readAllChannelStateChannelIds:
		if channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", channelStateCache.NodeId)
			channelStateCache.ChannelIdsOut <- nil
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
			if !exists {
				channelStateCache.ChannelIdsOut <- nil
				break
			}
			var channelIds []int
			for _, channelState := range settingsByChannel {
				channelIds = append(channelIds, channelState.ChannelId)
			}
			channelStateCache.ChannelIdsOut <- channelIds
			break
		}
		channelStateCache.ChannelIdsOut <- nil
	case readSharedChannelStateChannelIds:
		if channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", channelStateCache.NodeId)
			channelStateCache.ChannelIdsOut <- nil
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
			if !exists {
				channelStateCache.ChannelIdsOut <- nil
				break
			}
			var channelIds []int
			for _, channelState := range settingsByChannel {
				if slices.Contains(GetAllTorqNodeIds(), channelState.RemoteNodeId) {
					channelIds = append(channelIds, channelState.ChannelId)
				}
			}
			channelStateCache.ChannelIdsOut <- channelIds
			break
		}
		channelStateCache.ChannelIdsOut <- nil
	case readChannelBalanceState:
		if channelStateCache.ChannelId == 0 || channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", channelStateCache.ChannelId, channelStateCache.NodeId)
			channelStateCache.BalanceStateOut <- nil
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
			if !exists {
				channelStateCache.BalanceStateOut <- nil
				break
			}
			settings, exists := settingsByChannel[channelId(channelStateCache.ChannelId)]
			if !exists {
				channelStateCache.BalanceStateOut <- nil
				break
			}
			capacity := GetChannelSettingByChannelId(channelStateCache.ChannelId).Capacity
			channelBalanceState := processHtlcInclude(channelStateCache, settings, capacity)
			channelStateCache.BalanceStateOut <- &channelBalanceState
			break
		}
		channelStateCache.BalanceStateOut <- nil
	case readAllChannelBalanceStates:
		if channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", channelStateCache.NodeId)
			channelStateCache.BalanceStatesOut <- nil
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
			if !exists {
				channelStateCache.BalanceStatesOut <- nil
				break
			}
			var channelBalanceStates []ChannelBalanceStateSettingsCache
			for _, channelSetting := range GetChannelSettingsByNodeId(channelStateCache.NodeId) {
				if channelSetting.Status != core.Open {
					continue
				}
				settings, exists := settingsByChannel[channelId(channelSetting.ChannelId)]
				if !exists {
					log.Error().Msgf("Channel from channel cache that doesn't exist in channelState cache.")
					continue
				}
				if settings.LocalDisabled && channelStateCache.StateInclude != allChannels {
					continue
				}
				if settings.RemoteDisabled && channelStateCache.StateInclude == allLocalAndRemoteActiveChannels {
					continue
				}
				capacity := channelSetting.Capacity
				channelBalanceStates = append(channelBalanceStates, processHtlcInclude(channelStateCache, settings, capacity))
			}
			channelStateCache.BalanceStatesOut <- channelBalanceStates
			break
		}
		channelStateCache.BalanceStatesOut <- nil
	case writeInitialChannelStates:
		if channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", channelStateCache.NodeId)
			break
		}
		settingsByChannel := make(map[channelId]ChannelStateSettingsCache)
		eventTime := time.Now()
		aggregateChannels := make(map[nodeId]int)
		aggregateCapacity := make(map[nodeId]int64)
		aggregateLocalBalance := make(map[nodeId]int64)
		for _, channelStateSetting := range channelStateCache.ChannelStateSettings {
			channelSettings := GetChannelSettingByChannelId(channelStateSetting.ChannelId)
			capacity := channelSettings.Capacity
			_, aggregateExists := aggregateChannels[nodeId(channelStateSetting.RemoteNodeId)]
			if !aggregateExists {
				channelCountAggregate, capacityAggregate, localBalanceAggregate := getAggregate(channelStateCache.ChannelStateSettings, channelStateSetting.RemoteNodeId)
				aggregateChannels[nodeId(channelStateSetting.RemoteNodeId)] = channelCountAggregate
				aggregateCapacity[nodeId(channelStateSetting.RemoteNodeId)] = capacityAggregate
				aggregateLocalBalance[nodeId(channelStateSetting.RemoteNodeId)] = localBalanceAggregate
			}
			channelStateSetting.PeerChannelCount = aggregateChannels[nodeId(channelStateSetting.RemoteNodeId)]
			channelStateSetting.PeerChannelCapacity = aggregateCapacity[nodeId(channelStateSetting.RemoteNodeId)]
			channelStateSetting.PeerLocalBalance = aggregateLocalBalance[nodeId(channelStateSetting.RemoteNodeId)]
			settingsByChannel[channelId(channelStateSetting.ChannelId)] = channelStateSetting
			channelBalanceEvent := core.ChannelBalanceEvent{
				EventData: core.EventData{
					EventTime: eventTime,
					NodeId:    channelStateCache.NodeId,
				},
				BalanceDelta:         0,
				BalanceDeltaAbsolute: 0,
				ChannelBalanceEventData: core.ChannelBalanceEventData{
					Capacity:                      capacity,
					LocalBalance:                  channelStateSetting.LocalBalance,
					RemoteBalance:                 channelStateSetting.RemoteBalance,
					LocalBalancePerMilleRatio:     int(channelStateSetting.LocalBalance / capacity * 1000),
					PeerChannelCapacity:           aggregateCapacity[nodeId(channelStateSetting.RemoteNodeId)],
					PeerChannelCount:              aggregateChannels[nodeId(channelStateSetting.RemoteNodeId)],
					PeerLocalBalance:              aggregateLocalBalance[nodeId(channelStateSetting.RemoteNodeId)],
					PeerLocalBalancePerMilleRatio: int(aggregateLocalBalance[nodeId(channelStateSetting.RemoteNodeId)] / aggregateCapacity[nodeId(channelStateSetting.RemoteNodeId)] * 1000),
				},
				ChannelId: channelStateSetting.ChannelId,
			}
			existingChannelStateSettings, existingChannelSettingsExists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
			if existingChannelSettingsExists {
				existingState, existingStateExists := existingChannelStateSettings[channelId(channelStateSetting.ChannelId)]
				if existingStateExists {
					channelBalanceEvent.PreviousEventData = &core.ChannelBalanceEventData{
						Capacity:                      capacity,
						LocalBalance:                  existingState.LocalBalance,
						RemoteBalance:                 existingState.RemoteBalance,
						LocalBalancePerMilleRatio:     int(existingState.LocalBalance / capacity * 1000),
						PeerChannelCapacity:           aggregateCapacity[nodeId(channelStateSetting.RemoteNodeId)],
						PeerChannelCount:              aggregateChannels[nodeId(channelStateSetting.RemoteNodeId)],
						PeerLocalBalance:              existingState.PeerLocalBalance,
						PeerLocalBalancePerMilleRatio: int(existingState.PeerLocalBalance / existingState.PeerChannelCapacity * 1000),
					}
					channelBalanceEvent.BalanceDelta = channelBalanceEvent.PreviousEventData.LocalBalance - channelBalanceEvent.LocalBalance
					channelBalanceEvent.BalanceDeltaAbsolute = channelBalanceEvent.BalanceDelta
					if channelBalanceEvent.BalanceDeltaAbsolute < 0 {
						channelBalanceEvent.BalanceDeltaAbsolute = -1 * channelBalanceEvent.BalanceDeltaAbsolute
					}
					if channelBalanceEvent.BalanceDelta != 0 {
						ChannelBalanceChanges <- channelBalanceEvent
					}
				}
			}
		}
		channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] = settingsByChannel
	case writeInitialChannelState:
		if channelStateCache.ChannelId == 0 || channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", channelStateCache.ChannelId, channelStateCache.NodeId)
			break
		}
		channelStateSetting := channelStateCache.ChannelStateSetting
		var peerChannelCount int
		var peerChannelCapacity int64
		var peerLocalBalance int64
		for _, channelState := range channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] {
			if channelState.RemoteNodeId == channelStateSetting.RemoteNodeId {
				peerChannelCount++
				peerChannelCapacity += GetChannelSettingByChannelId(channelState.ChannelId).Capacity
				peerLocalBalance += channelState.LocalBalance
			}
		}
		channelStateSetting.PeerChannelCount = peerChannelCount + 1
		channelStateSetting.PeerChannelCapacity = peerChannelCapacity + GetChannelSettingByChannelId(channelStateSetting.ChannelId).Capacity
		channelStateSetting.PeerLocalBalance = peerLocalBalance + channelStateSetting.LocalBalance
		_, exists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
		if !exists {
			channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] = make(map[channelId]ChannelStateSettingsCache)
		}
		for _, existingChannelStateSetting := range channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] {
			if existingChannelStateSetting.RemoteNodeId == channelStateSetting.RemoteNodeId {
				existingChannelStateSetting.PeerChannelCapacity = channelStateSetting.PeerChannelCapacity
				existingChannelStateSetting.PeerChannelCount = channelStateSetting.PeerChannelCount
				existingChannelStateSetting.PeerLocalBalance = channelStateSetting.PeerLocalBalance
			}
			channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)][channelId(existingChannelStateSetting.ChannelId)] = existingChannelStateSetting
		}

		channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)][channelId(channelStateCache.ChannelId)] = channelStateSetting
	case writeChannelStateNodeStatus:
		if channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", channelStateCache.NodeId)
			break
		}
		currentStatus, exists := channelStateSettingsStatusCache[nodeId(channelStateCache.NodeId)]
		if exists {
			if channelStateCache.Status != currentStatus {
				channelStateSettingsStatusCache[nodeId(channelStateCache.NodeId)] = channelStateCache.Status
			}
			if channelStateCache.Status != core.Active && currentStatus == core.Active {
				channelStateSettingsDeactivationTimeCache[nodeId(channelStateCache.NodeId)] = time.Now()
			}
		} else {
			channelStateSettingsStatusCache[nodeId(channelStateCache.NodeId)] = channelStateCache.Status
			channelStateSettingsDeactivationTimeCache[nodeId(channelStateCache.NodeId)] = time.Now()
		}
	case writeChannelStateChannelStatus:
		if channelStateCache.ChannelId == 0 || channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", channelStateCache.ChannelId, channelStateCache.NodeId)
			break
		}
		if !isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
		if nodeExists {
			channelSetting, channelExists := nodeChannels[channelId(channelStateCache.ChannelId)]
			if channelExists {
				switch channelStateCache.Status {
				case core.Active:
					channelSetting.LocalDisabled = false
					nodeChannels[channelId(channelStateCache.ChannelId)] = channelSetting
				case core.Inactive:
					channelSetting.LocalDisabled = true
					nodeChannels[channelId(channelStateCache.ChannelId)] = channelSetting
				case core.Deleted:
					delete(nodeChannels, channelId(channelStateCache.ChannelId))
				}
				channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] = nodeChannels
			} else {
				channelSettings := GetChannelSettingByChannelId(channelStateCache.ChannelId)
				if channelSettings.Status == core.Open {
					log.Error().Msgf("Received channel event for uncached channel with channelId: %v", channelStateCache.ChannelId)
				}
			}
		} else {
			log.Error().Msgf("Received channel event for uncached node with nodeId: %v", channelStateCache.NodeId)
		}
	case writeChannelStateRoutingPolicy:
		if !isNodeReady(channelStateSettingsStatusCache, channelStateCache.NodeId,
			channelStateSettingsDeactivationTimeCache, channelStateCache.ForceResponse) {
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
		if nodeExists {
			channelSetting, channelExists := nodeChannels[channelId(channelStateCache.ChannelId)]
			if channelExists {
				if channelStateCache.Local {
					channelSetting.LocalDisabled = channelStateCache.Disabled
					channelSetting.LocalTimeLockDelta = channelStateCache.TimeLockDelta
					channelSetting.LocalMinHtlcMsat = channelStateCache.MinHtlcMsat
					channelSetting.LocalMaxHtlcMsat = channelStateCache.MaxHtlcMsat
					channelSetting.LocalFeeBaseMsat = channelStateCache.FeeBaseMsat
					channelSetting.LocalFeeRateMilliMsat = channelStateCache.FeeRateMilliMsat
					nodeChannels[channelId(channelStateCache.ChannelId)] = channelSetting
				} else {
					channelSetting.RemoteDisabled = channelStateCache.Disabled
					channelSetting.RemoteTimeLockDelta = channelStateCache.TimeLockDelta
					channelSetting.RemoteMinHtlcMsat = channelStateCache.MinHtlcMsat
					channelSetting.RemoteMaxHtlcMsat = channelStateCache.MaxHtlcMsat
					channelSetting.RemoteFeeBaseMsat = channelStateCache.FeeBaseMsat
					channelSetting.RemoteFeeRateMilliMsat = channelStateCache.FeeRateMilliMsat
					nodeChannels[channelId(channelStateCache.ChannelId)] = channelSetting
				}
				channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] = nodeChannels
			} else {
				channelSettings := GetChannelSettingByChannelId(channelStateCache.ChannelId)
				if channelSettings.Status == core.Open {
					log.Error().Msgf("Received channel graph event for uncached channel with channelId: %v", channelStateCache.ChannelId)
				}
			}
		} else {
			log.Error().Msgf("Received channel graph event for uncached node with nodeId: %v", channelStateCache.NodeId)
		}
	case writeChannelStateUpdateBalance:
		if channelStateCache.ChannelId == 0 || channelStateCache.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", channelStateCache.ChannelId, channelStateCache.NodeId)
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)]
		if nodeExists {
			channelStateSetting, channelExists := nodeChannels[channelId(channelStateCache.ChannelId)]
			if channelExists {
				eventTime := time.Now()
				channelSettings := GetChannelSettingByChannelId(channelStateSetting.ChannelId)
				channelBalanceEvent := core.ChannelBalanceEvent{
					EventData: core.EventData{
						EventTime: eventTime,
						NodeId:    channelStateCache.NodeId,
					},
					ChannelId:                channelStateSetting.ChannelId,
					BalanceDelta:             channelStateCache.Amount,
					BalanceDeltaAbsolute:     channelStateCache.Amount,
					BalanceUpdateEventOrigin: channelStateCache.BalanceUpdateEventOrigin,
					PreviousEventData: &core.ChannelBalanceEventData{
						Capacity:                      channelSettings.Capacity,
						LocalBalance:                  channelStateSetting.LocalBalance,
						RemoteBalance:                 channelStateSetting.RemoteBalance,
						LocalBalancePerMilleRatio:     int(channelStateSetting.LocalBalance / channelSettings.Capacity * 1000),
						PeerChannelCapacity:           channelStateSetting.PeerChannelCapacity,
						PeerChannelCount:              channelStateSetting.PeerChannelCount,
						PeerLocalBalance:              channelStateSetting.PeerLocalBalance,
						PeerLocalBalancePerMilleRatio: int(channelStateSetting.PeerLocalBalance / channelStateSetting.PeerChannelCapacity * 1000),
					},
				}
				if channelBalanceEvent.BalanceDeltaAbsolute < 0 {
					channelBalanceEvent.BalanceDeltaAbsolute = -1 * channelBalanceEvent.BalanceDeltaAbsolute
				}
				channelStateSetting.NumUpdates = channelStateSetting.NumUpdates + 1
				channelStateSetting.LocalBalance = channelStateSetting.LocalBalance + channelStateCache.Amount
				channelStateSetting.RemoteBalance = channelStateSetting.LocalBalance - channelStateCache.Amount
				nodeChannels[channelId(channelStateCache.ChannelId)] = channelStateSetting
				for _, nc := range nodeChannels {
					if nc.RemoteNodeId == channelStateSetting.RemoteNodeId {
						nc.PeerLocalBalance = nc.PeerLocalBalance + channelStateCache.Amount
					}
				}
				channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] = nodeChannels
				channelBalanceEvent.ChannelBalanceEventData = core.ChannelBalanceEventData{
					Capacity:                      channelSettings.Capacity,
					LocalBalance:                  channelStateSetting.LocalBalance,
					RemoteBalance:                 channelStateSetting.RemoteBalance,
					LocalBalancePerMilleRatio:     int(channelStateSetting.LocalBalance / channelSettings.Capacity * 1000),
					PeerChannelCapacity:           channelStateSetting.PeerChannelCapacity,
					PeerChannelCount:              channelStateSetting.PeerChannelCount,
					PeerLocalBalance:              channelStateSetting.PeerLocalBalance + channelStateCache.Amount,
					PeerLocalBalancePerMilleRatio: int(channelStateSetting.PeerLocalBalance / channelStateSetting.PeerChannelCapacity * 1000),
				}
				ChannelBalanceChanges <- channelBalanceEvent
			} else {
				channelSettings := GetChannelSettingByChannelId(channelStateCache.ChannelId)
				if channelSettings.Status == core.Open {
					log.Error().Msgf("Received channel balance update for uncached channel with channelId: %v", channelStateCache.ChannelId)
				}
			}
		} else {
			log.Error().Msgf("Received channel balance update for uncached node with nodeId: %v", channelStateCache.NodeId)
		}
	case writeChannelStateUpdateHtlcEvent:
		if (channelStateCache.HtlcEvent.OutgoingChannelId == nil || *channelStateCache.HtlcEvent.OutgoingChannelId == 0) &&
			(channelStateCache.HtlcEvent.IncomingChannelId == nil || *channelStateCache.HtlcEvent.IncomingChannelId == 0) ||
			channelStateCache.HtlcEvent.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) nor ( IncomingChannelId (%v) AND OutgoingChannelId (%v) ) allowed",
				channelStateCache.HtlcEvent.NodeId, channelStateCache.HtlcEvent.IncomingChannelId, channelStateCache.HtlcEvent.OutgoingChannelId)
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[nodeId(channelStateCache.HtlcEvent.NodeId)]
		if nodeExists {
			if channelStateCache.HtlcEvent.IncomingChannelId != nil && *channelStateCache.HtlcEvent.IncomingChannelId != 0 {
				channelSetting, channelExists := nodeChannels[channelId(*channelStateCache.HtlcEvent.IncomingChannelId)]
				if channelExists && channelStateCache.HtlcEvent.IncomingAmtMsat != nil && channelStateCache.HtlcEvent.IncomingHtlcId != nil {
					foundIt := false
					var pendingHtlc []Htlc
					for _, htlc := range channelSetting.PendingHtlcs {
						if channelStateCache.HtlcEvent.IncomingHtlcId != nil && htlc.HtlcIndex == *channelStateCache.HtlcEvent.IncomingHtlcId {
							foundIt = true
						} else {
							pendingHtlc = append(pendingHtlc, htlc)
						}
					}
					if foundIt {
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance - int64(*channelStateCache.HtlcEvent.IncomingAmtMsat/1000)
					} else {
						pendingHtlc = append(pendingHtlc, Htlc{
							Incoming:  true,
							Amount:    int64((*channelStateCache.HtlcEvent.IncomingAmtMsat) / 1000),
							HtlcIndex: *channelStateCache.HtlcEvent.IncomingHtlcId,
						})
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance + int64(*channelStateCache.HtlcEvent.IncomingAmtMsat/1000)
					}
					channelSetting.PendingHtlcs = pendingHtlc
					nodeChannels[channelId(*channelStateCache.HtlcEvent.IncomingChannelId)] = channelSetting
					channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] = nodeChannels
				} else {
					if !channelExists {
						channelSettings := GetChannelSettingByChannelId(*channelStateCache.HtlcEvent.IncomingChannelId)
						if channelSettings.Status == core.Open {
							log.Error().Msgf("Received Incoming HTLC channel balance update for uncached channel with channelId: %v", *channelStateCache.HtlcEvent.IncomingChannelId)
						}
					}
				}
			}
			if channelStateCache.HtlcEvent.OutgoingChannelId != nil && *channelStateCache.HtlcEvent.OutgoingChannelId != 0 {
				channelSetting, channelExists := nodeChannels[channelId(*channelStateCache.HtlcEvent.OutgoingChannelId)]
				if channelExists && channelStateCache.HtlcEvent.OutgoingAmtMsat != nil && channelStateCache.HtlcEvent.OutgoingHtlcId != nil {
					foundIt := false
					var pendingHtlc []Htlc
					for _, htlc := range channelSetting.PendingHtlcs {
						if channelStateCache.HtlcEvent.OutgoingHtlcId != nil && htlc.HtlcIndex == *channelStateCache.HtlcEvent.OutgoingHtlcId {
							foundIt = true
						} else {
							pendingHtlc = append(pendingHtlc, htlc)
						}
					}
					if foundIt {
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance + int64(*channelStateCache.HtlcEvent.IncomingAmtMsat/1000)
					} else {
						pendingHtlc = append(pendingHtlc, Htlc{
							Incoming:  false,
							Amount:    int64((*channelStateCache.HtlcEvent.OutgoingAmtMsat) / 1000),
							HtlcIndex: *channelStateCache.HtlcEvent.OutgoingHtlcId,
						})
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance - int64(*channelStateCache.HtlcEvent.IncomingAmtMsat/1000)
					}
					channelSetting.PendingHtlcs = pendingHtlc
					nodeChannels[channelId(*channelStateCache.HtlcEvent.OutgoingChannelId)] = channelSetting
					channelStateSettingsByChannelIdCache[nodeId(channelStateCache.NodeId)] = nodeChannels
				} else {
					if !channelExists {
						channelSettings := GetChannelSettingByChannelId(*channelStateCache.HtlcEvent.OutgoingChannelId)
						if channelSettings.Status == core.Open {
							log.Error().Msgf("Received Outgoing HTLC channel balance update for uncached channel with channelId: %v", *channelStateCache.HtlcEvent.OutgoingChannelId)
						}
					}
				}
			}
		} else {
			log.Error().Msgf("Received HTLC channel balance update for uncached node with nodeId: %v", channelStateCache.HtlcEvent.NodeId)
		}
	case removeChannelStateFromCache:
		for nodeId := range channelStateSettingsByChannelIdCache {
			nodeChannels := channelStateSettingsByChannelIdCache[nodeId]
			delete(nodeChannels, channelId(channelStateCache.ChannelId))
			channelStateSettingsByChannelIdCache[nodeId] = nodeChannels
		}
	case removeChannelStatesFromCache:
		delete(channelStateSettingsDeactivationTimeCache, nodeId(channelStateCache.NodeId))
		delete(channelStateSettingsByChannelIdCache, nodeId(channelStateCache.NodeId))
		delete(channelStateSettingsStatusCache, nodeId(channelStateCache.NodeId))
	}
}

func getAggregate(channelStateSettings []ChannelStateSettingsCache, remoteNodeId int) (int, int64, int64) {
	var channelCountAggregate int
	var capacityAggregate int64
	var localBalanceAggregate int64
	for _, channelStateSettingInner := range channelStateSettings {
		if channelStateSettingInner.RemoteNodeId == remoteNodeId {
			channelCountAggregate++
			capacityAggregate += GetChannelSettingByChannelId(channelStateSettingInner.ChannelId).Capacity
			localBalanceAggregate += channelStateSettingInner.LocalBalance
		}
	}
	return channelCountAggregate, capacityAggregate, localBalanceAggregate
}

func GetChannelStates(nodeId int, forceResponse bool) []ChannelStateSettingsCache {
	channelStatesResponseChannel := make(chan []ChannelStateSettingsCache)
	channelStateCache := ChannelStateCache{
		NodeId:        nodeId,
		ForceResponse: forceResponse,
		Type:          readAllChannelStates,
		StatesOut:     channelStatesResponseChannel,
	}
	ChannelStatesCacheChannel <- channelStateCache
	return <-channelStatesResponseChannel
}

func GetChannelStateChannelIds(nodeId int, forceResponse bool) []int {
	channelIdsResponseChannel := make(chan []int)
	channelStateCache := ChannelStateCache{
		NodeId:        nodeId,
		ForceResponse: forceResponse,
		Type:          readAllChannelStateChannelIds,
		ChannelIdsOut: channelIdsResponseChannel,
	}
	ChannelStatesCacheChannel <- channelStateCache
	return <-channelIdsResponseChannel
}

func GetChannelStateSharedChannelIds(nodeId int, forceResponse bool) []int {
	channelIdsResponseChannel := make(chan []int)
	channelStateCache := ChannelStateCache{
		NodeId:        nodeId,
		ForceResponse: forceResponse,
		Type:          readSharedChannelStateChannelIds,
		ChannelIdsOut: channelIdsResponseChannel,
	}
	ChannelStatesCacheChannel <- channelStateCache
	return <-channelIdsResponseChannel
}

func GetChannelStateNotSharedChannelIds(nodeId int, forceResponse bool) []int {
	var notSharedChannelIds []int
	allChannelIds := GetChannelStateChannelIds(nodeId, forceResponse)
	sharedChannelIds := GetChannelStateSharedChannelIds(nodeId, forceResponse)
	for _, channelId := range allChannelIds {
		if !slices.Contains(sharedChannelIds, channelId) {
			notSharedChannelIds = append(notSharedChannelIds, channelId)
		}
	}
	return notSharedChannelIds
}

func GetChannelState(nodeId, channelId int, forceResponse bool) *ChannelStateSettingsCache {
	channelStateResponseChannel := make(chan *ChannelStateSettingsCache)
	channelStateCache := ChannelStateCache{
		NodeId:        nodeId,
		ChannelId:     channelId,
		ForceResponse: forceResponse,
		Type:          readChannelState,
		StateOut:      channelStateResponseChannel,
	}
	ChannelStatesCacheChannel <- channelStateCache
	return <-channelStateResponseChannel
}

func GetChannelBalanceStates(nodeId int, forceResponse bool, channelStateInclude ChannelStateInclude, htlcInclude ChannelBalanceStateHtlcInclude) []ChannelBalanceStateSettingsCache {
	channelBalanceStateResponseChannel := make(chan []ChannelBalanceStateSettingsCache)
	channelStateCache := ChannelStateCache{
		NodeId:           nodeId,
		ForceResponse:    forceResponse,
		HtlcInclude:      htlcInclude,
		StateInclude:     channelStateInclude,
		Type:             readChannelBalanceState,
		BalanceStatesOut: channelBalanceStateResponseChannel,
	}
	ChannelStatesCacheChannel <- channelStateCache
	return <-channelBalanceStateResponseChannel
}

func GetChannelBalanceState(nodeId, channelId int, forceResponse bool, htlcInclude ChannelBalanceStateHtlcInclude) *ChannelBalanceStateSettingsCache {
	channelBalanceStateResponseChannel := make(chan *ChannelBalanceStateSettingsCache)
	channelStateCache := ChannelStateCache{
		NodeId:          nodeId,
		ChannelId:       channelId,
		ForceResponse:   forceResponse,
		HtlcInclude:     htlcInclude,
		Type:            readChannelBalanceState,
		BalanceStateOut: channelBalanceStateResponseChannel,
	}
	ChannelStatesCacheChannel <- channelStateCache
	return <-channelBalanceStateResponseChannel
}

func SetChannelStates(nodeId int, channelStateSettings []ChannelStateSettingsCache) {
	channelStateCache := ChannelStateCache{
		NodeId:               nodeId,
		ChannelStateSettings: channelStateSettings,
		Type:                 writeInitialChannelStates,
	}
	ChannelStatesCacheChannel <- channelStateCache
}

func SetChannelState(nodeId int, channelStateSetting ChannelStateSettingsCache) {
	channelStateCache := ChannelStateCache{
		NodeId:              nodeId,
		ChannelId:           channelStateSetting.ChannelId,
		ChannelStateSetting: channelStateSetting,
		Type:                writeInitialChannelState,
	}
	ChannelStatesCacheChannel <- channelStateCache
}

func SetChannelStateNodeStatus(nodeId int, status core.Status) {
	channelStateCache := ChannelStateCache{
		NodeId: nodeId,
		Status: status,
		Type:   writeChannelStateNodeStatus,
	}
	ChannelStatesCacheChannel <- channelStateCache
}

func SetChannelStateChannelStatus(nodeId int, channelId int, status core.Status) {
	channelStateCache := ChannelStateCache{
		NodeId:    nodeId,
		ChannelId: channelId,
		Status:    status,
		Type:      writeChannelStateChannelStatus,
	}
	ChannelStatesCacheChannel <- channelStateCache
}

func SetChannelStateRoutingPolicy(nodeId int, channelId int, local bool,
	disabled bool, timeLockDelta uint32, minHtlcMsat uint64, maxHtlcMsat uint64, feeBaseMsat int64, feeRateMilliMsat int64) {
	channelStateCache := ChannelStateCache{
		NodeId:           nodeId,
		ChannelId:        channelId,
		Local:            local,
		Disabled:         disabled,
		TimeLockDelta:    timeLockDelta,
		MinHtlcMsat:      minHtlcMsat,
		MaxHtlcMsat:      maxHtlcMsat,
		FeeBaseMsat:      feeBaseMsat,
		FeeRateMilliMsat: feeRateMilliMsat,
		Type:             writeChannelStateRoutingPolicy,
	}
	ChannelStatesCacheChannel <- channelStateCache
}

func SetChannelStateBalanceUpdateMsat(nodeId int, channelId int, increaseBalance bool, amount uint64,
	eventOrigin core.BalanceUpdateEventOrigin) {

	channelStateCache := ChannelStateCache{
		NodeId:                   nodeId,
		ChannelId:                channelId,
		BalanceUpdateEventOrigin: eventOrigin,
		Type:                     writeChannelStateUpdateBalance,
	}
	if increaseBalance {
		channelStateCache.Amount = int64(amount / 1000)
	} else {
		channelStateCache.Amount = -int64(amount / 1000)
	}
	ChannelStatesCacheChannel <- channelStateCache
}

func SetChannelStateBalanceUpdate(nodeId int, channelId int, increaseBalance bool, amount int64,
	eventOrigin core.BalanceUpdateEventOrigin) {

	channelStateCache := ChannelStateCache{
		NodeId:                   nodeId,
		ChannelId:                channelId,
		BalanceUpdateEventOrigin: eventOrigin,
		Type:                     writeChannelStateUpdateBalance,
	}
	if increaseBalance {
		channelStateCache.Amount = amount
	} else {
		channelStateCache.Amount = -amount
	}
	ChannelStatesCacheChannel <- channelStateCache
}

func SetChannelStateBalanceHtlcEvent(htlcEvent core.HtlcEvent, eventOrigin core.BalanceUpdateEventOrigin) {
	ChannelStatesCacheChannel <- ChannelStateCache{
		BalanceUpdateEventOrigin: eventOrigin,
		HtlcEvent:                htlcEvent,
		Type:                     writeChannelStateUpdateHtlcEvent,
	}
}

func isNodeReady(channelStateSettingsStatusCache map[nodeId]core.Status, nId int,
	channelStateSettingsDeactivationTimeCache map[nodeId]time.Time, forceResponse bool) bool {

	// Channel states not initialized yet
	if channelStateSettingsStatusCache[nodeId(nId)] != core.Active {
		deactivationTime, exists := channelStateSettingsDeactivationTimeCache[nodeId(nId)]
		if exists && time.Since(deactivationTime).Seconds() < toleratedSubscriptionDowntimeSeconds {
			log.Debug().Msgf("Node flagged as active even tough subscription is temporary down for nodeId: %v", nId)
		} else if !forceResponse {
			return false
		}
	}
	return true
}

func processHtlcInclude(channelStateCache ChannelStateCache, settings ChannelStateSettingsCache, capacity int64) ChannelBalanceStateSettingsCache {
	localBalance := settings.LocalBalance
	remoteBalance := settings.RemoteBalance
	if channelStateCache.HtlcInclude == pendingHtlcsLocalBalanceAdjustedDownwards ||
		channelStateCache.HtlcInclude == pendingHtlcsLocalAndRemoteBalanceAdjustedDownwards {
		localBalance = settings.LocalBalance - settings.PendingOutgoingHtlcAmount
	}
	if channelStateCache.HtlcInclude == pendingHtlcsRemoteBalanceAdjustedDownwards ||
		channelStateCache.HtlcInclude == pendingHtlcsLocalAndRemoteBalanceAdjustedDownwards {
		remoteBalance = settings.RemoteBalance - settings.PendingIncomingHtlcAmount
	}
	if channelStateCache.HtlcInclude == pendingHtlcsLocalBalanceAdjustedUpwards ||
		channelStateCache.HtlcInclude == pendingHtlcsLocalAndRemoteBalanceAdjustedUpwards {
		localBalance = settings.LocalBalance + settings.PendingIncomingHtlcAmount
	}
	if channelStateCache.HtlcInclude == pendingHtlcsRemoteBalanceAdjustedUpwards ||
		channelStateCache.HtlcInclude == pendingHtlcsLocalAndRemoteBalanceAdjustedUpwards {
		remoteBalance = settings.RemoteBalance + settings.PendingOutgoingHtlcAmount
	}
	return ChannelBalanceStateSettingsCache{
		NodeId:                     channelStateCache.NodeId,
		RemoteNodeId:               channelStateCache.RemoteNodeId,
		ChannelId:                  channelStateCache.ChannelId,
		HtlcInclude:                channelStateCache.HtlcInclude,
		LocalBalance:               localBalance,
		LocalBalancePerMilleRatio:  int(settings.LocalBalance / capacity * 1000),
		RemoteBalance:              remoteBalance,
		RemoteBalancePerMilleRatio: int(settings.RemoteBalance / capacity * 1000),
	}
}

func RemoveChannelStatesFromCache(nodeId int) {
	ChannelStatesCacheChannel <- ChannelStateCache{
		Type:   removeChannelStatesFromCache,
		NodeId: nodeId,
	}
}

func RemoveChannelStateFromCache(channelId int) {
	ChannelStatesCacheChannel <- ChannelStateCache{
		Type:      removeChannelStateFromCache,
		ChannelId: channelId,
	}
}
