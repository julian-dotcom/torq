package tags

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/cache"
	"github.com/lncapital/torq/pkg/core"
)

type Tag struct {
	TagId         int       `json:"tagId" db:"tag_id"`
	Name          string    `json:"name" db:"name"`
	Style         string    `json:"style" db:"style"`
	CategoryId    *int      `json:"categoryId" db:"category_id"`
	CategoryName  *string   `json:"categoryName" db:"category_name"`
	CategoryStyle *string   `json:"categoryStyle" db:"category_style"`
	CreatedOn     time.Time `json:"createdOn" db:"created_on"`
	UpdateOn      time.Time `json:"updatedOn" db:"updated_on"`
}

type TaggedNodes struct {
	Name               *string `db:"name" json:"name"`
	NodeId             int     `db:"node_id" json:"nodeId"`
	OpenChannelCount   int     `db:"open_channel_count" json:"openChannelCount"`
	ClosedChannelCount int     `db:"closed_channel_count" json:"closedChannelCount"`
}

type TaggedChannels struct {
	Name           *string `db:"name" json:"name"`
	ShortChannelId string  `db:"short_channel_id" json:"shortChannelId"`
	ChannelId      int     `db:"channel_id" json:"channelId"`
}

type TagResponse struct {
	Tag
	Channels []TaggedChannels `json:"channels"`
	Nodes    []TaggedNodes    `json:"nodes"`
}

func GetTagsByCategoryId(db *sqlx.DB, categoryId int) ([]TagResponse, error) {
	var tags []TagResponse
	err := db.Select(&tags, `SELECT tag.*, category.name as category_name, category.style as category_style FROM tag
		left JOIN category ON category.category_id = tag.category_id
		WHERE tag.category_id=$1;`, categoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TagResponse{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}

	for _, tag := range tags {
		tag.Channels, err = getTagChannels(tag.TagId)
		if err != nil {
			return []TagResponse{}, errors.Wrap(err, database.SqlExecutionError)
		}
		tag.Nodes, err = getTagNodes(db, tag.TagId)
		if err != nil {
			return []TagResponse{}, errors.Wrap(err, database.SqlExecutionError)
		}
	}

	return tags, nil
}

func GetTags(db *sqlx.DB) ([]TagResponse, error) {
	var tags []TagResponse
	err := db.Select(&tags, `
			SELECT tag.*, category.name as category_name, category.style as category_style FROM tag
			left JOIN category ON category.category_id = tag.category_id
			ORDER BY name ASC ;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TagResponse{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}

	for i, tag := range tags {
		tags[i].Channels, err = getTagChannels(tag.TagId)
		if err != nil {
			return []TagResponse{}, errors.Wrap(err, database.SqlExecutionError)
		}
		tags[i].Nodes, err = getTagNodes(db, tag.TagId)
		if err != nil {
			return []TagResponse{}, errors.Wrap(err, database.SqlExecutionError)
		}
	}

	return tags, nil
}

func InitializeTagsCache(db *sqlx.DB) error {
	var tags []Tag
	err := db.Select(&tags, `
			SELECT tag.*, category.name as category_name, category.style as category_style FROM tag
			left JOIN category ON category.category_id = tag.category_id
			ORDER BY name ASC ;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return errors.Wrap(err, database.SqlExecutionError)
	}
	for _, tag := range tags {
		SetTag(tag)
	}
	return nil
}

func GetTag(db *sqlx.DB, tagId int) (TagResponse, error) {
	var tag TagResponse
	err := db.Get(&tag, `SELECT tag.*, category.name as category_name, category.style  as category_style FROM tag
		left JOIN category ON category.category_id = tag.category_id
		WHERE tag_id=$1`, tagId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TagResponse{}, nil
		}
		return TagResponse{}, errors.Wrap(err, database.SqlExecutionError)
	}

	tag.Channels, err = getTagChannels(tagId)
	if err != nil {
		return TagResponse{}, errors.Wrap(err, database.SqlExecutionError)
	}
	tag.Nodes, err = getTagNodes(db, tagId)
	if err != nil {
		return TagResponse{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tag, nil
}

func createTag(db *sqlx.DB, tag Tag) (Tag, error) {
	tag.CreatedOn = time.Now().UTC()
	tag.UpdateOn = tag.CreatedOn
	err := db.QueryRowx(`INSERT INTO tag (name, style, created_on, updated_on, category_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING tag_id;`,
		tag.Name, tag.Style, tag.CreatedOn, tag.UpdateOn, tag.CategoryId).Scan(&tag.TagId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Tag{}, database.SqlUniqueConstraintError
			}
		}
		return Tag{}, errors.Wrap(err, database.SqlExecutionError)
	}
	SetTag(tag)
	return tag, nil
}

// updateTag you cannot update the category! If you want to reassign a tag to a new category you need to recreate the tag.
func updateTag(db *sqlx.DB, tag Tag) (Tag, error) {
	tag.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE tag SET name=$1, style=$2, updated_on=$3, category_id=$4 WHERE tag_id=$5;`,
		tag.Name, tag.Style, tag.UpdateOn, tag.CategoryId, tag.TagId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Tag{}, database.SqlUniqueConstraintError
			}
		}
		return Tag{}, errors.Wrap(err, database.SqlExecutionError)
	}
	SetTag(tag)
	return tag, nil
}

func deleteTag(db *sqlx.DB, tagId int) error {
	_, err := db.Exec(`DELETE FROM tag WHERE tag_id=$1;`, tagId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	RemoveTag(tagId)
	return nil
}

type TagEntityRequest struct {
	TagId                          int  `db:"tag_id"`
	ChannelId                      *int `db:"channel_id"`
	NodeId                         *int `db:"node_id"`
	CreatedByWorkflowVersionNodeId *int `db:"created_by_workflow_version_node_id"`
}

// tagEntity adds a tag to a node or channel
func TagEntity(db *sqlx.DB, req TagEntityRequest) (err error) {
	// Chack that either the channel or the node exists
	if req.ChannelId == nil && req.NodeId == nil {
		return errors.New("channel_id and node_id cannot both be nil")
	}

	if req.ChannelId != nil && req.NodeId != nil {
		return errors.New("channel_id and node_id cannot both be set")
	}

	createdOn := time.Now().UTC()
	if req.ChannelId != nil {
		_, err = db.Exec(
			`INSERT INTO tagged_entity (tag_id, channel_id, created_by_workflow_version_node_id, created_on)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ON CONSTRAINT unique_tagged_channel DO NOTHING;`, req.TagId, *req.ChannelId, req.CreatedByWorkflowVersionNodeId, createdOn)
		cache.AddTagIdByChannelId(*req.ChannelId, req.TagId)
	}

	if req.NodeId != nil {
		_, err = db.Exec(
			`INSERT INTO tagged_entity (tag_id, node_id, created_by_workflow_version_node_id, created_on)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ON CONSTRAINT unique_tagged_node DO NOTHING;`, req.TagId, *req.NodeId, req.CreatedByWorkflowVersionNodeId, createdOn)
		cache.AddTagIdByNodeId(*req.NodeId, req.TagId)
	}

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return database.SqlUniqueConstraintError
			}
		}
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func UntagEntity(db *sqlx.DB, req TagEntityRequest) (err error) {
	// Chack that either the channel or the node exists
	if req.ChannelId == nil && req.NodeId == nil {
		return errors.New("channel_id and node_id cannot both be nil")
	}

	if req.ChannelId != nil && req.NodeId != nil {
		return errors.New("channel_id and node_id cannot both be set")
	}

	if req.ChannelId != nil {
		_, err = db.Exec(`DELETE FROM tagged_entity WHERE tag_id=$1 AND channel_id=$2;`, req.TagId, *req.ChannelId)
		cache.RemoveTagIdByChannelId(*req.ChannelId, req.TagId)
	}

	if req.NodeId != nil {
		_, err = db.Exec(`DELETE FROM tagged_entity WHERE tag_id=$1 AND node_id=$2;`, req.TagId, *req.NodeId)
		cache.RemoveTagIdByNodeId(*req.NodeId, req.TagId)
	}

	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

type NodeTagsRequest struct {
	NodeId int `db:"node_id"`
}

type ChannelTagsRequest struct {
	ChannelId int  `db:"channel_id"`
	NodeId    *int `db:"node_id"`
}

// TODO FIXME FYI: second_node_id is not always remote_node_id
// Get all the channels belonging to a tag and inclide the node name based on the "second node id"
func getTagChannels(tagId int) ([]TaggedChannels, error) {
	var channels []TaggedChannels
	channelIds := cache.GetChannelIdsByTagId(tagId)
	for _, channelId := range channelIds {
		channelSettings := cache.GetChannelSettingByChannelId(channelId)
		nodeAlias := cache.GetNodeAlias(channelSettings.SecondNodeId)
		taggedChannel := TaggedChannels{
			ShortChannelId: channelSettings.ShortChannelId,
			ChannelId:      channelId,
		}
		if nodeAlias != "" {
			taggedChannel.Name = &nodeAlias
		}
		channels = append(channels, taggedChannel)
	}
	return channels, nil
}

// TODO FIXME FYI: second_node_id is not always remote_node_id
// Get all the Nodes belonging to a tag, include the number of channels belonging to each node
func getTagNodes(db *sqlx.DB, tagId int) ([]TaggedNodes, error) {
	var nodes []TaggedNodes
	nodeIds := cache.GetNodeIdsByTagId(tagId)
	for _, nodeId := range nodeIds {
		nodeAlias := cache.GetNodeAlias(nodeId)
		taggedNode := TaggedNodes{
			NodeId: nodeId,
		}
		if nodeAlias != "" {
			taggedNode.Name = &nodeAlias
		}

		var openChannelCount int
		err := db.Get(&openChannelCount, `
			SELECT COUNT(oc.channel_id)
			FROM node n
			JOIN channel oc ON oc.second_node_id = n.node_id AND oc.status_id = $2
			WHERE n.node_id=$1
			GROUP BY n.node_id;`, nodeId, core.Open)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Error().Err(err).Msgf("Could not obtain open channel count for tagId: %v and nodeId: %v", tagId, nodeId)
			}
		}
		taggedNode.OpenChannelCount = openChannelCount

		var closedChannelCount int
		err = db.Get(&closedChannelCount, `
			SELECT COUNT(noc.channel_id)
			FROM node n
			JOIN channel noc ON noc.second_node_id = n.node_id AND noc.status_id != $2
			WHERE n.node_id=$1
			GROUP BY n.node_id;`, nodeId, core.Open)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Error().Err(err).Msgf("Could not obtain closed channel count for tagId: %v and nodeId: %v", tagId, nodeId)
			}
		}
		taggedNode.ClosedChannelCount = closedChannelCount

		nodes = append(nodes, taggedNode)
	}
	return nodes, nil
}
