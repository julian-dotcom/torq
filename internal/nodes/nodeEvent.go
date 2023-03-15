package nodes

import (
	"time"
)

type NodeEvent struct {
	EventNodeId   int       `json:"eventNodeId" db:"event_node_id"`
	NodeId        int       `json:"nodeId" db:"node_id"`
	EventTime     time.Time `json:"eventTime" db:"timestamp"`
	PublicKey     string    `json:"publicKey" db:"pub_key"`
	Alias         string    `json:"alias" db:"alias"`
	Color         string    `json:"color" db:"color"`
	NodeAddresses string    `json:"nodeAddresses" db:"node_addresses"`
	Features      string    `json:"features" db:"features"`
	// Will never be updated so no UpdatedOn...
}
