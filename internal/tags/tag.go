package tags

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type tag struct {
	TagID       int       `json:"tagId" db:"tag_id"`
	ChannelDBID int       `json:"channelDBId" db:"channel_db_id"`
	Tag         string    `json:"tag" db:"tag"`
	CreatedOn   time.Time `json:"createdOn" db:"created_on"`
	UpdateOn    null.Time `json:"updatedOn" db:"updated_on"`
}
