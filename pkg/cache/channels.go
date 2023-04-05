package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/core"
)

var ChannelsCacheChannel = make(chan ChannelCache) //nolint:gochecknoglobals

type ChannelCacheOperationType uint
type channelId int
type channelPoint string
type shortChannelId string

const (
	// readActiveChannelIdByShortchannelId please provide ShortChannelId and Out
	readActiveChannelIdByShortchannelId ChannelCacheOperationType = iota
	// readChannelIdByShortchannelId please provide ShortChannelId and Out
	readChannelIdByShortchannelId
	// readActiveChannelIdByFundingTransaction please provide FundingTransactionHash, FundingOutputIndex and Out
	readActiveChannelIdByFundingTransaction
	// readChannelIdByFundingTransaction please provide FundingTransactionHash, FundingOutputIndex and Out
	readChannelIdByFundingTransaction
	// readChannelIdsByNodeId please provide NodeId and ChannelIdsOut
	readChannelIdsByNodeId
	// readStatusIdByChannelId please provide ChannelId and Out
	readStatusIdByChannelId
	// readChannelSettings please provide ChannelId and ChannelSettingOut
	readChannelSettings
	// readAllChannelSettings please provide FirstNodeId and ChannelSettingsOut
	readAllChannelSettings
	// writeChannel Please provide ChannelId, FundingTransactionHash, FundingOutputIndex and Status
	//(other values are optional in case of pending open channel)
	writeChannel
	// writeChannelStatusId Please provide ChannelId and Status
	writeChannelStatusId
)

type ChannelCache struct {
	Type                   ChannelCacheOperationType
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
	Status                 core.ChannelStatus
	ClosingTransactionHash *string
	ClosingNodeId          *int
	ClosingBlockHeight     *uint32
	ClosedOn               *time.Time
	Flags                  core.ChannelFlags
	Out                    chan<- ChannelCache
	ChannelSettingOut      chan<- ChannelSettingsCache
	ChannelSettingsOut     chan<- []ChannelSettingsCache
	ChannelIdsOut          chan<- []int
}

type ChannelSettingsCache struct {
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
	Status                 core.ChannelStatus
	ClosingTransactionHash *string
	ClosingNodeId          *int
	ClosingBlockHeight     *uint32
	ClosedOn               *time.Time
	Flags                  core.ChannelFlags
}

func (channel *ChannelSettingsCache) AddChannelFlags(flags core.ChannelFlags) {
	channel.Flags = channel.Flags.AddChannelFlag(flags)
}
func (channel *ChannelSettingsCache) HasChannelFlags(flags core.ChannelFlags) bool {
	return channel.Flags.HasChannelFlag(flags)
}
func (channel *ChannelSettingsCache) RemoveChannelFlags(flags core.ChannelFlags) {
	channel.Flags = channel.Flags.RemoveChannelFlag(flags)
}

func ChannelsCacheHandler(ch <-chan ChannelCache, ctx context.Context) {
	allChannelSettingsByChannelIdCache := make(map[channelId]ChannelSettingsCache, 0)
	shortChannelIdCache := make(map[shortChannelId]channelId, 0)
	allShortChannelIdCache := make(map[shortChannelId]channelId, 0)
	channelPointCache := make(map[channelPoint]channelId, 0)
	allChannelPointCache := make(map[channelPoint]channelId, 0)
	allChannelStatusCache := make(map[channelId]core.ChannelStatus, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case channelCache := <-ch:
			handleChannelOperation(channelCache,
				shortChannelIdCache, allShortChannelIdCache,
				channelPointCache, allChannelPointCache, allChannelStatusCache, allChannelSettingsByChannelIdCache)
		}
	}
}

func handleChannelOperation(channelCache ChannelCache,
	shortChannelIdCache map[shortChannelId]channelId, allShortChannelIdCache map[shortChannelId]channelId,
	channelPointCache map[channelPoint]channelId, allChannelPointCache map[channelPoint]channelId,
	allChannelStatusCache map[channelId]core.ChannelStatus, allChannelSettingsByChannelIdCache map[channelId]ChannelSettingsCache) {
	switch channelCache.Type {
	case readActiveChannelIdByShortchannelId:
		channelCache.ChannelId = int(shortChannelIdCache[shortChannelId(channelCache.ShortChannelId)])
		channelCache.Out <- channelCache
	case readChannelIdByShortchannelId:
		channelCache.ChannelId = int(allShortChannelIdCache[shortChannelId(channelCache.ShortChannelId)])
		channelCache.Out <- channelCache
	case readActiveChannelIdByFundingTransaction:
		channelCache.ChannelId =
			int(channelPointCache[createChannelPoint(channelCache.FundingTransactionHash, channelCache.FundingOutputIndex)])
		channelCache.Out <- channelCache
	case readChannelIdByFundingTransaction:
		channelCache.ChannelId =
			int(allChannelPointCache[createChannelPoint(channelCache.FundingTransactionHash, channelCache.FundingOutputIndex)])
		channelCache.Out <- channelCache
	case readStatusIdByChannelId:
		channelCache.Status = allChannelStatusCache[channelId(channelCache.ChannelId)]
		channelCache.Out <- channelCache
	case readChannelIdsByNodeId:
		var channelIds []int
		for _, channelSetting := range allChannelSettingsByChannelIdCache {
			if channelSetting.FirstNodeId == channelCache.NodeId {
				channelIds = append(channelIds, channelSetting.ChannelId)
			}
			if channelSetting.SecondNodeId == channelCache.NodeId {
				channelIds = append(channelIds, channelSetting.ChannelId)
			}
		}
		channelCache.ChannelIdsOut <- channelIds
	case readAllChannelSettings:
		var channelSettings []ChannelSettingsCache
		for _, channelSetting := range allChannelSettingsByChannelIdCache {
			if channelSetting.FirstNodeId == channelCache.NodeId {
				channelSettings = append(channelSettings, channelSetting)
			}
			if channelSetting.SecondNodeId == channelCache.NodeId {
				channelSettings = append(channelSettings, channelSetting)
			}
		}
		channelCache.ChannelSettingsOut <- channelSettings
	case readChannelSettings:
		channelCache.ChannelSettingOut <- allChannelSettingsByChannelIdCache[channelId(channelCache.ChannelId)]
	case writeChannel:
		if channelCache.ChannelId == 0 || channelCache.FundingTransactionHash == "" {
			log.Error().Msgf("No empty ChannelId (%v) or FundingTransactionHash (%v) allowed",
				channelCache.ChannelId, channelCache.FundingTransactionHash)
		} else {
			cp := createChannelPoint(channelCache.FundingTransactionHash, channelCache.FundingOutputIndex)
			parseAndConvertShortIds(channelCache)
			if channelCache.Status < core.CooperativeClosed {
				if channelCache.ShortChannelId != "" {
					shortChannelIdCache[shortChannelId(channelCache.ShortChannelId)] = channelId(channelCache.ChannelId)
				}
				channelPointCache[cp] = channelId(channelCache.ChannelId)
			} else {
				if channelCache.ShortChannelId != "" {
					delete(shortChannelIdCache, shortChannelId(channelCache.ShortChannelId))
				}
				delete(channelPointCache, cp)
			}
			if channelCache.ShortChannelId != "" {
				allShortChannelIdCache[shortChannelId(channelCache.ShortChannelId)] = channelId(channelCache.ChannelId)
			}
			allChannelPointCache[cp] = channelId(channelCache.ChannelId)
			allChannelStatusCache[channelId(channelCache.ChannelId)] = channelCache.Status
			allChannelSettingsByChannelIdCache[channelId(channelCache.ChannelId)] = ChannelSettingsCache{
				ChannelId:              channelCache.ChannelId,
				ShortChannelId:         channelCache.ShortChannelId,
				LndShortChannelId:      channelCache.LndShortChannelId,
				Status:                 channelCache.Status,
				FundingTransactionHash: channelCache.FundingTransactionHash,
				FundingOutputIndex:     channelCache.FundingOutputIndex,
				FundingBlockHeight:     channelCache.FundingBlockHeight,
				FundedOn:               channelCache.FundedOn,
				Capacity:               channelCache.Capacity,
				FirstNodeId:            channelCache.FirstNodeId,
				SecondNodeId:           channelCache.SecondNodeId,
				InitiatingNodeId:       channelCache.InitiatingNodeId,
				AcceptingNodeId:        channelCache.AcceptingNodeId,
				Private:                channelCache.Private,
				ClosingTransactionHash: channelCache.ClosingTransactionHash,
				ClosingNodeId:          channelCache.ClosingNodeId,
				ClosingBlockHeight:     channelCache.ClosingBlockHeight,
				ClosedOn:               channelCache.ClosedOn,
				Flags:                  channelCache.Flags,
			}
		}
	case writeChannelStatusId:
		if channelCache.ChannelId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) allowed", channelCache.ChannelId)
		} else {
			settings := allChannelSettingsByChannelIdCache[channelId(channelCache.ChannelId)]
			if channelCache.Status >= core.CooperativeClosed {
				if settings.ShortChannelId != "" {
					delete(shortChannelIdCache, shortChannelId(settings.ShortChannelId))
				}
				delete(channelPointCache, createChannelPoint(settings.FundingTransactionHash, settings.FundingOutputIndex))
			}
			allChannelStatusCache[channelId(channelCache.ChannelId)] = channelCache.Status
			settings.Status = channelCache.Status
			allChannelSettingsByChannelIdCache[channelId(channelCache.ChannelId)] = settings
		}
	}
}

func parseAndConvertShortIds(channelCache ChannelCache) {
	if channelCache.ShortChannelId == "" && channelCache.LndShortChannelId != 0 {
		channelCache.ShortChannelId = core.ConvertLNDShortChannelID(channelCache.LndShortChannelId)
	}
	if channelCache.LndShortChannelId == 0 && channelCache.ShortChannelId != "" {
		var err error
		channelCache.LndShortChannelId, err = core.ConvertShortChannelIDToLND(channelCache.ShortChannelId)
		if err != nil {
			log.Error().Msgf("Could not convert ShortChannelId (%v) into LndShortChannelId", channelCache.ShortChannelId)
		}
	}
}

func createChannelPoint(fundingTransactionHash string, fundingOutputIndex int) channelPoint {
	return channelPoint(fmt.Sprintf("%s:%v", fundingTransactionHash, fundingOutputIndex))
}

func GetActiveChannelIdByFundingTransaction(fundingTransactionHash string, fundingOutputIndex int) int {
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Type:                   readActiveChannelIdByFundingTransaction,
		Out:                    channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByChannelPoint(channelPoint string) int {
	fundingTransactionHash, fundingOutputIndex := core.ParseChannelPoint(channelPoint)
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Type:                   readChannelIdByFundingTransaction,
		Out:                    channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByFundingTransaction(fundingTransactionHash string, fundingOutputIndex int) int {
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		FundingTransactionHash: fundingTransactionHash,
		FundingOutputIndex:     fundingOutputIndex,
		Type:                   readChannelIdByFundingTransaction,
		Out:                    channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetActiveChannelIdByShortChannelId(shortChannelId string) int {
	if shortChannelId == "" || shortChannelId == "0x0x0" {
		return 0
	}
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		ShortChannelId: shortChannelId,
		Type:           readActiveChannelIdByShortchannelId,
		Out:            channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByShortChannelId(shortChannelId string) int {
	if shortChannelId == "" || shortChannelId == "0x0x0" {
		return 0
	}
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		ShortChannelId: shortChannelId,
		Type:           readChannelIdByShortchannelId,
		Out:            channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdByLndShortChannelId(lndShortChannelId uint64) int {
	if lndShortChannelId == 0 {
		return 0
	}
	scId := core.ConvertLNDShortChannelID(lndShortChannelId)
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		ShortChannelId: scId,
		Type:           readChannelIdByShortchannelId,
		Out:            channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	channelResponse := <-channelResponseChannel
	return channelResponse.ChannelId
}

func GetChannelIdsByNodeId(peerNodeId int) []int {
	channelIdsResponseChannel := make(chan []int)
	channelCache := ChannelCache{
		NodeId:        peerNodeId,
		Type:          readChannelIdsByNodeId,
		ChannelIdsOut: channelIdsResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
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

func GetChannelStatusByChannelId(channelId int) core.ChannelStatus {
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		ChannelId: channelId,
		Type:      readStatusIdByChannelId,
		Out:       channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	channelResponse := <-channelResponseChannel
	return channelResponse.Status
}

func GetChannelSettingsByNodeId(nodeId int) []ChannelSettingsCache {
	channelResponseChannel := make(chan []ChannelSettingsCache)
	channelCache := ChannelCache{
		NodeId:             nodeId,
		Type:               readAllChannelSettings,
		ChannelSettingsOut: channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	return <-channelResponseChannel
}

func GetChannelSettingByChannelId(channelId int) ChannelSettingsCache {
	channelResponseChannel := make(chan ChannelSettingsCache)
	channelCache := ChannelCache{
		ChannelId:         channelId,
		Type:              readChannelSettings,
		ChannelSettingOut: channelResponseChannel,
	}
	ChannelsCacheChannel <- channelCache
	return <-channelResponseChannel
}

func SetChannel(channelId int, shortChannelId *string, lndShortChannelId *uint64, status core.ChannelStatus,
	fundingTransactionHash string, fundingOutputIndex int,
	fundingBlockHeight *uint32, fundedOn *time.Time,
	capacity int64, private bool, firstNodeId int, secondNodeId int,
	initiatingNodeId *int, acceptingNodeId *int,
	closingTransactionHash *string, closingNodeId *int, closingBlockHeight *uint32, closedOn *time.Time,
	flags core.ChannelFlags) {

	channelCache := ChannelCache{
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
		Type:                   writeChannel,
	}
	if shortChannelId != nil && *shortChannelId != "" {
		channelCache.ShortChannelId = *shortChannelId
	}
	if lndShortChannelId != nil && *lndShortChannelId != 0 {
		channelCache.LndShortChannelId = *lndShortChannelId
	}
	ChannelsCacheChannel <- channelCache
}

func SetChannelStatus(channelId int, status core.ChannelStatus) {
	channelCache := ChannelCache{
		ChannelId: channelId,
		Status:    status,
		Type:      writeChannelStatusId,
	}
	ChannelsCacheChannel <- channelCache
}
