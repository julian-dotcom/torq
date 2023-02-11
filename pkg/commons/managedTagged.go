package commons

import (
	"context"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

var ManagedTaggedChannel = make(chan ManagedTagged) //nolint:gochecknoglobals

type ManagedTaggedCacheOperationType uint

const (
	READ_TAGGED ManagedTaggedCacheOperationType = iota
	READ_TAGGED_NODES
	READ_TAGGED_CHANNELS
	WRITE_TAGGED
	ADD_TAGGED
	REMOVE_TAGGED
)

type ManagedTagged struct {
	Type      ManagedTaggedCacheOperationType
	NodeId    int
	ChannelId int
	TagId     int
	TagIds    []int
	Out       chan []int
}

func ManagedTaggedCache(ch chan ManagedTagged, ctx context.Context) {
	tagsByNodeIdCache := make(map[int]map[int]bool, 0)
	tagsByChannelIdCache := make(map[int]map[int]bool, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedManagedTagged := <-ch:
			processManagedTagged(managedManagedTagged, tagsByNodeIdCache, tagsByChannelIdCache)
		}
	}
}

func processManagedTagged(managedTagged ManagedTagged, tagsByNodeIdCache map[int]map[int]bool, tagsByChannelIdCache map[int]map[int]bool) {
	switch managedTagged.Type {
	case READ_TAGGED:
		if managedTagged.NodeId == 0 && managedTagged.ChannelId == 0 {
			log.Error().Msgf("No empty NodeId and ChannelId allowed")
			SendToManagedTagIdsChannel(managedTagged.Out, []int{})
			break
		}
		var tagIds []int
		if managedTagged.ChannelId != 0 {
			channelTagIds, exists := tagsByChannelIdCache[managedTagged.ChannelId]
			if exists {
				for channelTagId := range channelTagIds {
					tagIds = append(tagIds, channelTagId)
				}
			}
		}
		if managedTagged.NodeId != 0 {
			nodeTagIds, exists := tagsByNodeIdCache[managedTagged.NodeId]
			if exists {
				for nodeTagId := range nodeTagIds {
					if !slices.Contains(tagIds, nodeTagId) {
						tagIds = append(tagIds, nodeTagId)
					}
				}
			}
		}
		SendToManagedTagIdsChannel(managedTagged.Out, tagIds)
	case READ_TAGGED_NODES:
		if managedTagged.TagId == 0 {
			log.Error().Msgf("No empty TagId allowed")
			SendToManagedNodeIdsChannel(managedTagged.Out, []int{})
			break
		}
		var nodeIds []int
		for nodeId, tagIds := range tagsByNodeIdCache {
			_, exists := tagIds[managedTagged.TagId]
			if exists {
				nodeIds = append(nodeIds, nodeId)
			}
		}
		SendToManagedNodeIdsChannel(managedTagged.Out, nodeIds)
	case READ_TAGGED_CHANNELS:
		if managedTagged.TagId == 0 {
			log.Error().Msgf("No empty TagId allowed")
			SendToManagedChannelIdsChannel(managedTagged.Out, []int{})
			break
		}
		var channelIds []int
		for channelId, tagIds := range tagsByChannelIdCache {
			_, exists := tagIds[managedTagged.TagId]
			if exists {
				channelIds = append(channelIds, channelId)
			}
		}
		SendToManagedChannelIdsChannel(managedTagged.Out, channelIds)
	case WRITE_TAGGED:
		if managedTagged.NodeId == 0 && managedTagged.ChannelId == 0 {
			log.Error().Msgf("No empty NodeId and ChannelId allowed")
		} else {
			if len(managedTagged.TagIds) == 0 {
				break
			}
			if managedTagged.NodeId != 0 {
				tagIds := make(map[int]bool)
				for _, tagId := range managedTagged.TagIds {
					tagIds[tagId] = true
				}
				tagsByNodeIdCache[managedTagged.NodeId] = tagIds
			}
			if managedTagged.ChannelId != 0 {
				tagIds := make(map[int]bool)
				for _, tagId := range managedTagged.TagIds {
					tagIds[tagId] = true
				}
				tagsByChannelIdCache[managedTagged.ChannelId] = tagIds
			}
		}
	case ADD_TAGGED:
		if (managedTagged.NodeId == 0 && managedTagged.ChannelId == 0) || managedTagged.TagId == 0 {
			log.Error().Msgf("No empty NodeId (%v) and ChannelId (%v) or TagId (%v) allowed",
				managedTagged.NodeId, managedTagged.ChannelId, managedTagged.TagId)
		} else {
			if managedTagged.NodeId != 0 {
				tagIds, exists := tagsByNodeIdCache[managedTagged.NodeId]
				if exists {
					_, exists := tagIds[managedTagged.TagId]
					if !exists {
						tagIds[managedTagged.TagId] = true
						tagsByNodeIdCache[managedTagged.NodeId] = tagIds
					}
					break
				}
				tagsByNodeIdCache[managedTagged.NodeId] = map[int]bool{managedTagged.TagId: true}
			}
			if managedTagged.ChannelId != 0 {
				tagIds, exists := tagsByChannelIdCache[managedTagged.ChannelId]
				if exists {
					_, exists := tagIds[managedTagged.TagId]
					if !exists {
						tagIds[managedTagged.TagId] = true
						tagsByChannelIdCache[managedTagged.ChannelId] = tagIds
					}
					break
				}
				tagsByChannelIdCache[managedTagged.ChannelId] = map[int]bool{managedTagged.TagId: true}
			}
		}
	case REMOVE_TAGGED:
		if (managedTagged.NodeId == 0 && managedTagged.ChannelId == 0) || managedTagged.TagId == 0 {
			log.Error().Msgf("No empty NodeId (%v) and ChannelId (%v) or TagId (%v) allowed",
				managedTagged.NodeId, managedTagged.ChannelId, managedTagged.TagId)
		} else {
			if managedTagged.NodeId != 0 {
				tagIds, exists := tagsByNodeIdCache[managedTagged.NodeId]
				if exists {
					delete(tagIds, managedTagged.TagId)
					tagsByNodeIdCache[managedTagged.NodeId] = tagIds
				}
			}
			if managedTagged.ChannelId != 0 {
				tagIds, exists := tagsByChannelIdCache[managedTagged.ChannelId]
				if exists {
					delete(tagIds, managedTagged.TagId)
					tagsByChannelIdCache[managedTagged.ChannelId] = tagIds
				}
			}
		}
	}
}

func GetNodeIdsByTagId(tagId int) []int {
	nodeIdsResponseChannel := make(chan []int, 1)
	managedManagedTagged := ManagedTagged{
		TagId: tagId,
		Type:  READ_TAGGED_NODES,
		Out:   nodeIdsResponseChannel,
	}
	ManagedTaggedChannel <- managedManagedTagged
	return <-nodeIdsResponseChannel
}

func GetChannelIdsByTagId(tagId int) []int {
	channelIdsResponseChannel := make(chan []int, 1)
	managedManagedTagged := ManagedTagged{
		TagId: tagId,
		Type:  READ_TAGGED_CHANNELS,
		Out:   channelIdsResponseChannel,
	}
	ManagedTaggedChannel <- managedManagedTagged
	return <-channelIdsResponseChannel
}

func GetTagIdsByNodeId(nodeId int) []int {
	tagIdsResponseChannel := make(chan []int, 1)
	managedManagedTagged := ManagedTagged{
		NodeId: nodeId,
		Type:   READ_TAGGED,
		Out:    tagIdsResponseChannel,
	}
	ManagedTaggedChannel <- managedManagedTagged
	return <-tagIdsResponseChannel
}

func GetTagIdsByChannelId(nodeId int, channelId int) []int {
	tagIdsResponseChannel := make(chan []int, 1)
	managedManagedTagged := ManagedTagged{
		ChannelId: channelId,
		NodeId:    nodeId,
		Type:      READ_TAGGED,
		Out:       tagIdsResponseChannel,
	}
	ManagedTaggedChannel <- managedManagedTagged
	return <-tagIdsResponseChannel
}

func AddTagIdByNodeId(nodeId int, tagId int) {
	managedManagedTagged := ManagedTagged{
		TagId:  tagId,
		NodeId: nodeId,
		Type:   ADD_TAGGED,
	}
	ManagedTaggedChannel <- managedManagedTagged
}

func AddTagIdByChannelId(channelId int, tagId int) {
	managedManagedTagged := ManagedTagged{
		TagId:     tagId,
		ChannelId: channelId,
		Type:      ADD_TAGGED,
	}
	ManagedTaggedChannel <- managedManagedTagged
}

func RemoveTagIdByNodeId(nodeId int, tagId int) {
	managedManagedTagged := ManagedTagged{
		TagId:  tagId,
		NodeId: nodeId,
		Type:   REMOVE_TAGGED,
	}
	ManagedTaggedChannel <- managedManagedTagged
}

func RemoveTagIdByChannelId(channelId int, tagId int) {
	managedManagedTagged := ManagedTagged{
		TagId:     tagId,
		ChannelId: channelId,
		Type:      REMOVE_TAGGED,
	}
	ManagedTaggedChannel <- managedManagedTagged
}

func SetTagIdsByNodeId(nodeId int, tagIds []int) {
	managedManagedTagged := ManagedTagged{
		TagIds: tagIds,
		NodeId: nodeId,
		Type:   WRITE_TAGGED,
	}
	ManagedTaggedChannel <- managedManagedTagged
}

func SetTagIdsByChannelId(channelId int, tagIds []int) {
	managedManagedTagged := ManagedTagged{
		TagIds:    tagIds,
		ChannelId: channelId,
		Type:      WRITE_TAGGED,
	}
	ManagedTaggedChannel <- managedManagedTagged
}
