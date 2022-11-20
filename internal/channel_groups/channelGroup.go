package channel_groups

import (
	"time"
)

type channelGroup struct {
	ChannelGroupId int       `json:"channelGroupId" db:"channel_group_id"`
	GroupOriginId  int       `json:"groupOriginId" db:"group_origin_id"`
	NodeId         int       `json:"nodeId" db:"node_id"`
	CategoryId     *int      `json:"categoryId" db:"category_id"`
	TagId          *int      `json:"tagId" db:"tag_id"`
	ChannelId      int       `json:"channelId" db:"channel_id"`
	CreatedOn      time.Time `json:"createdOn" db:"created_on"`
	// No UpdateOn as there will never be an update always create/delete.
}

type groupOrigin int

const (
	corridorOrigin = groupOrigin(iota)
)
