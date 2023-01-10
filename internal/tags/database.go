package tags

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/database"
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
	Name         *string `db:"name" json:"name"`
	NodeId       int     `db:"node_id" json:"nodeId"`
	ChannelCount int     `db:"channel_count" json:"channelCount"`
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
		tag.Channels, err = getTagChannels(db, tag.TagId)
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
			SELECT tag.*, category.name as category_name FROM tag
			left JOIN category ON category.category_id = tag.category_id
			ORDER BY name ASC ;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TagResponse{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}

	for _, tag := range tags {
		tag.Channels, err = getTagChannels(db, tag.TagId)
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

	tag.Channels, err = getTagChannels(db, tagId)
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
				return Tag{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return Tag{}, errors.Wrap(err, database.SqlExecutionError)
	}
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
				return Tag{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return Tag{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return tag, nil
}

func deleteTag(db *sqlx.DB, tagId int) error {
	_, err := db.Exec(`DELETE FROM tag WHERE tag_id=$1;`, tagId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

type TagEntityRequest struct {
	TagId     int  `db:"tag_id"`
	ChannelId *int `db:"channel_id"`
	NodeId    *int `db:"node_id"`
}

// tagEntity adds a tag to a node or channel
func tagEntity(db *sqlx.DB, req TagEntityRequest) (err error) {
	// Chack that either the channel or the node exists
	if req.ChannelId == nil && req.NodeId == nil {
		return errors.New("channel_id and node_id cannot both be nil")
	}

	if req.ChannelId != nil && req.NodeId != nil {
		return errors.New("channel_id and node_id cannot both be set")
	}

	if req.ChannelId != nil {
		_, err = db.Exec(`INSERT INTO tagged_entity (tag_id, channel_id) VALUES ($1, $2);`, req.TagId, *req.ChannelId)
	}

	if req.NodeId != nil {
		_, err = db.Exec(`INSERT INTO tagged_entity (tag_id, node_id) VALUES ($1, $2);`, req.TagId, *req.NodeId)
	}

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Constraint == "tagged_entity_tag_id_channel_id_key" {
				return errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func untagEntity(db *sqlx.DB, req TagEntityRequest) (err error) {
	// Chack that either the channel or the node exists
	if req.ChannelId == nil && req.NodeId == nil {
		return errors.New("channel_id and node_id cannot both be nil")
	}

	if req.ChannelId != nil && req.NodeId != nil {
		return errors.New("channel_id and node_id cannot both be set")
	}

	if req.ChannelId != nil {
		_, err = db.Exec(`DELETE FROM tagged_entity WHERE tag_id=$1 AND channel_id=$2;`, req.TagId, *req.ChannelId)
	}

	if req.NodeId != nil {
		_, err = db.Exec(`DELETE FROM tagged_entity WHERE tag_id=$1 AND node_id=$2;`, req.TagId, *req.NodeId)
	}

	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

type NodeTagsRequest struct {
	NodeId int `db:"node_id"`
}

// Get the tags for a node
func GetNodeTags(db *sqlx.DB, nodeId int) ([]Tag, error) {
	var tags []Tag
	err := db.Select(&tags, `
			SELECT tag.*, category.name as category_name FROM tag
			left JOIN category ON category.category_id = tag.category_id
			left JOIN tagged_entity ON tagged_entity.tag_id = tag.tag_id
			WHERE tagged_entity.node_id = $1
			ORDER BY name ASC ;`, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Tag{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return tags, nil
}

type ChannelTagsRequest struct {
	ChannelId int  `db:"channel_id"`
	NodeId    *int `db:"node_id"`
}

// GetChannelTags gets the tags for a channel, if both the channel id and node id is set it will also return the tags for the node
func GetChannelTags(db *sqlx.DB, req ChannelTagsRequest) ([]Tag, error) {
	var tags []Tag
	err := db.Select(&tags, `
				SELECT DISTINCT tag.tag_id, tag.*, category.name as category_name FROM tag
				left JOIN category ON category.category_id = tag.category_id
				left JOIN tagged_entity ON tagged_entity.tag_id = tag.tag_id
				WHERE tagged_entity.channel_id = $1
				OR ($2 IS NOT NULL AND tagged_entity.node_id = $2)
				ORDER BY name ASC ;`, req.ChannelId, req.NodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Tag{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return tags, nil
}

// Get all the channels belonging to a tag and inclide the node name based on the "second node id"
func getTagChannels(db *sqlx.DB, tagId int) ([]TaggedChannels, error) {
	var channels []TaggedChannels
	err := db.Select(&channels, `
				SELECT channel.channel_id, channel.short_channel_id, ne.name FROM channel
				LEFT JOIN tagged_entity ON tagged_entity.channel_id = channel.channel_id
				left JOIN (select last(alias, timestamp) as name, last(node_id, timestamp) as ni, event_node_id from node_event group by event_node_id) as ne ON ne.event_node_id = channel.second_node_id
				WHERE tagged_entity.tag_id = $1;`, tagId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TaggedChannels{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return channels, nil
}

// Get all the Nodes belonging to a tag, include the number of channels belonging to each node
func getTagNodes(db *sqlx.DB, tagId int) ([]TaggedNodes, error) {
	var nodes []TaggedNodes
	err := db.Select(&nodes, `
				SELECT ne.name, node.node_id, count(channel.channel_id) as channel_count FROM node
				left JOIN tagged_entity ON tagged_entity.node_id = node.node_id
				left JOIN channel ON channel.second_node_id = node.node_id
				left JOIN (select last(alias, timestamp) as name, last(node_id, timestamp) as ni, event_node_id from node_event group by event_node_id) as ne ON ne.event_node_id = channel.second_node_id
				WHERE tagged_entity.tag_id = $1
				GROUP BY node.node_id, ne.name;`, tagId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []TaggedNodes{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodes, nil
}
