package nodes

import (
	"time"
)

type Node struct {
	NodeId    int       `json:"nodeId" db:"node_id"`
	PublicKey string    `json:"publicKey" db:"public_key"`
	CreatedOn time.Time `json:"createdOn" db:"created_on"`
	// Will never be updated so no UpdatedOn...
}
