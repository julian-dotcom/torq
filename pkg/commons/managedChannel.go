package commons

import (
	"context"
	"fmt"

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
	FundingTransactionHash string
	FundingOutputIndex     int
	Capacity               int64
	FirstNodeId            int
	SecondNodeId           int
	InitiatingNodeId       *int
	AcceptingNodeId        *int
	Private                bool
	Status                 ChannelStatus
	Out                    chan ManagedChannel
	ChannelSettingOut      chan ManagedChannelSettings
	ChannelSettingsOut     chan []ManagedChannelSettings
}

type ManagedChannelSettings struct {
	ChannelId              int
	ShortChannelId         string
	FundingTransactionHash string
	FundingOutputIndex     int
	Capacity               int64
	FirstNodeId            int
	SecondNodeId           int
	InitiatingNodeId       *int
	AcceptingNodeId        *int
	Private                bool
	Status                 ChannelStatus
}

func ManagedChannelCache(ch chan ManagedChannel, ctx context.Context) {
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
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNELID_BY_SHORTCHANNELID:
		managedChannel.ChannelId = allShortChannelIdCache[managedChannel.ShortChannelId]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_ACTIVE_CHANNELID_BY_FUNDING_TRANSACTION:
		managedChannel.ChannelId = channelPointCache[createChannelPoint(managedChannel)]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNELID_BY_FUNDING_TRANSACTION:
		managedChannel.ChannelId = allChannelPointCache[createChannelPoint(managedChannel)]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_STATUSID_BY_CHANNELID:
		managedChannel.Status = allChannelStatusCache[managedChannel.ChannelId]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_ALL_CHANNEL_SETTINGS:
		var channelSettings []ManagedChannelSettings
		for _, channelSetting := range allChannelSettingsByChannelIdCache {
			if channelSetting.FirstNodeId == managedChannel.FirstNodeId {
				channelSettings = append(channelSettings, channelSetting)
			}
			if channelSetting.SecondNodeId == managedChannel.FirstNodeId {
				channelSettings = append(channelSettings, channelSetting)
			}
		}
		go SendToManagedChannelSettingsChannel(managedChannel.ChannelSettingsOut, channelSettings)
	case READ_CHANNEL_SETTINGS:
		go SendToManagedChannelSettingChannel(managedChannel.ChannelSettingOut, allChannelSettingsByChannelIdCache[managedChannel.ChannelId])
	case WRITE_CHANNEL:
		if managedChannel.ChannelId == 0 || managedChannel.FundingTransactionHash == "" {
			log.Error().Msgf("No empty ChannelId (%v) or FundingTransactionHash (%v) allowed", managedChannel.ChannelId, managedChannel.FundingTransactionHash)
		} else {
			channelPoint := createChannelPoint(managedChannel)
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
				Status:                 managedChannel.Status,
				FundingTransactionHash: managedChannel.FundingTransactionHash,
				FundingOutputIndex:     managedChannel.FundingOutputIndex,
				Capacity:               managedChannel.Capacity,
				FirstNodeId:            managedChannel.FirstNodeId,
				SecondNodeId:           managedChannel.SecondNodeId,
				InitiatingNodeId:       managedChannel.InitiatingNodeId,
				AcceptingNodeId:        managedChannel.AcceptingNodeId,
				Private:                managedChannel.Private,
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

func createChannelPoint(managedChannel ManagedChannel) string {
	return fmt.Sprintf("%s:%v", managedChannel.FundingTransactionHash, managedChannel.FundingOutputIndex)
}

func SendToManagedChannelChannel(ch chan ManagedChannel, managedChannel ManagedChannel) {
	ch <- managedChannel
}

func SendToManagedChannelSettingChannel(ch chan ManagedChannelSettings, channelSettings ManagedChannelSettings) {
	ch <- channelSettings
}

func SendToManagedChannelSettingsChannel(ch chan []ManagedChannelSettings, channelSettings []ManagedChannelSettings) {
	ch <- channelSettings
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
		FirstNodeId:        nodeId,
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

func SetChannel(channelId int, shortChannelId *string, status ChannelStatus,
	fundingTransactionHash string, fundingOutputIndex int, capacity int64, private bool, firstNodeId, secondNodeId int,
	initiatingNodeId, acceptingNodeId *int) {
	managedChannel := ManagedChannel{
		ChannelId:              channelId,
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Capacity:               capacity,
		FirstNodeId:            firstNodeId,
		SecondNodeId:           secondNodeId,
		InitiatingNodeId:       initiatingNodeId,
		AcceptingNodeId:        acceptingNodeId,
		Private:                private,
		Status:                 status,
		Type:                   WRITE_CHANNEL,
	}
	if shortChannelId != nil && *shortChannelId != "" {
		managedChannel.ShortChannelId = *shortChannelId
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
