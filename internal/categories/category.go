package categories

import (
	"time"
)

type Category struct {
	CategoryId int       `json:"categoryId" db:"category_id"`
	Name       string    `json:"name" db:"name"`
	Style      string    `json:"style" db:"style"`
	CreatedOn  time.Time `json:"createdOn" db:"created_on"`
	UpdateOn   time.Time `json:"updatedOn" db:"updated_on"`
}
