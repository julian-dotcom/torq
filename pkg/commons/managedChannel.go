package commons

import (
	"context"

	"github.com/rs/zerolog/log"
)

var ManagedChannelChannel = make(chan ManagedChannel) //nolint:gochecknoglobals

type ManagedChannelCacheOperationType uint

const (
	// READ_CHANNELID_BY_SHORTCHANNELID please provide ShortChannelId and Out
	READ_CHANNELID_BY_SHORTCHANNELID ManagedChannelCacheOperationType = iota
	// READ_CHANNELID_BY_CHANNELPOINT please provide LndChannelPoint and Out
	READ_CHANNELID_BY_CHANNELPOINT
	// READ_STATUSID_BY_CHANNELID please provide ChannelId and Out
	READ_STATUSID_BY_CHANNELID
	// READ_ALL_LNDCHANNELPOINTS please provide ChannelPointsOut
	READ_ALL_LNDCHANNELPOINTS
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
	Type                 ManagedChannelCacheOperationType
	ChannelId            int
	ShortChannelId       string
	LndChannelPoint      string
	StatusId             int
	Out                  chan ManagedChannel
	ChannelPointsOut     chan []string
	ChannelIdSettingsOut chan ManagedChannelSettings
}

type ManagedChannelSettings struct {
	ChannelId       int
	ShortChannelId  string
	LndChannelPoint string
	StatusId        int
}

// ManagedChannelCache parameter Context is for test cases...
func ManagedChannelCache(ch chan ManagedChannel, ctx context.Context) {
	channelSettingsByChannelIdCache := make(map[int]ManagedChannelSettings, 0)
	shortChannelIdCache := make(map[string]int, 0)
	lndChannelPointCache := make(map[string]int, 0)
	channelStatusIdCache := make(map[int]int, 0)
	for {
		// TODO FIXME FEELS WRONG?
		if ctx == nil {
			managedChannel := <-ch
			processManagedChannel(managedChannel, shortChannelIdCache, lndChannelPointCache, channelStatusIdCache, channelSettingsByChannelIdCache)
		} else {
			select {
			case <-ctx.Done():
				return
			case managedChannel := <-ch:
				processManagedChannel(managedChannel, shortChannelIdCache, lndChannelPointCache, channelStatusIdCache, channelSettingsByChannelIdCache)
			}
		}
	}
}

func processManagedChannel(managedChannel ManagedChannel, shortChannelIdCache map[string]int, lndChannelPointCache map[string]int, channelStatusIdCache map[int]int, channelSettingsByChannelIdCache map[int]ManagedChannelSettings) {
	switch managedChannel.Type {
	case READ_CHANNELID_BY_SHORTCHANNELID:
		managedChannel.ChannelId = shortChannelIdCache[managedChannel.ShortChannelId]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_CHANNELID_BY_CHANNELPOINT:
		managedChannel.ChannelId = lndChannelPointCache[managedChannel.LndChannelPoint]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_STATUSID_BY_CHANNELID:
		managedChannel.StatusId = channelStatusIdCache[managedChannel.ChannelId]
		go SendToManagedChannelChannel(managedChannel.Out, managedChannel)
	case READ_ALL_LNDCHANNELPOINTS:
		allChannelPoints := make([]string, 0, len(lndChannelPointCache))
		for lndChannelPoint := range lndChannelPointCache {
			allChannelPoints = append(allChannelPoints, lndChannelPoint)
		}
		go SendToManagedChannelIdsChannel(managedChannel.ChannelPointsOut, allChannelPoints)
	case READ_CHANNEL_SETTINGS:
		go SendToManagedChannelSettingsChannel(managedChannel.ChannelIdSettingsOut, channelSettingsByChannelIdCache[managedChannel.ChannelId])
	case WRITE_CHANNEL:
		if managedChannel.ChannelId == 0 || managedChannel.LndChannelPoint == "" {
			log.Error().Msgf("No empty ChannelId (%v) or LndChannelPoint (%v) allowed", managedChannel.ChannelId, managedChannel.LndChannelPoint)
		} else {
			if managedChannel.ShortChannelId != "" {
				shortChannelIdCache[managedChannel.ShortChannelId] = managedChannel.ChannelId
			}
			lndChannelPointCache[managedChannel.LndChannelPoint] = managedChannel.ChannelId
			channelStatusIdCache[managedChannel.ChannelId] = managedChannel.StatusId
			channelSettingsByChannelIdCache[managedChannel.ChannelId] = ManagedChannelSettings{
				ChannelId:       managedChannel.ChannelId,
				ShortChannelId:  managedChannel.ShortChannelId,
				StatusId:        managedChannel.StatusId,
				LndChannelPoint: managedChannel.LndChannelPoint,
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
		for lndChannelPoint, channelId := range lndChannelPointCache {
			if channelId == managedChannel.ChannelId {
				matchingChannelPoint = lndChannelPoint
				break
			}
		}
		if matchingChannelPoint != "" {
			delete(lndChannelPointCache, matchingChannelPoint)
		}
		delete(channelStatusIdCache, managedChannel.ChannelId)
		delete(channelSettingsByChannelIdCache, managedChannel.ChannelId)
	}
}

func SendToManagedChannelChannel(ch chan ManagedChannel, managedChannel ManagedChannel) {
	ch <- managedChannel
}

func SendToManagedChannelIdsChannel(ch chan []string, lndChannelPoints []string) {
	ch <- lndChannelPoints
}

func SendToManagedChannelSettingsChannel(ch chan ManagedChannelSettings, channelSettings ManagedChannelSettings) {
	ch <- channelSettings
}

func GetChannelIdFromChannelPoint(lndChannelPoint string) int {
	channelResponseChannel := make(chan ManagedChannel)
	managedChannel := ManagedChannel{
		LndChannelPoint: lndChannelPoint,
		Type:            READ_CHANNELID_BY_CHANNELPOINT,
		Out:             channelResponseChannel,
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

func GetAllLndChannelPoints() []string {
	channelResponseChannel := make(chan []string)
	managedChannel := ManagedChannel{
		Type:             READ_ALL_LNDCHANNELPOINTS,
		ChannelPointsOut: channelResponseChannel,
	}
	ManagedChannelChannel <- managedChannel
	return <-channelResponseChannel
}

func SetOpeningChannel(channelId int, lndChannelPoint string, statusId int) {
	managedChannel := ManagedChannel{
		ChannelId:       channelId,
		LndChannelPoint: lndChannelPoint,
		StatusId:        statusId,
		Type:            WRITE_CHANNEL,
	}
	ManagedChannelChannel <- managedChannel
}

func SetChannel(channelId int, shortChannelId string, statusId int, lndChannelPoint string) {
	managedChannel := ManagedChannel{
		ChannelId:       channelId,
		ShortChannelId:  shortChannelId,
		LndChannelPoint: lndChannelPoint,
		StatusId:        statusId,
		Type:            WRITE_CHANNEL,
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
