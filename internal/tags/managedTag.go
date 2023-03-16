package tags

import (
	"context"

	"github.com/rs/zerolog/log"
)

var ManagedTagChannel = make(chan ManagedTag) //nolint:gochecknoglobals

type ManagedTagCacheOperationType uint

const (
	readTag ManagedTagCacheOperationType = iota
	readTags
	writeTag
	removeTag
)

type ManagedTag struct {
	Type    ManagedTagCacheOperationType
	TagId   int
	TagIds  []int
	Tag     Tag
	Out     chan<- Tag
	TagsOut chan<- []Tag
}

func ManagedTagCache(ch <-chan ManagedTag, ctx context.Context) {
	tagsByIdCache := make(map[int]Tag, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedManagedTag := <-ch:
			processManagedTag(managedManagedTag, tagsByIdCache)
		}
	}
}

func processManagedTag(managedTag ManagedTag, tagsByIdCache map[int]Tag) {
	switch managedTag.Type {
	case readTag:
		if managedTag.TagId == 0 {
			SendToManagedTagChannel(managedTag.Out, Tag{})
			break
		}
		tag, exists := tagsByIdCache[managedTag.TagId]
		if exists {
			SendToManagedTagChannel(managedTag.Out, tag)
			break
		}
		SendToManagedTagChannel(managedTag.Out, Tag{})
	case readTags:
		if len(managedTag.TagIds) == 0 {
			SendToManagedTagsChannel(managedTag.TagsOut, []Tag{})
			break
		}
		var tags []Tag
		for _, tagId := range managedTag.TagIds {
			tag, exists := tagsByIdCache[tagId]
			if exists {
				tags = append(tags, tag)
			}
		}
		SendToManagedTagsChannel(managedTag.TagsOut, tags)
	case writeTag:
		if managedTag.Tag.TagId == 0 {
			log.Error().Msgf("No empty Tag.TagId allowed")
		} else {
			tagsByIdCache[managedTag.Tag.TagId] = managedTag.Tag
		}
	case removeTag:
		if managedTag.TagId == 0 && len(managedTag.TagIds) == 0 {
			log.Error().Msgf("No empty TagId and TagIds allowed")
		} else {
			if managedTag.TagId != 0 {
				delete(tagsByIdCache, managedTag.TagId)
			}
			if len(managedTag.TagIds) != 0 {
				for _, tagId := range managedTag.TagIds {
					delete(tagsByIdCache, tagId)
				}
			}
		}
	}
}

func SendToManagedTagChannel(ch chan<- Tag, tag Tag) {
	ch <- tag
}

func SendToManagedTagsChannel(ch chan<- []Tag, tags []Tag) {
	ch <- tags
}

func GetTagsByTagIds(tagIds []int) []Tag {
	tagsResponseChannel := make(chan []Tag)
	managedManagedTag := ManagedTag{
		TagIds:  tagIds,
		Type:    readTags,
		TagsOut: tagsResponseChannel,
	}
	ManagedTagChannel <- managedManagedTag
	return <-tagsResponseChannel
}

func GetTagByTagId(tagId int) Tag {
	tagResponseChannel := make(chan Tag)
	managedManagedTag := ManagedTag{
		TagId: tagId,
		Type:  readTags,
		Out:   tagResponseChannel,
	}
	ManagedTagChannel <- managedManagedTag
	return <-tagResponseChannel
}

func SetTag(tag Tag) {
	managedManagedTag := ManagedTag{
		Tag:  tag,
		Type: writeTag,
	}
	ManagedTagChannel <- managedManagedTag
}

func RemoveTag(tagId int) {
	managedManagedTag := ManagedTag{
		TagId: tagId,
		Type:  removeTag,
	}
	ManagedTagChannel <- managedManagedTag
}
