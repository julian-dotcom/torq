package commons

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

var ManagedChannelChannel = make(chan ManagedChannel) //nolint:gochecknoglobals

type ManagedChannelCacheOperationType uint

const (
	// READ_ACTIVE_CHANNELID_BY_SHORTCHANNELID please provide ShortChannelId and Out
	READ_ACTIVE_CHANNELID_BY_SHORTCHANNELID ManagedChannelCacheOperationType = iota
	// READ_CHANNELID_BY_SHORTCHANNELID please provide ShortChannelId and Out
	READ_CHANNELID_BY_SHORTCHANNELID
	// READ_ACTIVE_CHANNELID_BY_FUNDING_TRANSACTION please provide FundingTransactionHash, FundingOutputIndex and Out
	READ_ACTIVE_CHANNELID_BY_FUNDING_TRANSACTION
	// READ_CHANNELID_BY_FUNDING_TRANSACTION please provide FundingTransactionHash, FundingOutputIndex and Out
	READ_CHANNELID_BY_FUNDING_TRANSACTION
	// READ_CHANNELIDS_BY_NODE_ID please provide NodeId and ChannelIdsOut
	READ_CHANNELIDS_BY_NODE_ID
	// READ_STATUSID_BY_CHANNELID please provide ChannelId and Out
	READ_STATUSID_BY_CHANNELID
	// READ_CHANNEL_SETTINGS please provide ChannelId and ChannelSettingOut
	READ_CHANNEL_SETTINGS
	// READ_ALL_CHANNEL_SETTINGS please provide FirstNodeId and ChannelSettingsOut
	READ_ALL_CHANNEL_SETTINGS
	// WRITE_CHANNEL Please provide ChannelId, FundingTransactionHash, FundingOutputIndex and Status (other values are optional in case of pending open channel)
	WRITE_CHANNEL
	// WRITE_CHANNELSTATUSID Please provide ChannelId and Status
	WRITE_CHANNELSTATUSID
)

type ManagedChannel struct {
	Type                   ManagedChannelCacheOperationType
	ChannelId              int
	ShortChannelId         string
	LndShortChannelId      uint64
	FundingTransactionHash string
	FundingOutputIndex     int
	FundingBlockHeight     *uint32
	FundedOn               *time.Time
	Capacity               int64
	NodeId                 int
	FirstNodeId            int
	SecondNodeId           int
	InitiatingNodeId       *int
	AcceptingNodeId        *int
	Private                bool
	Status                 ChannelStatus
	ClosingTransactionHash *string
	ClosingNodeId          *int
	ClosingBlockHeight     *uint32
	ClosedOn               *time.Time
	Flags                  ChannelFlags
	Out                    chan<- ManagedChannel
	ChannelSettingOut      chan<- ManagedChannelSettings
	ChannelSettingsOut     chan<- []ManagedChannelSettings
	ChannelIdsOut          chan<- []int
}

type ManagedChannelSettings struct {
	ChannelId              int
	ShortChannelId         string
	LndShortChannelId      uint64
	FundingTransactionHash string
	FundingOutputIndex     int
	FundingBlockHeight     *uint32
	FundedOn               *time.Time
	Capacity               int64
	FirstNodeId            int
	SecondNodeId           int
	InitiatingNodeId       *int
	AcceptingNodeId        *int
	Private                bool
	Status                 ChannelStatus
	ClosingTransactionHash *string
	ClosingNodeId          *int
	ClosingBlockHeight     *uint32
	ClosedOn               *time.Time
	Flags                  ChannelFlags
}

func (channelSettings *ManagedChannelSettings) AddChannelFlags(flags ChannelFlags) {
	channelSettings.Flags |= flags
}
func (channelSettings *ManagedChannelSettings) HasChannelFlags(flags ChannelFlags) bool {
	return channelSettings.Flags&flags != 0
}
func (channelSettings *ManagedChannelSettings) RemoveChannelFlags(flags ChannelFlags) {
	channelSettings.Flags &= ^flags
}

func ManagedChannelCache(ch <-chan ManagedChannel, ctx context.Context) {
	allChannelSettingsByChannelIdCache := make(map[int]ManagedChannelSettings, 0)
	shortChannelIdCache := make(map[string]int, 0)
	allShortChannelIdCache := make(map[string]int, 0)
	channelPointCache := make(map[string]int, 0)
	allChannelPointCache := make(map[string]int, 0)
	allChannelStatusCache := make(map[int]ChannelStatus, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedChannel := <-ch:
			processManagedChannel(managedChannel,
				shortChannelIdCache, allShortChannelIdCache,
				channelPointCache, allChannelPointCache, allChannelStatusCache, allChannelSettingsByChannelIdCache)
		}
	}
}

func processManagedChannel(managedChannel ManagedChannel,
	shortChannelIdCache map[string]int, allShortChannelIdCache map[string]int,
	channelPointCache map[string]int, allChannelPointCache map[string]int,
	allChannelStatusCache map[int]ChannelStatus, allChannelSettingsByChannelIdCache map[int]ManagedChannelSettings) {
	switch managedChannel.Type {
	case READ_ACTIVE_CHANNELID_BY_SHORTCHANNELID:
		managedChannel.ChannelId = shortChannelIdCache[managedChannel.ShortChannelId]
		SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNELID_BY_SHORTCHANNELID:
		managedChannel.ChannelId = allShortChannelIdCache[managedChannel.ShortChannelId]
		SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_ACTIVE_CHANNELID_BY_FUNDING_TRANSACTION:
		managedChannel.ChannelId = channelPointCache[createChannelPoint(managedChannel)]
		SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNELID_BY_FUNDING_TRANSACTION:
		managedChannel.ChannelId = allChannelPointCache[createChannelPoint(managedChannel)]
		SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_STATUSID_BY_CHANNELID:
		managedChannel.Status = allChannelStatusCache[managedChannel.ChannelId]
		SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNELIDS_BY_NODE_ID:
		var channelIds []int
		for _, channelSetting := range allChannelSettingsByChannelIdCache {
			if channelSetting.FirstNodeId == managedChannel.NodeId {
				channelIds = append(channelIds, channelSetting.ChannelId)
			}
			if channelSetting.SecondNodeId == managedChannel.NodeId {
				channelIds = append(channelIds, channelSetting.ChannelId)
			}
		}
		SendToManagedChannelIdsChannel(managedChannel.ChannelIdsOut, channelIds)
	case READ_ALL_CHANNEL_SETTINGS:
		var channelSettings []ManagedChannelSettings
		for _, channelSetting := range allChannelSettingsByChannelIdCache {
			if channelSetting.FirstNodeId == managedChannel.NodeId {
				channelSettings = append(channelSettings, channelSetting)
			}
			if channelSetting.SecondNodeId == managedChannel.NodeId {
				channelSettings = append(channelSettings, channelSetting)
			}
		}
		SendToManagedChannelSettingsChannel(managedChannel.ChannelSettingsOut, channelSettings)
	case READ_CHANNEL_SETTINGS:
		SendToManagedChannelSettingChannel(managedChannel.ChannelSettingOut, allChannelSettingsByChannelIdCache[managedChannel.ChannelId])
	case WRITE_CHANNEL:
		if managedChannel.ChannelId == 0 || managedChannel.FundingTransactionHash == "" {
			log.Error().Msgf("No empty ChannelId (%v) or FundingTransactionHash (%v) allowed", managedChannel.ChannelId, managedChannel.FundingTransactionHash)
		} else {
			channelPoint := createChannelPoint(managedChannel)
			parseAndConvertShortIds(managedChannel)
			if managedChannel.Status < CooperativeClosed {
				if managedChannel.ShortChannelId != "" {
					shortChannelIdCache[managedChannel.ShortChannelId] = managedChannel.ChannelId
				}
				channelPointCache[channelPoint] = managedChannel.ChannelId
			}
			if managedChannel.ShortChannelId != "" {
				allShortChannelIdCache[managedChannel.ShortChannelId] = managedChannel.ChannelId
			}
			allChannelPointCache[channelPoint] = managedChannel.ChannelId
			allChannelStatusCache[managedChannel.ChannelId] = managedChannel.Status
			allChannelSettingsByChannelIdCache[managedChannel.ChannelId] = ManagedChannelSettings{
				ChannelId:              managedChannel.ChannelId,
				ShortChannelId:         managedChannel.ShortChannelId,
				LndShortChannelId:      managedChannel.LndShortChannelId,
				Status:                 managedChannel.Status,
				FundingTransactionHash: managedChannel.FundingTransactionHash,
				FundingOutputIndex:     managedChannel.FundingOutputIndex,
				FundingBlockHeight:     managedChannel.FundingBlockHeight,
				FundedOn:               managedChannel.FundedOn,
				Capacity:               managedChannel.Capacity,
				FirstNodeId:            managedChannel.FirstNodeId,
				SecondNodeId:           managedChannel.SecondNodeId,
				InitiatingNodeId:       managedChannel.InitiatingNodeId,
				AcceptingNodeId:        managedChannel.AcceptingNodeId,
				Private:                managedChannel.Private,
				ClosingTransactionHash: managedChannel.ClosingTransactionHash,
				ClosingNodeId:          managedChannel.ClosingNodeId,
				ClosingBlockHeight:     managedChannel.ClosingBlockHeight,
				ClosedOn:               managedChannel.ClosedOn,
				Flags:                  managedChannel.Flags,
			}
		}
	case WRITE_CHANNELSTATUSID:
		if managedChannel.ChannelId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) allowed", managedChannel.ChannelId)
		} else {
			if managedChannel.Status >= CooperativeClosed {
				var matchingShortChannelId string
				for shortChannelId, channelId := range shortChannelIdCache {
					if channelId == managedChannel.ChannelId {
						matchingShortChannelId = shortChannelId
						break
					}
				}
				if matchingShortChannelId != "" {
					delete(shortChannelIdCache, matchingShortChannelId)
				}
				var matchingChannelPoint string
				for channelPoint, channelId := range channelPointCache {
					if channelId == managedChannel.ChannelId {
						matchingChannelPoint = channelPoint
						break
					}
				}
				if matchingChannelPoint != "" {
					delete(channelPointCache, matchingChannelPoint)
				}
			}
			allChannelStatusCache[managedChannel.ChannelId] = managedChannel.Status
			settings := allChannelSettingsByChannelIdCache[managedChannel.ChannelId]
			settings.Status = managedChannel.Status
			allChannelSettingsByChannelIdCache[managedChannel.ChannelId] = settings
		}
	}
}

func parseAndConvertShortIds(managedChannel ManagedChannel) {
	if managedChannel.ShortChannelId == "" && managedChannel.LndShortChannelId != 0 {
		managedChannel.ShortChannelId = ConvertLNDShortChannelID(managedChannel.LndShortChannelId)
	}
	if managedChannel.LndShortChannelId == 0 && managedChannel.ShortChannelId != "" {
		var err error
		managedChannel.LndShortChannelId, err = ConvertShortChannelIDToLND(managedChannel.ShortChannelId)
		if err != nil {
			log.Error().Msgf("Could not convert ShortChannelId (%v) into LndShortChannelId", managedChannel.ShortChannelId)
		}
	}
}

func createChannelPoint(managedChannel ManagedChannel) string {
	return fmt.Sprintf("%s:%v", managedChannel.FundingTransactionHash, managedChannel.FundingOutputIndex)
}

func GetActiveChannelIdByFundingTransaction(fundingTransactionHash string, fundingOutputIndex int) int {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Type:                   READ_ACTIVE_CHANNELID_BY_FUNDING_TRANSACTION,
		Out:                    channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByChannelPoint(channelPoint string) int {
	fundingTransactionHash, fundingOutputIndex := ParseChannelPoint(channelPoint)
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Type:                   READ_CHANNELID_BY_FUNDING_TRANSACTION,
		Out:                    channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByFundingTransaction(fundingTransactionHash string, fundingOutputIndex int) int {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Type:                   READ_CHANNELID_BY_FUNDING_TRANSACTION,
		Out:                    channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetActiveChannelIdByShortChannelId(shortChannelId string) int {
	if shortChannelId == "" || shortChannelId == "0x0x0" {
		return 0
	}
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		ShortChannelId: shortChannelId,
		Type:           READ_ACTIVE_CHANNELID_BY_SHORTCHANNELID,
		Out:            channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByShortChannelId(shortChannelId string) int {
	if shortChannelId == "" || shortChannelId == "0x0x0" {
		return 0
	}
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		ShortChannelId: shortChannelId,
		Type:           READ_CHANNELID_BY_SHORTCHANNELID,
		Out:            channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByLndShortChannelId(lndShortChannelId uint64) int {
	if lndShortChannelId == 0 {
		return 0
	}
	shortChannelId := ConvertLNDShortChannelID(lndShortChannelId)
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		ShortChannelId: shortChannelId,
		Type:           READ_CHANNELID_BY_SHORTCHANNELID,
		Out:            channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdsByNodeId(peerNodeId int) []int {
	channelIdsResponseChannel := make(chan []int)
	managedChannel := ManagedChannel{
		NodeId:        peerNodeId,
		Type:          READ_CHANNELIDS_BY_NODE_ID,
		ChannelIdsOut: channelIdsResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	return <-channelIdsResponseChannel
}

//func GetAllChannelIds() []int {
//	var allChannelIds []int
//	torqNodeIds := GetAllTorqNodeIds()
//	for _, torqNodeId := range torqNodeIds {
//		allChannelIds = append(allChannelIds, GetChannelIdsByNodeId(torqNodeId)...)
//	}
//	return allChannelIds
//}

func GetChannelStatusByChannelId(channelId int) ChannelStatus {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		ChannelId: channelId,
		Type:      READ_STATUSID_BY_CHANNELID,
		Out:       channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.Status
}

func GetChannelSettingsByNodeId(nodeId int) []ManagedChannelSettings {
	channelResponseChannel := make(chan []ManagedChannelSettings)
	managedChannel := ManagedChannel{
		NodeId:             nodeId,
		Type:               READ_ALL_CHANNEL_SETTINGS,
		ChannelSettingsOut: channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	return <-channelResponseChannel
}

func GetChannelSettingByChannelId(channelId int) ManagedChannelSettings {
	channelResponseChannel := make(chan ManagedChannelSettings)
	managedChannel := ManagedChannel{
		ChannelId:         channelId,
		Type:              READ_CHANNEL_SETTINGS,
		ChannelSettingOut: channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	return <-channelResponseChannel
}

func SetChannel(channelId int, shortChannelId *string, lndShortChannelId *uint64, status ChannelStatus,
	fundingTransactionHash string, fundingOutputIndex int,
	fundingBlockHeight *uint32, fundedOn *time.Time,
	capacity int64, private bool, firstNodeId int, secondNodeId int,
	initiatingNodeId *int, acceptingNodeId *int,
	closingTransactionHash *string, closingNodeId *int, closingBlockHeight *uint32, closedOn *time.Time,
	flags ChannelFlags) {

	managedChannel := ManagedChannel{
		ChannelId:              channelId,
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		FundingBlockHeight:     fundingBlockHeight,
		FundedOn:               fundedOn,
		Capacity:               capacity,
		FirstNodeId:            firstNodeId,
		SecondNodeId:           secondNodeId,
		InitiatingNodeId:       initiatingNodeId,
		AcceptingNodeId:        acceptingNodeId,
		Private:                private,
		Status:                 status,
		ClosingTransactionHash: closingTransactionHash,
		ClosingNodeId:          closingNodeId,
		ClosingBlockHeight:     closingBlockHeight,
		ClosedOn:               closedOn,
		Flags:                  flags,
		Type:                   WRITE_CHANNEL,
	}
	if shortChannelId != nil && *shortChannelId != "" {
		managedChannel.ShortChannelId = *shortChannelId
	}
	if lndShortChannelId != nil && *lndShortChannelId != 0 {
		managedChannel.LndShortChannelId = *lndShortChannelId
	}
	ManagedChannelChannel <- managedChannel
}

func SetChannelStatus(channelId int, status ChannelStatus) {
	managedChannel := ManagedChannel{
		ChannelId: channelId,
		Status:    status,
		Type:      WRITE_CHANNELSTATUSID,
	}
	ManagedChannelChannel <- managedChannel
}
