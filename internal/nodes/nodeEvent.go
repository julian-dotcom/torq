package nodes

import (
	"encoding/json"
	"time"
)

type NodeAddress struct {
	Addr    string `json:"addr"`
	Network string `json:"network"`
}
type NodeEvent struct {
	EventNodeId   int             `json:"eventNodeId" db:"event_node_id"`
	NodeId        int             `json:"nodeId" db:"node_id"`
	EventTime     time.Time       `json:"eventTime" db:"timestamp"`
	PublicKey     string          `json:"publicKey" db:"pub_key"`
	Alias         string          `json:"alias" db:"alias"`
	Color         string          `json:"color" db:"color"`
	NodeAddresses json.RawMessage `json:"nodeAddresses" db:"node_addresses"`
	Features      string          `json:"features" db:"features"`
	// Will never be updated so no UpdatedOn...
}
