package cache

import (
	"context"
	"sort"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

var TaggedCacheChannel = make(chan TaggedCache) //nolint:gochecknoglobals

type TaggedCacheOperationType uint
type tagIdType int

const (
	readTagged TaggedCacheOperationType = iota
	readTaggedNodes
	readTaggedChannels
	writeTagged
	addTagged
	removeTagged
)

type TaggedCache struct {
	Type      TaggedCacheOperationType
	NodeId    int
	ChannelId int
	TagId     int
	TagIds    []int
	Out       chan<- []int
}

func TaggedCacheHandler(ch <-chan TaggedCache, ctx context.Context) {
	tagsByNodeIdCache := make(map[nodeIdType]map[tagIdType]bool, 0)
	tagsByChannelIdCache := make(map[channelIdType]map[tagIdType]bool, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case taggedCache := <-ch:
			handleTaggedOperation(taggedCache, tagsByNodeIdCache, tagsByChannelIdCache)
		}
	}
}

func handleTaggedOperation(taggedCache TaggedCache,
	tagsByNodeIdCache map[nodeIdType]map[tagIdType]bool,
	tagsByChannelIdCache map[channelIdType]map[tagIdType]bool) {

	switch taggedCache.Type {
	case readTagged:
		if taggedCache.NodeId == 0 && taggedCache.ChannelId == 0 {
			log.Error().Msgf("No empty NodeId and ChannelId allowed")
			taggedCache.Out <- []int{}
			break
		}
		var tagIds []int
		if taggedCache.ChannelId != 0 {
			channelTagIds, exists := tagsByChannelIdCache[channelIdType(taggedCache.ChannelId)]
			if exists {
				for channelTagId := range channelTagIds {
					tagIds = append(tagIds, int(channelTagId))
				}
			}
		}
		if taggedCache.NodeId != 0 {
			nodeTagIds, exists := tagsByNodeIdCache[nodeIdType(taggedCache.NodeId)]
			if exists {
				for nodeTagId := range nodeTagIds {
					if !slices.Contains(tagIds, int(nodeTagId)) {
						tagIds = append(tagIds, int(nodeTagId))
					}
				}
			}
		}
		sort.Ints(tagIds)
		taggedCache.Out <- tagIds
	case readTaggedNodes:
		if taggedCache.TagId == 0 {
			log.Error().Msgf("No empty TagId allowed")
			taggedCache.Out <- []int{}
			break
		}
		var nodeIds []int
		for nodeId, tagIds := range tagsByNodeIdCache {
			_, exists := tagIds[tagIdType(taggedCache.TagId)]
			if exists {
				nodeIds = append(nodeIds, int(nodeId))
			}
		}
		sort.Ints(nodeIds)
		taggedCache.Out <- nodeIds
	case readTaggedChannels:
		if taggedCache.TagId == 0 {
			log.Error().Msgf("No empty TagId allowed")
			taggedCache.Out <- []int{}
			break
		}
		var channelIds []int
		for cId, tagIds := range tagsByChannelIdCache {
			_, exists := tagIds[tagIdType(taggedCache.TagId)]
			if exists {
				channelIds = append(channelIds, int(cId))
			}
		}
		sort.Ints(channelIds)
		taggedCache.Out <- channelIds
	case writeTagged:
		if taggedCache.NodeId == 0 && taggedCache.ChannelId == 0 {
			log.Error().Msgf("No empty NodeId and ChannelId allowed")
		} else {
			if len(taggedCache.TagIds) == 0 {
				break
			}
			if taggedCache.NodeId != 0 {
				tagIds := make(map[tagIdType]bool)
				for _, tId := range taggedCache.TagIds {
					tagIds[tagIdType(tId)] = true
				}
				tagsByNodeIdCache[nodeIdType(taggedCache.NodeId)] = tagIds
			}
			if taggedCache.ChannelId != 0 {
				tagIds := make(map[tagIdType]bool)
				for _, tId := range taggedCache.TagIds {
					tagIds[tagIdType(tId)] = true
				}
				tagsByChannelIdCache[channelIdType(taggedCache.ChannelId)] = tagIds
			}
		}
	case addTagged:
		if (taggedCache.NodeId == 0 && taggedCache.ChannelId == 0) || taggedCache.TagId == 0 {
			log.Error().Msgf("No empty NodeId (%v) and ChannelId (%v) or TagId (%v) allowed",
				taggedCache.NodeId, taggedCache.ChannelId, taggedCache.TagId)
		} else {
			if taggedCache.NodeId != 0 {
				tagIds, exists := tagsByNodeIdCache[nodeIdType(taggedCache.NodeId)]
				if exists {
					_, exists := tagIds[tagIdType(taggedCache.TagId)]
					if !exists {
						tagIds[tagIdType(taggedCache.TagId)] = true
						tagsByNodeIdCache[nodeIdType(taggedCache.NodeId)] = tagIds
					}
					break
				}
				tagsByNodeIdCache[nodeIdType(taggedCache.NodeId)] = map[tagIdType]bool{tagIdType(taggedCache.TagId): true}
			}
			if taggedCache.ChannelId != 0 {
				tagIds, exists := tagsByChannelIdCache[channelIdType(taggedCache.ChannelId)]
				if exists {
					_, exists := tagIds[tagIdType(taggedCache.TagId)]
					if !exists {
						tagIds[tagIdType(taggedCache.TagId)] = true
						tagsByChannelIdCache[channelIdType(taggedCache.ChannelId)] = tagIds
					}
					break
				}
				tagsByChannelIdCache[channelIdType(taggedCache.ChannelId)] = map[tagIdType]bool{tagIdType(taggedCache.TagId): true}
			}
		}
	case removeTagged:
		if (taggedCache.NodeId == 0 && taggedCache.ChannelId == 0) || taggedCache.TagId == 0 {
			log.Error().Msgf("No empty NodeId (%v) and ChannelId (%v) or TagId (%v) allowed",
				taggedCache.NodeId, taggedCache.ChannelId, taggedCache.TagId)
		} else {
			if taggedCache.NodeId != 0 {
				tagIds, exists := tagsByNodeIdCache[nodeIdType(taggedCache.NodeId)]
				if exists {
					delete(tagIds, tagIdType(taggedCache.TagId))
					tagsByNodeIdCache[nodeIdType(taggedCache.NodeId)] = tagIds
				}
			}
			if taggedCache.ChannelId != 0 {
				tagIds, exists := tagsByChannelIdCache[channelIdType(taggedCache.ChannelId)]
				if exists {
					delete(tagIds, tagIdType(taggedCache.TagId))
					tagsByChannelIdCache[channelIdType(taggedCache.ChannelId)] = tagIds
				}
			}
		}
	}
}

func GetNodeIdsByTagId(tagId int) []int {
	nodeIdsResponseChannel := make(chan []int)
	taggedCache := TaggedCache{
		TagId: tagId,
		Type:  readTaggedNodes,
		Out:   nodeIdsResponseChannel,
	}
	TaggedCacheChannel <- taggedCache
	return <-nodeIdsResponseChannel
}

func GetChannelIdsByTagId(tagId int) []int {
	channelIdsResponseChannel := make(chan []int)
	taggedCache := TaggedCache{
		TagId: tagId,
		Type:  readTaggedChannels,
		Out:   channelIdsResponseChannel,
	}
	TaggedCacheChannel <- taggedCache
	return <-channelIdsResponseChannel
}

func GetTagIdsByNodeId(nodeId int) []int {
	tagIdsResponseChannel := make(chan []int)
	taggedCache := TaggedCache{
		NodeId: nodeId,
		Type:   readTagged,
		Out:    tagIdsResponseChannel,
	}
	TaggedCacheChannel <- taggedCache
	return <-tagIdsResponseChannel
}

func GetTagIdsByChannelId(nodeId int, channelId int) []int {
	tagIdsResponseChannel := make(chan []int)
	taggedCache := TaggedCache{
		ChannelId: channelId,
		NodeId:    nodeId,
		Type:      readTagged,
		Out:       tagIdsResponseChannel,
	}
	TaggedCacheChannel <- taggedCache
	return <-tagIdsResponseChannel
}

func AddTagIdByNodeId(nodeId int, tagId int) {
	taggedCache := TaggedCache{
		TagId:  tagId,
		NodeId: nodeId,
		Type:   addTagged,
	}
	TaggedCacheChannel <- taggedCache
}

func AddTagIdByChannelId(channelId int, tagId int) {
	taggedCache := TaggedCache{
		TagId:     tagId,
		ChannelId: channelId,
		Type:      addTagged,
	}
	TaggedCacheChannel <- taggedCache
}

func RemoveTagIdByNodeId(nodeId int, tagId int) {
	taggedCache := TaggedCache{
		TagId:  tagId,
		NodeId: nodeId,
		Type:   removeTagged,
	}
	TaggedCacheChannel <- taggedCache
}

func RemoveTagIdByChannelId(channelId int, tagId int) {
	taggedCache := TaggedCache{
		TagId:     tagId,
		ChannelId: channelId,
		Type:      removeTagged,
	}
	TaggedCacheChannel <- taggedCache
}

func SetTagIdsByNodeId(nodeId int, tagIds []int) {
	taggedCache := TaggedCache{
		TagIds: tagIds,
		NodeId: nodeId,
		Type:   writeTagged,
	}
	TaggedCacheChannel <- taggedCache
}

func SetTagIdsByChannelId(channelId int, tagIds []int) {
	taggedCache := TaggedCache{
		TagIds:    tagIds,
		ChannelId: channelId,
		Type:      writeTagged,
	}
	TaggedCacheChannel <- taggedCache
}
