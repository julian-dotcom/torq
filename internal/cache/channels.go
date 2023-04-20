package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/core"
)

var ChannelsCacheChannel = make(chan ChannelCache) //nolint:gochecknoglobals

type ChannelCacheOperationType uint
type channelIdType int
type channelPointType string
type shortChannelIdType string

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
	ShortChannelId         *string
	LndShortChannelId      *uint64
	FundingTransactionHash *string
	FundingOutputIndex     *int
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
	ShortChannelId         *string
	LndShortChannelId      *uint64
	FundingTransactionHash *string
	FundingOutputIndex     *int
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
	allChannelSettingsByChannelIdCache := make(map[channelIdType]ChannelSettingsCache, 0)
	shortChannelIdCache := make(map[shortChannelIdType]channelIdType, 0)
	allShortChannelIdCache := make(map[shortChannelIdType]channelIdType, 0)
	channelPointCache := make(map[channelPointType]channelIdType, 0)
	allChannelPointCache := make(map[channelPointType]channelIdType, 0)
	allChannelStatusCache := make(map[channelIdType]core.ChannelStatus, 0)
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
	shortChannelIdCache map[shortChannelIdType]channelIdType, allShortChannelIdCache map[shortChannelIdType]channelIdType,
	channelPointCache map[channelPointType]channelIdType, allChannelPointCache map[channelPointType]channelIdType,
	allChannelStatusCache map[channelIdType]core.ChannelStatus, allChannelSettingsByChannelIdCache map[channelIdType]ChannelSettingsCache) {
	switch channelCache.Type {
	case readActiveChannelIdByShortchannelId:
		if channelCache.ShortChannelId == nil || *channelCache.ShortChannelId == "" {
			channelCache.ChannelId = 0
		} else {
			channelCache.ChannelId = int(shortChannelIdCache[shortChannelIdType(*channelCache.ShortChannelId)])
		}
		channelCache.Out <- channelCache
	case readChannelIdByShortchannelId:
		if channelCache.ShortChannelId == nil || *channelCache.ShortChannelId == "" {
			channelCache.ChannelId = 0
		} else {
			channelCache.ChannelId = int(allShortChannelIdCache[shortChannelIdType(*channelCache.ShortChannelId)])
		}
		channelCache.Out <- channelCache
	case readActiveChannelIdByFundingTransaction:
		channelPoint, err := createChannelPoint(channelCache.FundingTransactionHash, channelCache.FundingOutputIndex)
		if err != nil {
			channelCache.ChannelId = 0
		} else {
			channelCache.ChannelId = int(channelPointCache[channelPoint])
		}
		channelCache.Out <- channelCache
	case readChannelIdByFundingTransaction:
		channelPoint, err := createChannelPoint(channelCache.FundingTransactionHash, channelCache.FundingOutputIndex)
		if err != nil {
			channelCache.ChannelId = 0
		} else {
			channelCache.ChannelId = int(allChannelPointCache[channelPoint])
		}
		channelCache.Out <- channelCache
	case readStatusIdByChannelId:
		channelCache.Status = allChannelStatusCache[channelIdType(channelCache.ChannelId)]
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
		channelCache.ChannelSettingOut <- allChannelSettingsByChannelIdCache[channelIdType(channelCache.ChannelId)]
	case writeChannel:
		if channelCache.ChannelId == 0 {
			log.Error().Msgf("No empty ChannelId allowed")
			break
		}
		cp, err := createChannelPoint(channelCache.FundingTransactionHash, channelCache.FundingOutputIndex)
		if err != nil {
			cp = ""
		}
		parseAndConvertShortIds(channelCache)
		if channelCache.Status < core.CooperativeClosed {
			if channelCache.ShortChannelId != nil && *channelCache.ShortChannelId != "" {
				shortChannelIdCache[shortChannelIdType(*channelCache.ShortChannelId)] = channelIdType(channelCache.ChannelId)
			}
			if cp != "" {
				channelPointCache[cp] = channelIdType(channelCache.ChannelId)
			}
		} else {
			if channelCache.ShortChannelId != nil && *channelCache.ShortChannelId != "" {
				delete(shortChannelIdCache, shortChannelIdType(*channelCache.ShortChannelId))
			}
			if cp != "" {
				delete(channelPointCache, cp)
			}
		}
		if channelCache.ShortChannelId != nil && *channelCache.ShortChannelId != "" {
			allShortChannelIdCache[shortChannelIdType(*channelCache.ShortChannelId)] = channelIdType(channelCache.ChannelId)
		}
		if cp != "" {
			allChannelPointCache[cp] = channelIdType(channelCache.ChannelId)
		}
		allChannelStatusCache[channelIdType(channelCache.ChannelId)] = channelCache.Status
		allChannelSettingsByChannelIdCache[channelIdType(channelCache.ChannelId)] = ChannelSettingsCache{
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
	case writeChannelStatusId:
		if channelCache.ChannelId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) allowed", channelCache.ChannelId)
			break
		}
		settings := allChannelSettingsByChannelIdCache[channelIdType(channelCache.ChannelId)]
		if channelCache.Status >= core.CooperativeClosed {
			if settings.ShortChannelId != nil || *settings.ShortChannelId != "" {
				delete(shortChannelIdCache, shortChannelIdType(*settings.ShortChannelId))
			}
			cp, err := createChannelPoint(settings.FundingTransactionHash, settings.FundingOutputIndex)
			if err == nil {
				delete(channelPointCache, cp)
			}
		}
		allChannelStatusCache[channelIdType(channelCache.ChannelId)] = channelCache.Status
		settings.Status = channelCache.Status
		allChannelSettingsByChannelIdCache[channelIdType(channelCache.ChannelId)] = settings
	}
}

func parseAndConvertShortIds(channelCache ChannelCache) {
	if (channelCache.ShortChannelId == nil || *channelCache.ShortChannelId == "") &&
		channelCache.LndShortChannelId != nil && *channelCache.LndShortChannelId != 0 {
		scId := core.ConvertLNDShortChannelID(*channelCache.LndShortChannelId)
		channelCache.ShortChannelId = &scId
	}
	if (channelCache.LndShortChannelId == nil || *channelCache.LndShortChannelId == 0) &&
		channelCache.ShortChannelId != nil && *channelCache.ShortChannelId != "" && *channelCache.ShortChannelId != "0x0x0" {
		scId, err := core.ConvertShortChannelIDToLND(*channelCache.ShortChannelId)
		if err != nil {
			log.Error().Msgf("Could not convert ShortChannelId (%v) into LndShortChannelId", channelCache.ShortChannelId)
			return
		}
		channelCache.LndShortChannelId = &scId
	}
}

func createChannelPoint(fundingTransactionHash *string, fundingOutputIndex *int) (channelPointType, error) {
	if fundingTransactionHash == nil || *fundingTransactionHash == "" || fundingOutputIndex == nil {
		return "", errors.New("Could not find funding transaction information")
	}
	return channelPointType(fmt.Sprintf("%s:%v", fundingTransactionHash, fundingOutputIndex)), nil
}

func GetActiveChannelIdByFundingTransaction(fundingTransactionHash *string, fundingOutputIndex *int) int {
	if fundingTransactionHash == nil || *fundingTransactionHash == "" || fundingOutputIndex == nil {
		return 0
	}
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

func GetChannelIdByFundingTransaction(fundingTransactionHash *string, fundingOutputIndex *int) int {
	if fundingTransactionHash == nil || *fundingTransactionHash == "" || fundingOutputIndex == nil {
		return 0
	}
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

func GetActiveChannelIdByShortChannelId(shortChannelId *string) int {
	if shortChannelId == nil || *shortChannelId == "" || *shortChannelId == "0x0x0" {
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

func GetChannelIdByShortChannelId(shortChannelId *string) int {
	if shortChannelId == nil || *shortChannelId == "" || *shortChannelId == "0x0x0" {
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

func GetChannelIdByLndShortChannelId(lndShortChannelId *uint64) int {
	if lndShortChannelId == nil || *lndShortChannelId == 0 {
		return 0
	}
	scId := core.ConvertLNDShortChannelID(*lndShortChannelId)
	channelResponseChannel := make(chan ChannelCache)
	channelCache := ChannelCache{
		ShortChannelId: &scId,
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
	fundingTransactionHash *string, fundingOutputIndex *int,
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
		channelCache.ShortChannelId = shortChannelId
	}
	if lndShortChannelId != nil && *lndShortChannelId != 0 {
		channelCache.LndShortChannelId = lndShortChannelId
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
