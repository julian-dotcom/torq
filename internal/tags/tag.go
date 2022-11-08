package tags

import (
	"time"
)

type Tag struct {
	TagId     int       `json:"tagId" db:"tag_id"`
	Name      string    `json:"name" db:"name"`
	Style     string    `json:"style" db:"style"`
	CreatedOn time.Time `json:"createdOn" db:"created_on"`
	UpdateOn  time.Time `json:"updatedOn" db:"updated_on"`
}
