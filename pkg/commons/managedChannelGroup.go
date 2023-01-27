package commons

import (
	"context"

	"github.com/rs/zerolog/log"
)

type ChannelGroup struct {
	TagId         *int    `json:"tagId" db:"tag_id"`
	TagName       *string `json:"tagName" db:"tag_name"`
	TagStyle      *string `json:"tagStyle" db:"tag_style"`
	CategoryId    *int    `json:"categoryId" db:"category_id"`
	CategoryName  *string `json:"categoryName" db:"category_name"`
	CategoryStyle *string `json:"categoryStyle" db:"category_style"`
}

var ManagedChannelGroupChannel = make(chan ManagedChannelGroup) //nolint:gochecknoglobals

type ManagedChannelGroupCacheOperationType uint

const (
	// READ_CHANNELGROUPS please provide ChannelId and Out
	READ_CHANNELGROUPS ManagedChannelGroupCacheOperationType = iota
	// WRITE_CHANNELGROUPS Please provide ChannelId, ChannelGroups
	WRITE_CHANNELGROUPS
)

type ChannelGroupInclude uint

const (
	CATEGORIES_ONLY ChannelGroupInclude = iota
	DISTINCT_REGULAR_AND_TAG_CATEGORIES
	ALL_REGULAR_AND_TAG_CATEGORIES
	TAGS_ONLY
)

type ManagedChannelGroup struct {
	Type          ManagedChannelGroupCacheOperationType
	ChannelId     int
	Include       ChannelGroupInclude
	ChannelGroups []ChannelGroup
	Out           chan *ManagedChannelGroupSettings
}

type ManagedChannelGroupSettings struct {
	ChannelId     int
	ChannelGroups []ChannelGroup
}

func ManagedChannelGroupCache(ch chan ManagedChannelGroup, ctx context.Context) {
	channelGroupSettingsByChannelIdCache := make(map[int]ManagedChannelGroupSettings, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedChannelGroup := <-ch:
			processManagedChannelGroupSettings(managedChannelGroup, channelGroupSettingsByChannelIdCache)
		}
	}
}

func processManagedChannelGroupSettings(managedChannelGroup ManagedChannelGroup,
	channelGroupSettingsByChannelIdCache map[int]ManagedChannelGroupSettings) {
	switch managedChannelGroup.Type {
	case READ_CHANNELGROUPS:
		_, exists := channelGroupSettingsByChannelIdCache[managedChannelGroup.ChannelId]
		if !exists {
			SendToManagedChannelGroupSettingsChannel(managedChannelGroup.Out, nil)
			break
		}
		result := ManagedChannelGroupSettings{
			ChannelId:     managedChannelGroup.ChannelId,
			ChannelGroups: []ChannelGroup{},
		}
		if managedChannelGroup.Include == TAGS_ONLY {
			for _, channelGroup := range channelGroupSettingsByChannelIdCache[managedChannelGroup.ChannelId].ChannelGroups {
				if channelGroup.TagId != nil {
					result.ChannelGroups = append(result.ChannelGroups, channelGroup)
				}
			}
			SendToManagedChannelGroupSettingsChannel(managedChannelGroup.Out, &result)
			break
		}
		for _, channelGroup := range channelGroupSettingsByChannelIdCache[managedChannelGroup.ChannelId].ChannelGroups {
			if channelGroup.CategoryId != nil && channelGroup.TagId == nil {
				result.ChannelGroups = append(result.ChannelGroups, channelGroup)
			}
		}
		if managedChannelGroup.Include == CATEGORIES_ONLY {
			SendToManagedChannelGroupSettingsChannel(managedChannelGroup.Out, &result)
			break
		}
	group:
		for _, channelGroup := range channelGroupSettingsByChannelIdCache[managedChannelGroup.ChannelId].ChannelGroups {
			if channelGroup.CategoryId != nil && channelGroup.TagId != nil {
				if managedChannelGroup.Include == DISTINCT_REGULAR_AND_TAG_CATEGORIES {
					for _, existingChannelGroup := range result.ChannelGroups {
						if existingChannelGroup.CategoryId == channelGroup.CategoryId {
							continue group
						}
					}
				}
				result.ChannelGroups = append(result.ChannelGroups, channelGroup)
			}
		}
		SendToManagedChannelGroupSettingsChannel(managedChannelGroup.Out, &result)
	case WRITE_CHANNELGROUPS:
		if managedChannelGroup.ChannelId == 0 {
			log.Error().Msgf("No empty ChannelId (%v) allowed", managedChannelGroup.ChannelId)
		} else {
			channelGroupSettingsByChannelIdCache[managedChannelGroup.ChannelId] = ManagedChannelGroupSettings{
				ChannelId:     managedChannelGroup.ChannelId,
				ChannelGroups: managedChannelGroup.ChannelGroups,
			}
		}
	}
}

func GetChannelGroupsByChannelId(channelId int, include ChannelGroupInclude) *ManagedChannelGroupSettings {
	channelGroupResponseChannel := make(chan *ManagedChannelGroupSettings, 1)
	managedChannelGroup := ManagedChannelGroup{
		ChannelId: channelId,
		Include:   include,
		Type:      READ_CHANNELGROUPS,
		Out:       channelGroupResponseChannel,
	}
	ManagedChannelGroupChannel <- managedChannelGroup
	channelGroupResponse := <-channelGroupResponseChannel
	return channelGroupResponse
}

func SetChannelGroupsByChannelId(channelId int, channelGroups []ChannelGroup) {
	managedChannelGroup := ManagedChannelGroup{
		ChannelId:     channelId,
		ChannelGroups: channelGroups,
		Type:          WRITE_CHANNELGROUPS,
	}
	ManagedChannelGroupChannel <- managedChannelGroup
}
