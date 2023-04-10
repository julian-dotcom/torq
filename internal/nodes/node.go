package nodes

import (
	"time"

	"github.com/lncapital/torq/internal/core"
)

type Node struct {
	NodeId    int          `json:"nodeId" db:"node_id"`
	PublicKey string       `json:"publicKey" db:"public_key"`
	Chain     core.Chain   `json:"chain" db:"chain"`
	Network   core.Network `json:"network" db:"network"`
	CreatedOn time.Time    `json:"createdOn" db:"created_on"`
	// Will never be updated so no UpdatedOn...
}

type NodeSummary struct {
	NodeId    int          `json:"nodeId" db:"node_id"`
	PublicKey string       `json:"publicKey" db:"public_key"`
	Chain     core.Chain   `json:"chain" db:"chain"`
	Network   core.Network `json:"network" db:"network"`
	CreatedOn time.Time    `json:"createdOn" db:"created_on"`
	Status    core.Status  `json:"status" db:"status_id"`
	Name      string       `json:"name" db:"name"`
	// Will never be updated so no UpdatedOn...
}

type NodeConnectionHistory struct {
	NodeId           int                        `json:"nodeId" db:"node_id"`
	TorqNodeId       int                        `json:"torqNodeId" db:"torq_node_id"`
	ConnectionStatus core.NodeConnectionStatus  `json:"connectionStatus" db:"connection_status"`
	Address          string                     `json:"address" db:"address"`
	Setting          core.NodeConnectionSetting `json:"setting" db:"setting"`
	CreatedOn        time.Time                  `json:"createdOn" db:"created_on"`
}

type NodeConnectionHistoryWithDetail struct {
	NodeId           int                        `json:"nodeId" db:"node_id"`
	TorqNodeId       int                        `json:"torqNodeId" db:"torq_node_id"`
	ConnectionStatus core.NodeConnectionStatus  `json:"connectionStatus" db:"connection_status"`
	Address          string                     `json:"address" db:"address"`
	Setting          core.NodeConnectionSetting `json:"setting" db:"setting"`
	PubKey           string                     `json:"public_key" db:"public_key"`
}
