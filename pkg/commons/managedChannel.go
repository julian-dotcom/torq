package commons

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

var ManagedChannelChannel = make(chan ManagedChannel) //nolint:gochecknoglobals

type ManagedChannelCacheOperationType uint

const (
	// READ_CHANNELID_BY_SHORTCHANNELID please provide ShortChannelId and Out
	READ_CHANNELID_BY_SHORTCHANNELID ManagedChannelCacheOperationType = iota
	// READ_CHANNELID_BY_FUNDING_TRANSACTION please provide FundingTransactionHash, FundingOutputIndex and Out
	READ_CHANNELID_BY_FUNDING_TRANSACTION
	// READ_STATUSID_BY_CHANNELID please provide ChannelId and Out
	READ_STATUSID_BY_CHANNELID
	// READ_CHANNEL_SETTINGS please provide ChannelId and ChannelIdSettingsOut
	READ_CHANNEL_SETTINGS
	// WRITE_CHANNEL Please provide ChannelId and StatusId (other values are optional in case of pending open channel)
	WRITE_CHANNEL
	// WRITE_CHANNELSTATUSID Please provide ChannelId and StatusId
	WRITE_CHANNELSTATUSID
	// DELETE_CHANNEL Please only provide ChannelId
	DELETE_CHANNEL
)

type ManagedChannel struct {
	Type                   ManagedChannelCacheOperationType
	ChannelId              int
	ShortChannelId         string
	FundingTransactionHash string
	FundingOutputIndex     int
	StatusId               int
	Out                    chan ManagedChannel
	ChannelIdSettingsOut   chan ManagedChannelSettings
}

type ManagedChannelSettings struct {
	ChannelId              int
	ShortChannelId         string
	FundingTransactionHash string
	FundingOutputIndex     int
	StatusId               int
}

// ManagedChannelCache parameter Context is for test cases...
func ManagedChannelCache(ch chan ManagedChannel, ctx context.Context) {
	channelSettingsByChannelIdCache := make(map[int]ManagedChannelSettings, 0)
	shortChannelIdCache := make(map[string]int, 0)
	channelPointCache := make(map[string]int, 0)
	channelStatusIdCache := make(map[int]int, 0)
	for {
		// TODO FIXME FEELS WRONG?
		if ctx == nil {
			managedChannel := <-ch
			processManagedChannel(managedChannel, shortChannelIdCache, channelPointCache, channelStatusIdCache, channelSettingsByChannelIdCache)
		} else {
			select {
			case <-ctx.Done():
				return
			case managedChannel := <-ch:
				processManagedChannel(managedChannel, shortChannelIdCache, channelPointCache, channelStatusIdCache, channelSettingsByChannelIdCache)
			}
		}
	}
}

func processManagedChannel(managedChannel ManagedChannel, shortChannelIdCache map[string]int, channelPointCache map[string]int, channelStatusIdCache map[int]int, channelSettingsByChannelIdCache map[int]ManagedChannelSettings) {
	switch managedChannel.Type {
	case READ_CHANNELID_BY_SHORTCHANNELID:
		managedChannel.ChannelId = shortChannelIdCache[managedChannel.ShortChannelId]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNELID_BY_FUNDING_TRANSACTION:
		managedChannel.ChannelId = channelPointCache[createChannelPoint(managedChannel)]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_STATUSID_BY_CHANNELID:
		managedChannel.StatusId = channelStatusIdCache[managedChannel.ChannelId]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNEL_SETTINGS:
		go SendToManagedChannelSettingsChannel(managedChannel.ChannelIdSettingsOut, channelSettingsByChannelIdCache[managedChannel.ChannelId])
	case WRITE_CHANNEL:
		if managedChannel.ChannelId == 0 || managedChannel.FundingTransactionHash == "" || managedChannel.FundingOutputIndex == 0 {
			log.Error().Msgf("No empty ChannelId (%v) or FundingTransactionHash (%v) or FundingOutputIndex (%v) allowed", managedChannel.ChannelId, managedChannel.FundingTransactionHash, managedChannel.FundingOutputIndex)
		} else {
			if managedChannel.ShortChannelId != "" {
				shortChannelIdCache[managedChannel.ShortChannelId] = managedChannel.ChannelId
			}
			channelPointCache[createChannelPoint(managedChannel)] = managedChannel.ChannelId
			channelStatusIdCache[managedChannel.ChannelId] = managedChannel.StatusId
			channelSettingsByChannelIdCache[managedChannel.ChannelId] = ManagedChannelSettings{
				ChannelId:              managedChannel.ChannelId,
				ShortChannelId:         managedChannel.ShortChannelId,
				StatusId:               managedChannel.StatusId,
				FundingTransactionHash: managedChannel.FundingTransactionHash,
				FundingOutputIndex:     managedChannel.FundingOutputIndex,
			}
		}
	case WRITE_CHANNELSTATUSID:
		if managedChannel.ChannelId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) allowed", managedChannel.ChannelId)
		} else {
			channelStatusIdCache[managedChannel.ChannelId] = managedChannel.StatusId
			settings := channelSettingsByChannelIdCache[managedChannel.ChannelId]
			settings.StatusId = managedChannel.StatusId
			channelSettingsByChannelIdCache[managedChannel.ChannelId] = settings
		}
	case DELETE_CHANNEL:
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
		delete(channelStatusIdCache, managedChannel.ChannelId)
		delete(channelSettingsByChannelIdCache, managedChannel.ChannelId)
	}
}

func createChannelPoint(managedChannel ManagedChannel) string {
	return fmt.Sprintf("%s:%v", managedChannel.FundingTransactionHash, managedChannel.FundingOutputIndex)
}

func SendToManagedChannelChannel(ch chan ManagedChannel, managedChannel ManagedChannel) {
	ch <- managedChannel
}

func SendToManagedChannelSettingsChannel(ch chan ManagedChannelSettings, channelSettings ManagedChannelSettings) {
	ch <- channelSettings
}

func GetChannelIdFromFundingTransaction(fundingTransactionHash string, fundingOutputIndex int) int {
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

func GetChannelIdFromShortChannelId(shortChannelId string) int {
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

func GetChannelStatusIdFromChannelId(channelId int) int {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		ChannelId: channelId,
		Type:      READ_STATUSID_BY_CHANNELID,
		Out:       channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse.StatusId
}

func GetChannelSettingsFromChannelId(channelId int) ManagedChannelSettings {
	channelResponseChannel := make(chan ManagedChannelSettings)
	managedChannel := ManagedChannel{
		ChannelId:            channelId,
		Type:                 READ_CHANNEL_SETTINGS,
		ChannelIdSettingsOut: channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	channelResponse := <-channelResponseChannel
	return channelResponse
}

func SetChannel(channelId int, shortChannelId string, statusId int, fundingTransactionHash string, fundingOutputIndex int) {
	managedChannel := ManagedChannel{
		ChannelId:              channelId,
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		StatusId:               statusId,
		Type:                   WRITE_CHANNEL,
	}
	if shortChannelId != "" {
		managedChannel.ShortChannelId = shortChannelId
	}
	ManagedChannelChannel <- managedChannel
}

func SetChannelStatus(channelId int, statusId int) {
	managedChannel := ManagedChannel{
		ChannelId: channelId,
		StatusId:  statusId,
		Type:      WRITE_CHANNELSTATUSID,
	}
	ManagedChannelChannel <- managedChannel
}

func RemoveChannel(channelId int) {
	managedChannel := ManagedChannel{
		ChannelId: channelId,
		Type:      DELETE_CHANNEL,
	}
	ManagedChannelChannel <- managedChannel
}
