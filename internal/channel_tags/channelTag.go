package channel_tags

import (
	"time"
)

type channelTag struct {
	ChannelTagId int       `json:"channelTagId" db:"channel_tag_id"`
	TagOriginId  int       `json:"tagOriginId" db:"tag_origin_id"`
	FromNodeId   int       `json:"fromNodeId" db:"from_node_id"`
	ToNodeId     int       `json:"toNodeId" db:"to_node_id"`
	ChannelId    int       `json:"channelId" db:"channel_id"`
	TagId        int       `json:"tagId" db:"tag_id"`
	CreatedOn    time.Time `json:"createdOn" db:"created_on"`
	// No UpdateOn as there will never be an update always create/delete.
}

type tagOrigin int

const (
	corridor = tagOrigin(iota)
)
