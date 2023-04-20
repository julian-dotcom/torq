package tags

import (
	"context"

	"github.com/rs/zerolog/log"
)

var TagsCacheChannel = make(chan TagCache) //nolint:gochecknoglobals

type TagCacheOperationType uint
type tagIdType int

const (
	readTag TagCacheOperationType = iota
	readTags
	writeTag
	removeTag
)

type TagCache struct {
	Type    TagCacheOperationType
	TagId   int
	TagIds  []int
	Tag     Tag
	Out     chan<- Tag
	TagsOut chan<- []Tag
}

func TagsCacheHandler(ch <-chan TagCache, ctx context.Context) {
	tagsByIdCache := make(map[tagIdType]Tag, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case tagCache := <-ch:
			handleTagOperation(tagCache, tagsByIdCache)
		}
	}
}

func handleTagOperation(tagCache TagCache, tagsByIdCache map[tagIdType]Tag) {
	switch tagCache.Type {
	case readTag:
		if tagCache.TagId == 0 {
			tagCache.Out <- Tag{}
			close(tagCache.Out)
			break
		}
		tag, exists := tagsByIdCache[tagIdType(tagCache.TagId)]
		if exists {
			tagCache.Out <- tag
			close(tagCache.Out)
			break
		}
		tagCache.Out <- Tag{}
		close(tagCache.Out)
	case readTags:
		if len(tagCache.TagIds) == 0 {
			tagCache.TagsOut <- []Tag{}
			close(tagCache.TagsOut)
			break
		}
		var tags []Tag
		for _, tId := range tagCache.TagIds {
			tag, exists := tagsByIdCache[tagIdType(tId)]
			if exists {
				tags = append(tags, tag)
			}
		}
		tagCache.TagsOut <- tags
		close(tagCache.TagsOut)
	case writeTag:
		if tagCache.Tag.TagId == 0 {
			log.Error().Msgf("No empty Tag.TagId allowed")
		} else {
			tagsByIdCache[tagIdType(tagCache.Tag.TagId)] = tagCache.Tag
		}
	case removeTag:
		if tagCache.TagId == 0 && len(tagCache.TagIds) == 0 {
			log.Error().Msgf("No empty TagId and TagIds allowed")
		} else {
			if tagCache.TagId != 0 {
				delete(tagsByIdCache, tagIdType(tagCache.TagId))
			}
			if len(tagCache.TagIds) != 0 {
				for _, tId := range tagCache.TagIds {
					delete(tagsByIdCache, tagIdType(tId))
				}
			}
		}
	}
}

func GetTagsByTagIds(tagIds []int) []Tag {
	tagsResponseChannel := make(chan []Tag)
	tagCache := TagCache{
		TagIds:  tagIds,
		Type:    readTags,
		TagsOut: tagsResponseChannel,
	}
	TagsCacheChannel <- tagCache
	return <-tagsResponseChannel
}

func GetTagByTagId(tagId int) Tag {
	tagResponseChannel := make(chan Tag)
	tagCache := TagCache{
		TagId: tagId,
		Type:  readTags,
		Out:   tagResponseChannel,
	}
	TagsCacheChannel <- tagCache
	return <-tagResponseChannel
}

func SetTag(tag Tag) {
	tagCache := TagCache{
		Tag:  tag,
		Type: writeTag,
	}
	TagsCacheChannel <- tagCache
}

func RemoveTag(tagId int) {
	tagCache := TagCache{
		TagId: tagId,
		Type:  removeTag,
	}
	TagsCacheChannel <- tagCache
}
