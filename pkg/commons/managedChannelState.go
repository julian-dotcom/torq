package commons

import (
	"context"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/broadcast"
)

var ManagedChannelStateChannel = make(chan ManagedChannelState) //nolint:gochecknoglobals

type ManagedChannelStateCacheOperationType uint

const (
	// READ_CHANNELSTATE please provide NodeId, ChannelId and StateOut
	READ_CHANNELSTATE ManagedChannelStateCacheOperationType = iota
	// READ_ALL_CHANNELSTATES please provide NodeId and StatesOut
	READ_ALL_CHANNELSTATES
	// READ_CHANNELBALANCESTATE please provide NodeId, ChannelId, HtlcInclude and BalanceStateOut
	READ_CHANNELBALANCESTATE
	// READ_ALL_CHANNELBALANCESTATES please provide NodeId, StateInclude, HtlcInclude and BalanceStatesOut
	READ_ALL_CHANNELBALANCESTATES
	// WRITE_INITIAL_CHANNELSTATE This requires the lock being active for writing! Please provide the complete information set
	WRITE_INITIAL_CHANNELSTATE
	// READ_CHANNELSTATELOCK please provide NodeId and LockOut
	READ_CHANNELSTATELOCK
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
	ChannelStateSettings ManagedChannelStateSettings
	HtlcInclude          ChannelBalanceStateHtlcInclude
	StateInclude         ChannelStateInclude
	StateOut             chan *ManagedChannelStateSettings
	StatesOut            chan []ManagedChannelStateSettings
	BalanceStateOut      chan *ManagedChannelBalanceStateSettings
	BalanceStatesOut     chan []ManagedChannelBalanceStateSettings
	LockOut              chan *sync.RWMutex
}

type ManagedChannelStateSettings struct {
	NodeId       int `json:"nodeId"`
	RemoteNodeId int `json:"remoteNodeId"`
	ChannelId    int `json:"channelId"`

	LocalBalance          int64  `json:"localBalance"`
	LocalDisabled         bool   `json:"localDisabled"`
	LocalFeeBaseMsat      int64  `json:"localFeeBaseMsat"`
	LocalFeeRateMilliMsat int64  `json:"localFeeRateMilliMsat"`
	LocalMinHtlc          int64  `json:"localMinHtlc"`
	LocalMaxHtlcMsat      uint64 `json:"localMaxHtlcMsat"`
	LocalTimeLockDelta    uint32 `json:"localTimeLockDelta"`

	RemoteBalance          int64  `json:"remoteBalance"`
	RemoteDisabled         bool   `json:"remoteDisabled"`
	RemoteFeeBaseMsat      int64  `json:"remoteFeeBaseMsat"`
	RemoteFeeRateMilliMsat int64  `json:"remoteFeeRateMilliMsat"`
	RemoteMinHtlc          int64  `json:"remoteMinHtlc"`
	RemoteMaxHtlcMsat      uint64 `json:"remoteMaxHtlcMsat"`
	RemoteTimeLockDelta    uint32 `json:"remoteTimeLockDelta"`

	UnsettledBalance int64 `json:"unsettledBalance"`

	PendingHtlcs []Htlc `json:"pendingHtlcs"`
	// INCREASING LOCAL BALANCE HTLCs
	PendingIncreasingHtlcCount  int   `json:"pendingIncreasingHtlcCount"`
	PendingIncreasingHtlcAmount int64 `json:"pendingIncreasingHtlcAmount"`
	// DECREASING LOCAL BALANCE HTLCs
	PendingDecreasingHtlcCount  int   `json:"pendingDecreasingHtlcCount"`
	PendingDecreasingHtlcAmount int64 `json:"pendingDecreasingHtlcAmount"`

	// STALE INFORMATION ONLY OBTAINED VIA LND REGULAR CHECKINS SO NOT MAINTAINED
	CommitFee             int64                `json:"commitFee"`
	CommitWeight          int64                `json:"commitWeight"`
	FeePerKw              int64                `json:"feePerKw"`
	TotalSatoshisSent     int64                `json:"totalSatoshisSent"`
	NumUpdates            uint64               `json:"numUpdates"`
	ChanStatusFlags       string               `json:"chanStatusFlags"`
	CommitmentType        lnrpc.CommitmentType `json:"commitmentType"`
	Lifetime              int64                `json:"lifetime"`
	TotalSatoshisReceived int64                `json:"totalSatoshisReceived"`
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
func ManagedChannelStateCache(ch chan ManagedChannelState, broadcaster broadcast.BroadcastServer, ctx context.Context) {
	channelStateSettingsByChannelIdCache := make(map[int]map[int]ManagedChannelStateSettings, 0)
	channelStateSettingsStatusCache := make(map[int]Status, 0)
	channelStateSettingsLockCache := make(map[int]*sync.RWMutex, 0)
	channelStateSettingsDeactivationTimeCache := make(map[int]time.Time, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedChannelState := <-ch:
			processManagedChannelStateSettings(managedChannelState,
				channelStateSettingsLockCache, channelStateSettingsStatusCache, channelStateSettingsByChannelIdCache,
				channelStateSettingsDeactivationTimeCache)
		case event := <-broadcaster.Subscribe():
			processBroadcastedEvent(event,
				channelStateSettingsLockCache, channelStateSettingsStatusCache, channelStateSettingsByChannelIdCache,
				channelStateSettingsDeactivationTimeCache)
		}
	}
}
func processBroadcastedEvent(event interface{},
	channelStateSettingsLockCache map[int]*sync.RWMutex,
	channelStateSettingsStatusCache map[int]Status,
	channelStateSettingsByChannelIdCache map[int]map[int]ManagedChannelStateSettings,
	channelStateSettingsDeactivationTimeCache map[int]time.Time) {
	var nodeChannels map[int]ManagedChannelStateSettings
	var channelSetting ManagedChannelStateSettings
	var exists bool

	if serviceEvent, ok := event.(ServiceEvent); ok {
		if serviceEvent.NodeId == 0 || serviceEvent.Type != LndService {
			return
		}
		currentStatus, exists := channelStateSettingsStatusCache[serviceEvent.NodeId]
		if exists {
			if serviceEvent.Status != currentStatus {
				channelStateSettingsStatusCache[serviceEvent.NodeId] = serviceEvent.Status
			}
		} else {
			channelStateSettingsStatusCache[serviceEvent.NodeId] = serviceEvent.Status
		}
		if serviceEvent.Status != Active && serviceEvent.PreviousStatus == Active {
			channelStateSettingsDeactivationTimeCache[serviceEvent.NodeId] = serviceEvent.EventTime
		}
	} else if channelGraphEvent, ok := event.(ChannelGraphEvent); ok {
		if channelGraphEvent.NodeId == 0 || channelGraphEvent.ChannelId == nil || *channelGraphEvent.ChannelId == 0 ||
			channelGraphEvent.AnnouncingNodeId == nil || *channelGraphEvent.AnnouncingNodeId == 0 ||
			channelGraphEvent.ConnectingNodeId == nil || *channelGraphEvent.ConnectingNodeId == 0 {
			return
		}
		if !isNodeReady(channelStateSettingsStatusCache, channelGraphEvent.NodeId, channelStateSettingsLockCache, channelStateSettingsDeactivationTimeCache) {
			return
		}
		defer channelStateSettingsLockCache[channelGraphEvent.NodeId].RUnlock()
		nodeChannels, exists = channelStateSettingsByChannelIdCache[channelGraphEvent.NodeId]
		if exists {
			channelSetting, exists = nodeChannels[*channelGraphEvent.ChannelId]
			if exists {
				if *channelGraphEvent.AnnouncingNodeId == channelGraphEvent.NodeId {
					channelSetting.LocalDisabled = channelGraphEvent.Disabled
					channelSetting.LocalTimeLockDelta = channelGraphEvent.TimeLockDelta
					channelSetting.LocalMinHtlc = channelGraphEvent.MinHtlc
					channelSetting.LocalMaxHtlcMsat = channelGraphEvent.MaxHtlcMsat
					channelSetting.LocalFeeBaseMsat = channelGraphEvent.FeeBaseMsat
					channelSetting.LocalFeeRateMilliMsat = channelGraphEvent.FeeRateMilliMsat
				}
				if *channelGraphEvent.ConnectingNodeId == channelGraphEvent.NodeId {
					channelSetting.RemoteDisabled = channelGraphEvent.Disabled
					channelSetting.RemoteTimeLockDelta = channelGraphEvent.TimeLockDelta
					channelSetting.RemoteMinHtlc = channelGraphEvent.MinHtlc
					channelSetting.RemoteMaxHtlcMsat = channelGraphEvent.MaxHtlcMsat
					channelSetting.RemoteFeeBaseMsat = channelGraphEvent.FeeBaseMsat
					channelSetting.RemoteFeeRateMilliMsat = channelGraphEvent.FeeRateMilliMsat
				}
			} else {
				log.Error().Msgf("Received channel graph event for uncached channel with channelId: %v", *channelGraphEvent.ChannelId)
			}
		} else {
			log.Error().Msgf("Received channel graph event for uncached node with nodeId: %v", channelGraphEvent.NodeId)
		}
	} else if channelEvent, ok := event.(ChannelEvent); ok {
		if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
			return
		}
		if !isNodeReady(channelStateSettingsStatusCache, channelEvent.NodeId, channelStateSettingsLockCache, channelStateSettingsDeactivationTimeCache) {
			return
		}
		defer channelStateSettingsLockCache[channelGraphEvent.NodeId].RUnlock()

		nodeChannels, exists = channelStateSettingsByChannelIdCache[channelEvent.NodeId]
		if exists {
			channelSetting, exists = nodeChannels[channelEvent.ChannelId]
			if exists {
				switch channelEvent.Type {
				case lnrpc.ChannelEventUpdate_ACTIVE_CHANNEL:
					channelSetting.LocalDisabled = false
				case lnrpc.ChannelEventUpdate_INACTIVE_CHANNEL:
					channelSetting.LocalDisabled = true
				case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
					delete(nodeChannels, channelEvent.ChannelId)
				}
			} else {
				log.Error().Msgf("Received channel event for uncached channel with channelId: %v", *channelGraphEvent.ChannelId)
			}
		} else {
			log.Error().Msgf("Received channel event for uncached node with nodeId: %v", channelGraphEvent.NodeId)
		}
	} else if invoiceEvent, ok := event.(InvoiceEvent); ok {
		if invoiceEvent.NodeId == 0 {
			return
		}
		//} else if openChannelEvent, ok := event.(channels.OpenChannelResponse); ok {
		//	if openChannelEvent.NodeId == nil {
		//		return
		//	}
		//} else if closeChannelEvent, ok := event.(channels.CloseChannelResponse); ok {
		//	if closeChannelEvent.NodeId == nil {
		//		return
		//	}
		//} else if newPaymentEvent, ok := event.(payments.NewPaymentResponse); ok {
		//	if newPaymentEvent.NodeId == nil {
		//		return
		//	}
	}
}

func processManagedChannelStateSettings(managedChannelState ManagedChannelState,
	channelStateSettingsLockCache map[int]*sync.RWMutex,
	channelStateSettingsStatusCache map[int]Status,
	channelStateSettingsByChannelIdCache map[int]map[int]ManagedChannelStateSettings,
	channelStateSettingsDeactivationTimeCache map[int]time.Time) {
	switch managedChannelState.Type {
	case READ_CHANNELSTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			go SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId, channelStateSettingsLockCache, channelStateSettingsDeactivationTimeCache) {
			defer channelStateSettingsLockCache[managedChannelState.NodeId].RUnlock()
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				go SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
				break
			}
			settings, exists := settingsByChannel[managedChannelState.ChannelId]
			if !exists {
				go SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
				break
			}
			go SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, &settings)
			break
		}
		go SendToManagedChannelStateSettingsChannel(managedChannelState.StateOut, nil)
	case READ_ALL_CHANNELSTATES:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			go SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId, channelStateSettingsLockCache, channelStateSettingsDeactivationTimeCache) {
			defer channelStateSettingsLockCache[managedChannelState.NodeId].RUnlock()
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				go SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, nil)
				break
			}
			var channelStates []ManagedChannelStateSettings
			for _, channelState := range settingsByChannel {
				channelStates = append(channelStates, channelState)
			}
			go SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, channelStates)
			break
		}
		go SendToManagedChannelStatesSettingsChannel(managedChannelState.StatesOut, nil)
	case READ_CHANNELSTATELOCK:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			go SendToManagedChannelStateSettingsLockChannel(managedChannelState.LockOut, nil)
			break
		}
		go SendToManagedChannelStateSettingsLockChannel(managedChannelState.LockOut, channelStateSettingsLockCache[managedChannelState.NodeId])
	case READ_CHANNELBALANCESTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId, channelStateSettingsLockCache, channelStateSettingsDeactivationTimeCache) {
			defer channelStateSettingsLockCache[managedChannelState.NodeId].RUnlock()
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
				break
			}
			settings, exists := settingsByChannel[managedChannelState.ChannelId]
			if !exists {
				go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
				break
			}
			capacity := GetChannelSettingByChannelId(managedChannelState.ChannelId).Capacity
			channelBalanceState := processHtlcInclude(managedChannelState, settings, capacity)
			go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, &channelBalanceState)
			break
		}
		go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
	case READ_ALL_CHANNELBALANCESTATES:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			go SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
			break
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId, channelStateSettingsLockCache, channelStateSettingsDeactivationTimeCache) {
			defer channelStateSettingsLockCache[managedChannelState.NodeId].RUnlock()
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				go SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
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
			go SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, channelBalanceStates)
			break
		}
		go SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
	case WRITE_INITIAL_CHANNELSTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			break
		}
		if RWMutexWriteLocked(channelStateSettingsLockCache[managedChannelState.NodeId]) {
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				settingsByChannel = make(map[int]ManagedChannelStateSettings, 0)
			}
			settingsByChannel[managedChannelState.ChannelId] = managedChannelState.ChannelStateSettings
			channelStateSettingsByChannelIdCache[managedChannelState.NodeId] = settingsByChannel
		} else {
			log.Error().Msgf("Attempted to manipulate the channel state cache without the lock. nodeId: %v", managedChannelState.NodeId)
		}
	}
}

func SendToManagedChannelBalanceStatesSettingsChannel(ch chan []ManagedChannelBalanceStateSettings, managedChannelBalanceStateSettings []ManagedChannelBalanceStateSettings) {
	ch <- managedChannelBalanceStateSettings
}

func SendToManagedChannelBalanceStateSettingsChannel(ch chan *ManagedChannelBalanceStateSettings, managedChannelBalanceStateSettings *ManagedChannelBalanceStateSettings) {
	ch <- managedChannelBalanceStateSettings
}

func SendToManagedChannelStatesSettingsChannel(ch chan []ManagedChannelStateSettings, managedChannelBalanceStateSettings []ManagedChannelStateSettings) {
	ch <- managedChannelBalanceStateSettings
}

func SendToManagedChannelStateSettingsChannel(ch chan *ManagedChannelStateSettings, managedChannelStateSettings *ManagedChannelStateSettings) {
	ch <- managedChannelStateSettings
}

func SendToManagedChannelStateSettingsLockChannel(ch chan *sync.RWMutex, lock *sync.RWMutex) {
	ch <- lock
}

func GetChannelState(nodeId, channelId int) *ManagedChannelStateSettings {
	channelStateResponseChannel := make(chan *ManagedChannelStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:    nodeId,
		ChannelId: channelId,
		Type:      READ_CHANNELSTATE,
		StateOut:  channelStateResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelStateResponseChannel
}

func GetChannelStates(nodeId int) []ManagedChannelStateSettings {
	channelStatesResponseChannel := make(chan []ManagedChannelStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:    nodeId,
		Type:      READ_ALL_CHANNELSTATES,
		StatesOut: channelStatesResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelStatesResponseChannel
}

func GetChannelBalanceState(nodeId, channelId int, htlcInclude ChannelBalanceStateHtlcInclude) *ManagedChannelBalanceStateSettings {
	channelBalanceStateResponseChannel := make(chan *ManagedChannelBalanceStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:          nodeId,
		ChannelId:       channelId,
		HtlcInclude:     htlcInclude,
		Type:            READ_CHANNELBALANCESTATE,
		BalanceStateOut: channelBalanceStateResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelBalanceStateResponseChannel
}

func GetChannelBalanceStates(nodeId int, channelStateInclude ChannelStateInclude, htlcInclude ChannelBalanceStateHtlcInclude) []ManagedChannelBalanceStateSettings {
	channelBalanceStateResponseChannel := make(chan []ManagedChannelBalanceStateSettings)
	managedChannelState := ManagedChannelState{
		NodeId:           nodeId,
		HtlcInclude:      htlcInclude,
		StateInclude:     channelStateInclude,
		Type:             READ_CHANNELBALANCESTATE,
		BalanceStatesOut: channelBalanceStateResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelBalanceStateResponseChannel
}

func GetChannelStateLock(nodeId int) *sync.RWMutex {
	channelStateLockResponseChannel := make(chan *sync.RWMutex)
	managedChannelState := ManagedChannelState{
		NodeId:  nodeId,
		Type:    READ_CHANNELSTATELOCK,
		LockOut: channelStateLockResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelStateLockResponseChannel
}

// SetChannelState YOU NEED A WRITE LOCK FOR THIS METHOD SEE LockChannelState
func SetChannelState(nodeId, channelId int, channelStateSettings ManagedChannelStateSettings) {
	managedChannelState := ManagedChannelState{
		NodeId:               nodeId,
		ChannelId:            channelId,
		ChannelStateSettings: channelStateSettings,
		Type:                 WRITE_INITIAL_CHANNELSTATE,
	}
	ManagedChannelStateChannel <- managedChannelState
}

func isNodeReady(channelStateSettingsStatusCache map[int]Status, nodeId int,
	channelStateSettingsLockCache map[int]*sync.RWMutex, channelStateSettingsDeactivationTimeCache map[int]time.Time) bool {
	// Channel states not initialized yet
	if channelStateSettingsStatusCache[nodeId] != Active {
		deactivationTime, exists := channelStateSettingsDeactivationTimeCache[nodeId]
		if exists && time.Since(deactivationTime).Seconds() < TOLERATED_SUBSCRIPTION_DOWNTIME_SECONDS {
			log.Debug().Msgf("Node flagged as active even tough subscription is temporary down for nodeId: %v", nodeId)
			return true
		}
		return false
	}
	channelStateSettingsLockCache[nodeId].RLock()
	return true
}

func processHtlcInclude(managedChannelState ManagedChannelState, settings ManagedChannelStateSettings, capacity int64) ManagedChannelBalanceStateSettings {
	localBalance := settings.LocalBalance
	remoteBalance := settings.RemoteBalance
	if managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_DOWNWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_DOWNWARDS {
		localBalance = settings.LocalBalance - settings.PendingDecreasingHtlcAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_DOWNWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_DOWNWARDS {
		remoteBalance = settings.RemoteBalance - settings.PendingIncreasingHtlcAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_UPWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS {
		localBalance = settings.LocalBalance + settings.PendingIncreasingHtlcAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_UPWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS {
		remoteBalance = settings.RemoteBalance + settings.PendingDecreasingHtlcAmount
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
