package commons

import (
	"context"
	"sync"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/broadcast"
)

var ManagedChannelStateChannel = make(chan ManagedChannelState) //nolint:gochecknoglobals

type ManagedChannelStateCacheOperationType uint

const (
	// READ_CHANNELSTATE please provide NodeId, ChannelId and Out
	READ_CHANNELSTATE ManagedChannelStateCacheOperationType = iota
	// READ_CHANNELBALANCESTATE please provide NodeId, ChannelId, HtlcInclude and BalanceStateOut
	READ_CHANNELBALANCESTATE
	// READ_ALL_CHANNELBALANCESTATES please provide NodeId, StateInclude, HtlcInclude and BalanceStatesOut
	READ_ALL_CHANNELBALANCESTATES
	// WRITE_INITIAL_CHANNELSTATE This requires the lock being active for writing! Please provide the complete information set
	WRITE_INITIAL_CHANNELSTATE
	// READ_CHANNELSTATELOCK please provide NodeId and Out
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

type ManagedChannelState struct {
	Type                 ManagedChannelStateCacheOperationType
	NodeId               int
	ChannelId            int
	ChannelStateSettings ManagedChannelStateSettings
	HtlcInclude          ChannelBalanceStateHtlcInclude
	StateInclude         ChannelStateInclude
	Out                  chan *ManagedChannelStateSettings
	BalanceStateOut      chan *ManagedChannelBalanceStateSettings
	BalanceStatesOut     chan []ManagedChannelBalanceStateSettings
	LockOut              chan *sync.RWMutex
}

type ManagedChannelStateSettings struct {
	NodeId                 int    `json:"nodeId"`
	ChannelId              int    `json:"channelId"`
	LocalBalance           int64  `json:"localBalance"`
	LocalDisabled          bool   `json:"localDisabled"`
	LocalFeeBaseMsat       int64  `json:"localFeeBaseMsat"`
	LocalFeeRateMilliMsat  int64  `json:"localFeeRateMilliMsat"`
	LocalMinHtlc           int64  `json:"localMinHtlc"`
	LocalMaxHtlcMsat       uint64 `json:"localMaxHtlcMsat"`
	LocalTimeLockDelta     uint32 `json:"localTimeLockDelta"`
	RemoteBalance          int64  `json:"remoteBalance"`
	RemoteDisabled         bool   `json:"remoteDisabled"`
	RemoteFeeBaseMsat      int64  `json:"remoteFeeBaseMsat"`
	RemoteFeeRateMilliMsat int64  `json:"remoteFeeRateMilliMsat"`
	RemoteMinHtlc          int64  `json:"remoteMinHtlc"`
	RemoteMaxHtlcMsat      uint64 `json:"remoteMaxHtlcMsat"`
	RemoteTimeLockDelta    uint32 `json:"remoteTimeLockDelta"`
	// PAYMENT HTLCs
	PendingPaymentHTLCsCount  int   `json:"pendingPaymentHTLCsCount"`
	PendingPaymentHTLCsAmount int64 `json:"pendingPaymentHTLCsAmount"`
	// INVOICE HTLCs
	PendingInvoiceHTLCsCount  int   `json:"pendingInvoiceHTLCsCount"`
	PendingInvoiceHTLCsAmount int64 `json:"pendingInvoiceHTLCsAmount"`
	// FORWARDING HTLCs Decreasing/Increasing IN RELATION TO LOCAL BALANCE
	PendingDecreasingForwardHTLCsCount  int   `json:"pendingDecreasingForwardCount"`
	PendingDecreasingForwardHTLCsAmount int64 `json:"pendingDecreasingForwardAmount"`
	PendingIncreasingForwardHTLCsCount  int   `json:"pendingIncreasingForwardHTLCsCount"`
	PendingIncreasingForwardHTLCsAmount int64 `json:"pendingIncreasingForwardHTLCsAmount"`
	// STALE INFORMATION ONLY OBTAINED VIA LND REGULAR CHECKINS SO NOT MAINTAINED
	CommitFee             int64                `json:"commitFee"`
	CommitWeight          int64                `json:"commitWeight"`
	FeePerKw              int64                `json:"feePerKw"`
	TotalSatoshisSent     int64                `json:"totalSatoshisSent"`
	NumUpdates            uint64               `json:"numUpdates"`
	ChanStatusFlags       string               `json:"chanStatusFlags"`
	LocalChanReserveSat   int64                `json:"localChanReserveSat"`
	RemoteChanReserveSat  int64                `json:"remoteChanReserveSat"`
	CommitmentType        lnrpc.CommitmentType `json:"commitmentType"`
	Lifetime              int64                `json:"lifetime"`
	TotalSatoshisReceived int64                `json:"totalSatoshisReceived"`
}

type ManagedChannelBalanceStateSettings struct {
	NodeId                     int                            `json:"nodeId"`
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
	for {
		select {
		case <-ctx.Done():
			return
		case managedChannelState := <-ch:
			processManagedChannelStateSettings(managedChannelState, channelStateSettingsLockCache, channelStateSettingsStatusCache, channelStateSettingsByChannelIdCache)
		case event := <-broadcaster.Subscribe():
			processBroadcastedEvent(event, channelStateSettingsLockCache, channelStateSettingsStatusCache, channelStateSettingsByChannelIdCache)
		}
	}
}
func processBroadcastedEvent(event interface{},
	channelStateSettingsLockCache map[int]*sync.RWMutex,
	channelStateSettingsStatusCache map[int]Status,
	channelStateSettingsByChannelIdCache map[int]map[int]ManagedChannelStateSettings) {
	var nodeChannels map[int]ManagedChannelStateSettings
	var channelSetting ManagedChannelStateSettings
	var exists bool

	if channelGraphEvent, ok := event.(ChannelGraphEvent); ok {
		if channelGraphEvent.NodeId == 0 || channelGraphEvent.ChannelId == nil || *channelGraphEvent.ChannelId == 0 ||
			channelGraphEvent.AnnouncingNodeId == nil || *channelGraphEvent.AnnouncingNodeId == 0 ||
			channelGraphEvent.ConnectingNodeId == nil || *channelGraphEvent.ConnectingNodeId == 0 {
			return
		}
		if !isNodeReady(channelStateSettingsStatusCache, channelGraphEvent.NodeId, channelStateSettingsLockCache) {
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
		if !isNodeReady(channelStateSettingsStatusCache, channelEvent.NodeId, channelStateSettingsLockCache) {
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
	channelStateSettingsByChannelIdCache map[int]map[int]ManagedChannelStateSettings) {
	switch managedChannelState.Type {
	case READ_CHANNELSTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
			go SendToManagedChannelStateSettingsChannel(managedChannelState.Out, nil)
		}
		if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId, channelStateSettingsLockCache) {
			defer channelStateSettingsLockCache[managedChannelState.NodeId].RUnlock()
			settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
			if !exists {
				go SendToManagedChannelStateSettingsChannel(managedChannelState.Out, nil)
			}
			settings, exists := settingsByChannel[managedChannelState.ChannelId]
			if !exists {
				go SendToManagedChannelStateSettingsChannel(managedChannelState.Out, nil)
			}
			go SendToManagedChannelStateSettingsChannel(managedChannelState.Out, &settings)
		} else {
			go SendToManagedChannelStateSettingsChannel(managedChannelState.Out, nil)
		}
	case READ_CHANNELSTATELOCK:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
			go SendToManagedChannelStateSettingsLockChannel(managedChannelState.LockOut, nil)
		}
		go SendToManagedChannelStateSettingsLockChannel(managedChannelState.LockOut, channelStateSettingsLockCache[managedChannelState.NodeId])
	case READ_CHANNELBALANCESTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
		} else {
			if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId, channelStateSettingsLockCache) {
				defer channelStateSettingsLockCache[managedChannelState.NodeId].RUnlock()
				settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
				if !exists {
					go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
				}
				settings, exists := settingsByChannel[managedChannelState.ChannelId]
				if !exists {
					go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
				}
				capacity := GetChannelSettingByChannelId(managedChannelState.ChannelId).Capacity
				channelBalanceState := processHtlcInclude(managedChannelState, settings, capacity)
				go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, &channelBalanceState)
			} else {
				go SendToManagedChannelBalanceStateSettingsChannel(managedChannelState.BalanceStateOut, nil)
			}
		}
	case READ_ALL_CHANNELBALANCESTATES:
		if managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed", managedChannelState.NodeId)
		} else {
			if isNodeReady(channelStateSettingsStatusCache, managedChannelState.NodeId, channelStateSettingsLockCache) {
				defer channelStateSettingsLockCache[managedChannelState.NodeId].RUnlock()
				settingsByChannel, exists := channelStateSettingsByChannelIdCache[managedChannelState.NodeId]
				if !exists {
					go SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
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
			} else {
				go SendToManagedChannelBalanceStatesSettingsChannel(managedChannelState.BalanceStatesOut, nil)
			}
		}
	case WRITE_INITIAL_CHANNELSTATE:
		if managedChannelState.ChannelId == 0 || managedChannelState.NodeId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) nor NodeId (%v) allowed", managedChannelState.ChannelId, managedChannelState.NodeId)
		} else {
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
}

func SendToManagedChannelBalanceStatesSettingsChannel(ch chan []ManagedChannelBalanceStateSettings, managedChannelBalanceStateSettings []ManagedChannelBalanceStateSettings) {
	ch <- managedChannelBalanceStateSettings
}

func SendToManagedChannelBalanceStateSettingsChannel(ch chan *ManagedChannelBalanceStateSettings, managedChannelBalanceStateSettings *ManagedChannelBalanceStateSettings) {
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
		Out:       channelStateResponseChannel,
	}
	ManagedChannelStateChannel <- managedChannelState
	return <-channelStateResponseChannel
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

func isNodeReady(channelStateSettingsStatusCache map[int]Status, nodeId int, channelStateSettingsLockCache map[int]*sync.RWMutex) bool {
	// Channel states not initialized yet
	if channelStateSettingsStatusCache[nodeId] != Active {
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
		localBalance = settings.LocalBalance - settings.PendingDecreasingForwardHTLCsAmount - settings.PendingPaymentHTLCsAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_DOWNWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_DOWNWARDS {
		remoteBalance = settings.RemoteBalance - settings.PendingIncreasingForwardHTLCsAmount - settings.PendingInvoiceHTLCsAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_BALANCE_ADJUSTED_UPWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS {
		localBalance = settings.LocalBalance + settings.PendingIncreasingForwardHTLCsAmount + settings.PendingInvoiceHTLCsAmount
	}
	if managedChannelState.HtlcInclude == PENDING_HTLCS_REMOTE_BALANCE_ADJUSTED_UPWARDS ||
		managedChannelState.HtlcInclude == PENDING_HTLCS_LOCAL_AND_REMOTE_BALANCE_ADJUSTED_UPWARDS {
		remoteBalance = settings.RemoteBalance + settings.PendingDecreasingForwardHTLCsAmount + settings.PendingPaymentHTLCsAmount
	}
	return ManagedChannelBalanceStateSettings{
		NodeId:                     managedChannelState.NodeId,
		ChannelId:                  managedChannelState.ChannelId,
		HtlcInclude:                managedChannelState.HtlcInclude,
		LocalBalance:               localBalance,
		LocalBalancePerMilleRatio:  int(settings.LocalBalance / capacity * 1000),
		RemoteBalance:              remoteBalance,
		RemoteBalancePerMilleRatio: int(settings.RemoteBalance / capacity * 1000),
	}
}
