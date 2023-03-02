package commons

import (
	"context"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
)

const toleratedSubscriptionDowntimeSeconds = 15

var ManagedChannelStateChannel = make(chan ManagedChannelState) //nolint:gochecknoglobals

type ManagedChannelStateCacheOperationType uint

const (
	// READ_CHANNELSTATE please provide NodeId, ChannelId and StateOut
	READ_CHANNELSTATE ManagedChannelStateCacheOperationType = iota
	// READ_ALL_CHANNELSTATES please provide NodeId and StatesOut
	READ_ALL_CHANNELSTATES
	// READ_ALL_CHANNELSTATE_CHANNELIDS please provide NodeId and ChannelIdsOut
	READ_ALL_CHANNELSTATE_CHANNELIDS
	// READ_CHANNELBALANCESTATE please provide NodeId, ChannelId, HtlcInclude and BalanceStateOut
	READ_CHANNELBALANCESTATE
	// READ_ALL_CHANNELBALANCESTATES please provide NodeId, StateInclude, HtlcInclude and BalanceStatesOut
	READ_ALL_CHANNELBALANCESTATES
	// WRITE_INITIAL_CHANNELSTATES This requires the lock being active for writing! Please provide the complete information set
	WRITE_INITIAL_CHANNELSTATES
	WRITE_CHANNELSTATE_NODESTATUS
	WRITE_CHANNELSTATE_CHANNELSTATUS
	WRITE_CHANNELSTATE_ROUTINGPOLICY
	WRITE_CHANNELSTATE_UPDATEBALANCE
	WRITE_CHANNELSTATE_UPDATEHTLCEVENT
)

type ChannelBalanceStateHtlcInclude uint

const (
	PENDING_HTLCS_IGNORED ChannelBalanceStateHtlcInclude = iota
	// PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_DOWNWARDS:
	//   LocalBalance = ConfirmedLocalBalance - PendingDecreasingForwardHTLCsAmount - PendingPaymentHTLCsAmount
	PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_DOWNWARDS
	// PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_DOWNWARDS:
	//   RemoteBalance = ConfirmedRemoteBalance - PendingIncreasingForwardHTLCsAmount - PendingInvoiceHTLCsAmount
	PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_DOWNWARDS
	// PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_DOWNWARDS:
	//   LocalBalance = ConfirmedLocalBalance - PendingDecreasingForwardHTLCsAmount - PendingPaymentHTLCsAmount
	//   RemoteBalance = ConfirmedRemoteBalance - PendingIncreasingForwardHTLCsAmount - PendingInvoiceHTLCsAmount
	PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_DOWNWARDS
	// PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_UPWARDS:
	//   LocalBalance = ConfirmedLocalBalance + PendingIncreasingForwardHTLCsAmount + PendingInvoiceHTLCsAmount
	PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_UPWARDS
	//   RemoteBalance = ConfirmedRemoteBalance + PendingDecreasingForwardHTLCsAmount + PendingPaymentHTLCsAmount
	PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_UPWARDS
	// PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS:
	//   LocalBalance = ConfirmedLocalBalance + PendingIncreasingForwardHTLCsAmount + PendingInvoiceHTLCsAmount
	//   RemoteBalance = ConfirmedRemoteBalance + PendingDecreasingForwardHTLCsAmount + PendingPaymentHTLCsAmount
	PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS
)

type ChannelStateInclude uint

const (
	ALL_LOCAL_ACTIVE_CHANNELS ChannelStateInclude = iota
	ALL_LOCAL_AND_REMOTE_ACTIVE_CHANNELS
	ALL_CHANNELS
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

type ManagedChannelState struct {
	Type                 ManagedChannelStateCacheOperationType
	NodeId               int
	RemoteNodeId         int
	ChannelId            int
	Status               Status
	Local                bool
	Balance              int64
	Disabled             bool
	FeeBaseMsat          int64
	FeeRateMilliMsat     int64
	MinHtlcMsat          uint64
	MaxHtlcMsat          uint64
	TimeLockDelta        uint32
	Amount               int64
	ForceResponse        bool
	HtlcEvent            HtlcEvent
	ChannelStateSettings []ManagedChannelStateSettings
	HtlcInclude          ChannelBalanceStateHtlcInclude
	ChannelIdsOut        chan []int
	StateInclude         ChannelStateInclude
	StateOut             chan *ManagedChannelStateSettings
	StatesOut            chan []ManagedChannelStateSettings
	BalanceStateOut      chan *ManagedChannelBalanceStateSettings
	BalanceStatesOut     chan []ManagedChannelBalanceStateSettings
}

type ManagedChannelStateSettings struct {
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

type ManagedChannelBalanceStateSettings struct {
	NodeId                     int                            `json:"nodeId"`
	RemoteNodeId               int                            `json:"remoteNodeId"`
	ChannelId                  int                            `json:"channelId"`
	HtlcInclude                ChannelBalanceStateHtlcInclude `json:"htlcInclude"`
	LocalBalance               int64                          `json:"localBalance"`
	LocalBalancePerMilleRatio  int                            `json:"localBalancePerMilleRatio"`
	RemoteBalance              int64                          `json:"remoteBalance"`
	RemoteBalancePerMilleRatio int                            `json:"remoteBalancePerMilleRatio"`
}

// ManagedChannelStateCache parameter Context is for test cases...
func ManagedChannelStateCache(ch chan ManagedChannelState, ctx context.Context, channelBalanceEventChannel chan ChannelBalanceEvent) {
	channelStateSettingsByChannelIdCache := make(map[int]map[int]ManagedChannelStateSettings, 0)
	channelStateSettingsStatusCache := make(map[int]Status, 0)
	channelStateSettingsDeactivationTimeCache := make(map[int]time.Time, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedChannelState := <-ch:
			processManagedChannelStateSettings(managedChannelState,
				channelStateSettingsStatusCache, channelStateSettingsByChannelIdCache,
				channelStateSettingsDeactivationTimeCache, channelBalanceEventChannel)
		}
	}
}

func processManagedChannelStateSettings(managedChannelState ManagedChannelState,
	channelStateSettingsStatusCache map[int]Status,
	channelStateSettingsByChannelIdCache map[int]map[int]ManagedChannelStateSettings,
	channelStateSettingsDeactivationTimeCache map[int]time.Time,
	channelBalanceEventChannel chan ChannelBalanceEvent) {
	switch managedChannelState.Type {
	case READ_CHANNELSTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId,
			channelStateSettingsDeactivationTimeCache, managedChannelState.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
				break
			}
			settings, exists := settingsByChannel[managedChannelState.ChannelId]
			if !exists {
				SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
				break
			}
			SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, &settings)
			break
		}
		SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
	case READ_ALL_CHANNELSTATES:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId,
			channelStateSettingsDeactivationTimeCache, managedChannelState.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, nil)
				break
			}
			var channelStates []ManagedChannelStateSettings
			for _, channelState := range settingsByChannel {
				channelStates = append(channelStates, channelState)
			}
			SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, channelStates)
			break
		}
		SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, nil)
	case READ_ALL_CHANNELSTATE_CHANNELIDS:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			SendToManagedChannelIdsChannel(managedChannelState.ChannelIdsOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId,
			channelStateSettingsDeactivationTimeCache, managedChannelState.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				SendToManagedChannelIdsChannel(managedChannelState.ChannelIdsOut, nil)
				break
			}
			var channelIds []int
			for _, channelState := range settingsByChannel {
				channelIds = append(channelIds, channelState.ChannelId)
			}
			SendToManagedChannelIdsChannel(managedChannelState.ChannelIdsOut, channelIds)
			break
		}
		SendToManagedChannelIdsChannel(managedChannelState.ChannelIdsOut, nil)
	case READ_CHANNELBALANCESTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId,
			channelStateSettingsDeactivationTimeCache, managedChannelState.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
				break
			}
			settings, exists := settingsByChannel[managedChannelState.ChannelId]
			if !exists {
				SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
				break
			}
			capacity := GetChannelSettingByChannelId(managedChannelState.ChannelId).Capacity
			channelBalanceState := processHtlcInclude(managedChannelState, settings, capacity)
			SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, &channelBalanceState)
			break
		}
		SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
	case READ_ALL_CHANNELBALANCESTATES:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId,
			channelStateSettingsDeactivationTimeCache, managedChannelState.ForceResponse) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
				break
			}
			var channelBalanceStates []ManagedChannelBalanceStateSettings
			for _, channelSetting := range GetChannelSettingsByNodeId(managedChannelState.NodeId) {
				if channelSetting.Status != Open {
					continue
				}
				settings, exists := settingsByChannel[channelSetting.ChannelId]
				if !exists {
					log.Error().Msgf("Channel from channel cache that doesn't exist in channelState cache.")
					continue
				}
				if settings.LocalDisabled && managedChannelState.StateInclude != ALL_CHANNELS {
					continue
				}
				if settings.RemoteDisabled && managedChannelState.StateInclude == ALL_LOCAL_AND_REMOTE_ACTIVE_CHANNELS {
					continue
				}
				capacity := channelSetting.Capacity
				channelBalanceStates = append(channelBalanceStates, processHtlcInclude(managedChannelState, settings, capacity))
			}
			SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, channelBalanceStates)
			break
		}
		SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
	case WRITE_INITIAL_CHANNELSTATES:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			break
		}
		settingsByChannel := make(map[int]ManagedChannelStateSettings)
		eventTime := time.Now()
		aggregateChannels := make(map[int]int)
		aggregateCapacity := make(map[int]int64)
		aggregateLocalBalance := make(map[int]int64)
		for _, channelStateSetting := range managedChannelState.ChannelStateSettings {
			channelSettings := GetChannelSettingByChannelId(channelStateSetting.ChannelId)
			capacity := channelSettings.Capacity
			_, aggregateExists := aggregateChannels[channelStateSetting.RemoteNodeId]
			if !aggregateExists {
				channelCountAggregate, capacityAggregate, localBalanceAggregate := getAggregate(managedChannelState.ChannelStateSettings, channelStateSetting.RemoteNodeId)
				aggregateChannels[channelStateSetting.RemoteNodeId] = channelCountAggregate
				aggregateCapacity[channelStateSetting.RemoteNodeId] = capacityAggregate
				aggregateLocalBalance[channelStateSetting.RemoteNodeId] = localBalanceAggregate
			}
			channelStateSetting.PeerChannelCount = aggregateChannels[channelStateSetting.RemoteNodeId]
			channelStateSetting.PeerChannelCapacity = aggregateCapacity[channelStateSetting.RemoteNodeId]
			channelStateSetting.PeerLocalBalance = aggregateLocalBalance[channelStateSetting.RemoteNodeId]
			settingsByChannel[channelStateSetting.ChannelId] = channelStateSetting
			channelBalanceEvent := ChannelBalanceEvent{
				EventData: EventData{
					EventTime: eventTime,
					NodeId:    managedChannelState.NodeId,
				},
				ChannelBalanceEventData: ChannelBalanceEventData{
					Capacity:                      capacity,
					LocalBalance:                  channelStateSetting.LocalBalance,
					RemoteBalance:                 channelStateSetting.RemoteBalance,
					LocalBalancePerMilleRatio:     int(channelStateSetting.LocalBalance / capacity * 1000),
					PeerChannelCapacity:           aggregateCapacity[channelStateSetting.RemoteNodeId],
					PeerChannelCount:              aggregateChannels[channelStateSetting.RemoteNodeId],
					PeerLocalBalance:              aggregateLocalBalance[channelStateSetting.RemoteNodeId],
					PeerLocalBalancePerMilleRatio: int(aggregateLocalBalance[channelStateSetting.RemoteNodeId] / aggregateCapacity[channelStateSetting.RemoteNodeId] * 1000),
				},
				ChannelId: channelStateSetting.ChannelId,
			}
			existingChannelStateSettings, existingChannelSettingsExists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if existingChannelSettingsExists {
				existingState, existingStateExists := existingChannelStateSettings[channelStateSetting.ChannelId]
				if existingStateExists {
					channelBalanceEvent.PreviousEventData = &ChannelBalanceEventData{
						Capacity:                      capacity,
						LocalBalance:                  existingState.LocalBalance,
						RemoteBalance:                 existingState.RemoteBalance,
						LocalBalancePerMilleRatio:     int(existingState.LocalBalance / capacity * 1000),
						PeerChannelCapacity:           aggregateCapacity[channelStateSetting.RemoteNodeId],
						PeerChannelCount:              aggregateChannels[channelStateSetting.RemoteNodeId],
						PeerLocalBalance:              existingState.PeerLocalBalance,
						PeerLocalBalancePerMilleRatio: int(existingState.PeerLocalBalance / existingState.PeerChannelCapacity * 1000),
					}
					if channelBalanceEvent.PreviousEventData != nil &&
						channelBalanceEvent.PreviousEventData.LocalBalance != channelBalanceEvent.LocalBalance {
						channelBalanceEventChannel <- channelBalanceEvent
					}
				}
			}
		}
		channelStateSettingsByChannelIdCache[managedChannelState.NodeId] = settingsByChannel
	case WRITE_CHANNELSTATE_NODESTATUS:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			break
		}
		currentStatus, exists := channelStateSettingsStatusCache[managedChannelState.NodeId]
		if exists {
			if managedChannelState.Status != currentStatus {
				channelStateSettingsStatusCache[managedChannelState.NodeId] = managedChannelState.Status
			}
			if managedChannelState.Status != Active && currentStatus == Active {
				channelStateSettingsDeactivationTimeCache[managedChannelState.NodeId] = time.Now()
			}
		} else {
			channelStateSettingsStatusCache[managedChannelState.NodeId] = managedChannelState.Status
			channelStateSettingsDeactivationTimeCache[managedChannelState.NodeId] = time.Now()
		}
	case WRITE_CHANNELSTATE_CHANNELSTATUS:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			break
		}
		if !isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId,
			channelStateSettingsDeactivationTimeCache, managedChannelState.ForceResponse) {
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
		if nodeExists {
			channelSetting, channelExists := nodeChannels[managedChannelState.ChannelId]
			if channelExists {
				switch managedChannelState.Status {
				case Active:
					channelSetting.LocalDisabled = false
					nodeChannels[managedChannelState.ChannelId] = channelSetting
				case Inactive:
					channelSetting.LocalDisabled = true
					nodeChannels[managedChannelState.ChannelId] = channelSetting
				case Deleted:
					delete(nodeChannels, managedChannelState.ChannelId)
				}
				channelStateSettingsByChannelIdCache[managedChannelState.NodeId] = nodeChannels
			} else {
				managedChannelSettings := GetChannelSettingByChannelId(managedChannelState.ChannelId)
				if managedChannelSettings.Status == Open {
					log.Error().Msgf("Received channel event for uncached channel with channelId: %v", managedChannelState.ChannelId)
				}
			}
		} else {
			log.Error().Msgf("Received channel event for uncached node with nodeId: %v", managedChannelState.NodeId)
		}
	case WRITE_CHANNELSTATE_ROUTINGPOLICY:
		if !isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId,
			channelStateSettingsDeactivationTimeCache, managedChannelState.ForceResponse) {
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
		if nodeExists {
			channelSetting, channelExists := nodeChannels[managedChannelState.ChannelId]
			if channelExists {
				if managedChannelState.Local {
					channelSetting.LocalDisabled = managedChannelState.Disabled
					channelSetting.LocalTimeLockDelta = managedChannelState.TimeLockDelta
					channelSetting.LocalMinHtlcMsat = managedChannelState.MinHtlcMsat
					channelSetting.LocalMaxHtlcMsat = managedChannelState.MaxHtlcMsat
					channelSetting.LocalFeeBaseMsat = managedChannelState.FeeBaseMsat
					channelSetting.LocalFeeRateMilliMsat = managedChannelState.FeeRateMilliMsat
					nodeChannels[managedChannelState.ChannelId] = channelSetting
				} else {
					channelSetting.RemoteDisabled = managedChannelState.Disabled
					channelSetting.RemoteTimeLockDelta = managedChannelState.TimeLockDelta
					channelSetting.RemoteMinHtlcMsat = managedChannelState.MinHtlcMsat
					channelSetting.RemoteMaxHtlcMsat = managedChannelState.MaxHtlcMsat
					channelSetting.RemoteFeeBaseMsat = managedChannelState.FeeBaseMsat
					channelSetting.RemoteFeeRateMilliMsat = managedChannelState.FeeRateMilliMsat
					nodeChannels[managedChannelState.ChannelId] = channelSetting
				}
				channelStateSettingsByChannelIdCache[managedChannelState.NodeId] = nodeChannels
			} else {
				managedChannelSettings := GetChannelSettingByChannelId(managedChannelState.ChannelId)
				if managedChannelSettings.Status == Open {
					log.Error().Msgf("Received channel graph event for uncached channel with channelId: %v", managedChannelState.ChannelId)
				}
			}
		} else {
			log.Error().Msgf("Received channel graph event for uncached node with nodeId: %v", managedChannelState.NodeId)
		}
	case WRITE_CHANNELSTATE_UPDATEBALANCE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
		if nodeExists {
			channelStateSetting, channelExists := nodeChannels[managedChannelState.ChannelId]
			if channelExists {
				eventTime := time.Now()
				channelSettings := GetChannelSettingByChannelId(channelStateSetting.ChannelId)
				channelBalanceEvent := ChannelBalanceEvent{
					EventData: EventData{
						EventTime: eventTime,
						NodeId:    managedChannelState.NodeId,
					},
					ChannelId: channelStateSetting.ChannelId,
					PreviousEventData: &ChannelBalanceEventData{
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
				channelStateSetting.NumUpdates = channelStateSetting.NumUpdates + 1
				channelStateSetting.LocalBalance = channelStateSetting.LocalBalance + managedChannelState.Amount
				channelStateSetting.RemoteBalance = channelStateSetting.LocalBalance - managedChannelState.Amount
				nodeChannels[managedChannelState.ChannelId] = channelStateSetting
				for _, nc := range nodeChannels {
					if nc.RemoteNodeId == channelStateSetting.RemoteNodeId {
						nc.PeerLocalBalance = nc.PeerLocalBalance + managedChannelState.Amount
					}
				}
				channelStateSettingsByChannelIdCache[managedChannelState.NodeId] = nodeChannels
				channelBalanceEvent.ChannelBalanceEventData = ChannelBalanceEventData{
					Capacity:                      channelSettings.Capacity,
					LocalBalance:                  channelStateSetting.LocalBalance,
					RemoteBalance:                 channelStateSetting.RemoteBalance,
					LocalBalancePerMilleRatio:     int(channelStateSetting.LocalBalance / channelSettings.Capacity * 1000),
					PeerChannelCapacity:           channelStateSetting.PeerChannelCapacity,
					PeerChannelCount:              channelStateSetting.PeerChannelCount,
					PeerLocalBalance:              channelStateSetting.PeerLocalBalance + managedChannelState.Amount,
					PeerLocalBalancePerMilleRatio: int(channelStateSetting.PeerLocalBalance / channelStateSetting.PeerChannelCapacity * 1000),
				}
				channelBalanceEventChannel <- channelBalanceEvent
			} else {
				managedChannelSettings := GetChannelSettingByChannelId(managedChannelState.ChannelId)
				if managedChannelSettings.Status == Open {
					log.Error().Msgf("Received channel balance update for uncached channel with channelId: %v", managedChannelState.ChannelId)
				}
			}
		} else {
			log.Error().Msgf("Received channel balance update for uncached node with nodeId: %v", managedChannelState.NodeId)
		}
	case WRITE_CHANNELSTATE_UPDATEHTLCEVENT:
		if (managedChannelState.HtlcEvent.OutgoingChannelId == nil || *managedChannelState.HtlcEvent.OutgoingChannelId == 0) &&
			(managedChannelState.HtlcEvent.IncomingChannelId == nil || *managedChannelState.HtlcEvent.IncomingChannelId == 0) ||
			managedChannelState.HtlcEvent.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) nor ( IncomingChannelId (%v) AND OutgoingChannelId (%v) ) allowed",
				managedChannelState.HtlcEvent.NodeId, managedChannelState.HtlcEvent.IncomingChannelId, managedChannelState.HtlcEvent.OutgoingChannelId)
			break
		}
		nodeChannels, nodeExists := channelStateSettingsByChannelIdCache[managedChannelState.HtlcEvent.NodeId]
		if nodeExists {
			if managedChannelState.HtlcEvent.IncomingChannelId != nil && *managedChannelState.HtlcEvent.IncomingChannelId != 0 {
				channelSetting, channelExists := nodeChannels[*managedChannelState.HtlcEvent.IncomingChannelId]
				if channelExists && managedChannelState.HtlcEvent.IncomingAmtMsat != nil && managedChannelState.HtlcEvent.IncomingHtlcId != nil {
					foundIt := false
					var pendingHtlc []Htlc
					for _, htlc := range channelSetting.PendingHtlcs {
						if managedChannelState.HtlcEvent.IncomingHtlcId != nil && htlc.HtlcIndex == *managedChannelState.HtlcEvent.IncomingHtlcId {
							foundIt = true
						} else {
							pendingHtlc = append(pendingHtlc, htlc)
						}
					}
					if foundIt {
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance - int64(*managedChannelState.HtlcEvent.IncomingAmtMsat/1000)
					} else {
						pendingHtlc = append(pendingHtlc, Htlc{
							Incoming:  true,
							Amount:    int64((*managedChannelState.HtlcEvent.IncomingAmtMsat) / 1000),
							HtlcIndex: *managedChannelState.HtlcEvent.IncomingHtlcId,
						})
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance + int64(*managedChannelState.HtlcEvent.IncomingAmtMsat/1000)
					}
					channelSetting.PendingHtlcs = pendingHtlc
					nodeChannels[*managedChannelState.HtlcEvent.IncomingChannelId] = channelSetting
					channelStateSettingsByChannelIdCache[managedChannelState.NodeId] = nodeChannels
				} else {
					if !channelExists {
						managedChannelSettings := GetChannelSettingByChannelId(*managedChannelState.HtlcEvent.IncomingChannelId)
						if managedChannelSettings.Status == Open {
							log.Error().Msgf("Received Incoming HTLC channel balance update for uncached channel with channelId: %v", *managedChannelState.HtlcEvent.IncomingChannelId)
						}
					}
				}
			}
			if managedChannelState.HtlcEvent.OutgoingChannelId != nil && *managedChannelState.HtlcEvent.OutgoingChannelId != 0 {
				channelSetting, channelExists := nodeChannels[*managedChannelState.HtlcEvent.OutgoingChannelId]
				if channelExists && managedChannelState.HtlcEvent.OutgoingAmtMsat != nil && managedChannelState.HtlcEvent.OutgoingHtlcId != nil {
					foundIt := false
					var pendingHtlc []Htlc
					for _, htlc := range channelSetting.PendingHtlcs {
						if managedChannelState.HtlcEvent.OutgoingHtlcId != nil && htlc.HtlcIndex == *managedChannelState.HtlcEvent.OutgoingHtlcId {
							foundIt = true
						} else {
							pendingHtlc = append(pendingHtlc, htlc)
						}
					}
					if foundIt {
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance + int64(*managedChannelState.HtlcEvent.IncomingAmtMsat/1000)
					} else {
						pendingHtlc = append(pendingHtlc, Htlc{
							Incoming:  false,
							Amount:    int64((*managedChannelState.HtlcEvent.OutgoingAmtMsat) / 1000),
							HtlcIndex: *managedChannelState.HtlcEvent.OutgoingHtlcId,
						})
						channelSetting.UnsettledBalance = channelSetting.UnsettledBalance - int64(*managedChannelState.HtlcEvent.IncomingAmtMsat/1000)
					}
					channelSetting.PendingHtlcs = pendingHtlc
					nodeChannels[*managedChannelState.HtlcEvent.OutgoingChannelId] = channelSetting
					channelStateSettingsByChannelIdCache[managedChannelState.NodeId] = nodeChannels
				} else {
					if !channelExists {
						managedChannelSettings := GetChannelSettingByChannelId(*managedChannelState.HtlcEvent.OutgoingChannelId)
						if managedChannelSettings.Status == Open {
							log.Error().Msgf("Received Outgoing HTLC channel balance update for uncached channel with channelId: %v", *managedChannelState.HtlcEvent.OutgoingChannelId)
						}
					}
				}
			}
		} else {
			log.Error().Msgf("Received HTLC channel balance update for uncached node with nodeId: %v", managedChannelState.HtlcEvent.NodeId)
		}
	}
}

func getAggregate(channelStateSettings []ManagedChannelStateSettings, remoteNodeId int) (int, int64, int64) {
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

func GetChannelStates(nodeId int, forceResponse bool) []ManagedChannelStateSettings {
	channelStatesResponseChannel := make(chan []ManagedChannelStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:        nodeId,
		ForceResponse: forceResponse,
		Type:          READ_ALL_CHANNELSTATES,
		StatesOut:     channelStatesResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelStatesResponseChannel
}

func GetChannelStateChannelIds(nodeId int, forceResponse bool) []int {
	channelIdsResponseChannel := make(chan []int)
	managedChannelState := ManagedChannelState{
		NodeId:        nodeId,
		ForceResponse: forceResponse,
		Type:          READ_ALL_CHANNELSTATE_CHANNELIDS,
		ChannelIdsOut: channelIdsResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelIdsResponseChannel
}

func GetChannelState(nodeId, channelId int, forceResponse bool) *ManagedChannelStateSettings {
	channelStateResponseChannel := make(chan *ManagedChannelStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:        nodeId,
		ChannelId:     channelId,
		ForceResponse: forceResponse,
		Type:          READ_CHANNELSTATE,
		StateOut:      channelStateResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelStateResponseChannel
}

func GetChannelBalanceStates(nodeId int, forceResponse bool, channelStateInclude ChannelStateInclude, htlcInclude ChannelBalanceStateHtlcInclude) []ManagedChannelBalanceStateSettings {
	channelBalanceStateResponseChannel := make(chan []ManagedChannelBalanceStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:           nodeId,
		ForceResponse:    forceResponse,
		HtlcInclude:      htlcInclude,
		StateInclude:     channelStateInclude,
		Type:             READ_CHANNELBALANCESTATE,
		BalanceStatesOut: channelBalanceStateResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelBalanceStateResponseChannel
}

func GetChannelBalanceState(nodeId, channelId int, forceResponse bool, htlcInclude ChannelBalanceStateHtlcInclude) *ManagedChannelBalanceStateSettings {
	channelBalanceStateResponseChannel := make(chan *ManagedChannelBalanceStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:          nodeId,
		ChannelId:       channelId,
		ForceResponse:   forceResponse,
		HtlcInclude:     htlcInclude,
		Type:            READ_CHANNELBALANCESTATE,
		BalanceStateOut: channelBalanceStateResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelBalanceStateResponseChannel
}

func SetChannelStates(nodeId int, channelStateSettings []ManagedChannelStateSettings) {
	managedChannelState := ManagedChannelState{
		NodeId:               nodeId,
		ChannelStateSettings: channelStateSettings,
		Type:                 WRITE_INITIAL_CHANNELSTATES,
	}
	ManagedChannelStateChannel <- managedChannelState
}

func SetChannelStateNodeStatus(nodeId int, status Status) {
	managedChannelState := ManagedChannelState{
		NodeId: nodeId,
		Status: status,
		Type:   WRITE_CHANNELSTATE_NODESTATUS,
	}
	ManagedChannelStateChannel <- managedChannelState
}

func SetChannelStateChannelStatus(nodeId int, channelId int, status Status) {
	managedChannelState := ManagedChannelState{
		NodeId:    nodeId,
		ChannelId: channelId,
		Status:    status,
		Type:      WRITE_CHANNELSTATE_CHANNELSTATUS,
	}
	ManagedChannelStateChannel <- managedChannelState
}

func SetChannelStateRoutingPolicy(nodeId int, channelId int, local bool,
	disabled bool, timeLockDelta uint32, minHtlcMsat uint64, maxHtlcMsat uint64, feeBaseMsat int64, feeRateMilliMsat int64) {
	managedChannelState := ManagedChannelState{
		NodeId:           nodeId,
		ChannelId:        channelId,
		Local:            local,
		Disabled:         disabled,
		TimeLockDelta:    timeLockDelta,
		MinHtlcMsat:      minHtlcMsat,
		MaxHtlcMsat:      maxHtlcMsat,
		FeeBaseMsat:      feeBaseMsat,
		FeeRateMilliMsat: feeRateMilliMsat,
		Type:             WRITE_CHANNELSTATE_ROUTINGPOLICY,
	}
	ManagedChannelStateChannel <- managedChannelState
}

func SetChannelStateBalanceUpdateMsat(nodeId int, channelId int, increaseBalance bool, amount uint64) {
	managedChannelState := ManagedChannelState{
		NodeId:    nodeId,
		ChannelId: channelId,
		Type:      WRITE_CHANNELSTATE_UPDATEBALANCE,
	}
	if increaseBalance {
		managedChannelState.Amount = int64(amount / 1000)
	} else {
		managedChannelState.Amount = -int64(amount / 1000)
	}
	ManagedChannelStateChannel <- managedChannelState
}

func SetChannelStateBalanceUpdate(nodeId int, channelId int, increaseBalance bool, amount int64) {
	managedChannelState := ManagedChannelState{
		NodeId:    nodeId,
		ChannelId: channelId,
		Type:      WRITE_CHANNELSTATE_UPDATEBALANCE,
	}
	if increaseBalance {
		managedChannelState.Amount = amount
	} else {
		managedChannelState.Amount = -amount
	}
	ManagedChannelStateChannel <- managedChannelState
}

func SetChannelStateBalanceHtlcEvent(htlcEvent HtlcEvent) {
	ManagedChannelStateChannel <- ManagedChannelState{
		HtlcEvent: htlcEvent,
		Type:      WRITE_CHANNELSTATE_UPDATEHTLCEVENT,
	}
}

func isNodeReady(channelStateSettingsStatusCache map[int]Status, nodeId int,
	channelStateSettingsDeactivationTimeCache map[int]time.Time, forceResponse bool) bool {

	// Channel states not initialized yet
	if channelStateSettingsStatusCache[nodeId] != Active {
		deactivationTime, exists := channelStateSettingsDeactivationTimeCache[nodeId]
		if exists && time.Since(deactivationTime).Seconds() < toleratedSubscriptionDowntimeSeconds {
			log.Debug().Msgf("Node flagged as active even tough subscription is temporary down for nodeId: %v", nodeId)
		} else if !forceResponse {
			return false
		}
	}
	return true
}

func processHtlcInclude(managedChannelState ManagedChannelState, settings ManagedChannelStateSettings, capacity int64) ManagedChannelBalanceStateSettings {
	localBalance := settings.LocalBalance
	remoteBalance := settings.RemoteBalance
	if managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_DOWNWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_DOWNWARDS {
		localBalance = settings.LocalBalance - settings.PendingOutgoingHtlcAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_DOWNWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_DOWNWARDS {
		remoteBalance = settings.RemoteBalance - settings.PendingIncomingHtlcAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_UPWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS {
		localBalance = settings.LocalBalance + settings.PendingIncomingHtlcAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_UPWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS {
		remoteBalance = settings.RemoteBalance + settings.PendingOutgoingHtlcAmount
	}
	return ManagedChannelBalanceStateSettings{
		NodeId:                     managedChannelState.NodeId,
		RemoteNodeId:               managedChannelState.RemoteNodeId,
		ChannelId:                  managedChannelState.ChannelId,
		HtlcInclude:                managedChannelState.HtlcInclude,
		LocalBalance:               localBalance,
		LocalBalancePerMilleRatio:  int(settings.LocalBalance / capacity * 1000),
		RemoteBalance:              remoteBalance,
		RemoteBalancePerMilleRatio: int(settings.RemoteBalance / capacity * 1000),
	}
}
