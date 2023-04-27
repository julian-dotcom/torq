package peers

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/tags"
)

type PeerNode struct {
	NodeId               int                         `json:"nodeId" db:"node_id"`
	Alias                *string                     `json:"peerAlias" db:"alias"`
	PublicKey            string                      `json:"pubKey" db:"public_key"`
	TorqNodeId           *int                        `json:"torqNodeId" db:"torq_node_id"`
	TorqNodeAlias        *string                     `json:"torqNodeAlias" db:"torq_node_alias"`
	Setting              *core.NodeConnectionSetting `json:"setting" db:"setting"`
	ConnectionStatus     *ConnectionStatus           `json:"connectionStatus" db:"connection_status"`
	Address              *string                     `json:"address" db:"address"`
	SecondsConnected     int                         `json:"secondsConnected" db:"seconds_connected"`
	SecondsDisconnected  int                         `json:"secondsDisconnected" db:"seconds_disconnected"`
	DateLastDisconnected *time.Time                  `json:"dateLastDisconnected" db:"date_last_disconnected"`
	DateLastConnected    *time.Time                  `json:"dateLastConnected" db:"date_last_connected"`
	Tags                 []tags.Tag                  `json:"tags"`
}

func (p PeerNode) MarshalJSON() ([]byte, error) {
	type Alias PeerNode // create an alias to avoid infinite recursion
	statusStr := ""
	if p.ConnectionStatus != nil {
		statusStr = p.ConnectionStatus.String()
	}
	settingStr := ""
	if p.Setting != nil {
		settingStr = p.Setting.String()
	}
	jsonBytes, err := json.Marshal(&struct {
		*Alias
		ConnectionStatus string `json:"connectionStatus"`
		Setting          string `json:"setting"`
	}{
		Alias:            (*Alias)(&p),
		ConnectionStatus: statusStr,
		Setting:          settingStr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Marshalling PeerNode to JSON.")
	}
	return jsonBytes, nil
}

func GetPeerNodes(db *sqlx.DB, network core.Network) ([]PeerNode, error) {
	nodeIds := cache.GetChannelPeerNodeIds(core.Bitcoin, network)
	connectedPeerNodeIds := cache.GetConnectedPeerNodeIds(core.Bitcoin, network)
	for _, connectedPeerNodeId := range connectedPeerNodeIds {
		if !slices.Contains(nodeIds, connectedPeerNodeId) {
			nodeIds = append(nodeIds, connectedPeerNodeId)
		}
	}

	var nodes []PeerNode
	err := db.Select(&nodes, `
	SELECT
		n.node_id,
		ne.alias,
		nch.torq_node_id,
		netorq.alias AS torq_node_alias,
		n.public_key,
		nch.connection_status,
		nch.setting,
		last_disconnection.disconnected_since AS date_last_disconnected,
		last_connection.connected_since AS date_last_connected,
		CASE
		    WHEN connection_status = $2 THEN 0
		    WHEN last_disconnection.disconnected_since IS NULL THEN 0
		    ELSE  EXTRACT(EPOCH FROM (now() - last_disconnection.disconnected_since))::int
		END seconds_disconnected,
		CASE
		    WHEN connection_status = $1 THEN 0
		    WHEN last_connection.connected_since IS NULL THEN 0
		    ELSE  EXTRACT(EPOCH FROM (now() - last_connection.connected_since))::int
		END seconds_connected
	FROM Node n
	LEFT JOIN (
		SELECT LAST(node_id, created_on) as node_id,
		       LAST(torq_node_id, created_on) as torq_node_id,
		       LAST(connection_status, created_on) as connection_status,
		       LAST(setting, created_on) as setting
		FROM node_connection_history
		GROUP BY node_id
	) nch on nch.node_id = n.node_id
	LEFT JOIN (
        SELECT node_id, max(created_on) as disconnected_since
        FROM (
            SELECT node_id,
                    LAST(connection_status, created_on) AS connection_status,
                    LAG(connection_status) OVER (PARTITION BY node_id ORDER BY created_on ) AS previous_connection_status,
                    created_on
            FROM node_connection_history
            GROUP BY node_id, created_on, connection_status
            ) nch_with_lag_status
        WHERE nch_with_lag_status.connection_status = $1
          AND (nch_with_lag_status.previous_connection_status = $2 OR nch_with_lag_status.previous_connection_status IS NULL)
        GROUP BY node_id
    ) last_disconnection on last_disconnection.node_id = n.node_id
	LEFT JOIN (
        SELECT node_id, max(created_on) as connected_since
        FROM (
            SELECT node_id,
                    LAST(connection_status, created_on) AS connection_status,
                    created_on
            FROM node_connection_history
            GROUP BY node_id, created_on, connection_status
            ) nch_with_lag_status
        WHERE nch_with_lag_status.connection_status = $2
        GROUP BY node_id
    ) last_connection on last_connection.node_id = n.node_id
	LEFT JOIN (
		SELECT LAST(event_node_id, timestamp) as node_id,
		       LAST(alias, timestamp) as alias
		FROM node_event
		GROUP BY event_node_id
	) ne ON ne.node_id = n.node_id
	LEFT JOIN (
		SELECT LAST(event_node_id, timestamp) as node_id,
		       LAST(alias, timestamp) as alias
		FROM node_event
		GROUP BY event_node_id
	) netorq ON netorq.node_id = nch.torq_node_id
	JOIN node_connection_details as ncd ON ncd.node_id = nch.torq_node_id
	WHERE nch.torq_node_id IS NOT NULL
		AND ncd.status_id NOT IN ($3, $4)
		AND (n.node_id = ANY($5) OR (nch.setting IS NOT NULL AND n.network = $6));`,
		Disconnected, Connected, core.Deleted, core.Archived, pq.Array(nodeIds), network)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []PeerNode{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodes, nil
}
