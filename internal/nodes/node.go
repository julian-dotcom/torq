package nodes

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type Node struct {
	NodeId                      int             `json:"nodeId" db:"node_id"`
	PublicKey                   string          `json:"pubKey" db:"public_key"`
	Chain                       commons.Chain   `json:"chain" db:"chain"`
	Network                     commons.Network `json:"network" db:"network"`
	CreatedOn                   time.Time       `json:"createdOn" db:"created_on"`
	NodeConnectionDetailsNodeId *int            `json:"nodeConnectionDetailsNodeId" db:"node_connection_details_node_id" description:"The node that established a connection to this node"`
	ConnectionStatusId          commons.Status  `json:"connectionStatus" db:"connection_status_id"`
	Host                        string          `json:"host" db:"host"`
	UpdatedOn                   time.Time       `json:"updatedOn" db:"updated_on"`
}

type NodeSummary struct {
	NodeId    int             `json:"nodeId" db:"node_id"`
	PublicKey string          `json:"publicKey" db:"public_key"`
	Chain     commons.Chain   `json:"chain" db:"chain"`
	Network   commons.Network `json:"network" db:"network"`
	CreatedOn time.Time       `json:"createdOn" db:"created_on"`
	Status    commons.Status  `json:"status" db:"status_id"`
	Name      string          `json:"name" db:"name"`
	// Will never be updated so no UpdatedOn...
}
