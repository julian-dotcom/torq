package nodes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
)

func GetNodeByPublicKey(db *sqlx.DB, publicKey string) (Node, error) {
	var n Node
	err := db.Get(&n, `SELECT * FROM node WHERE public_key=$1;`, publicKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Node{}, nil
		}
		return Node{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return n, nil
}

func getAllNodeInformationByNetwork(db *sqlx.DB, network core.Network) ([]NodeInformation, error) {
	nds, err := getNodesByNetwork(db, false, network)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []NodeInformation{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	var ps []NodeInformation
	for _, node := range nds {
		nodeEvent, err := getLatestNodeEvent(db, node.NodeId)
		if err != nil {
			return nil, err
		}
		ni := NodeInformation{
			NodeId:    node.NodeId,
			PublicKey: node.PublicKey,
			Status:    node.Status,
			TorqAlias: node.Name,
			Alias:     nodeEvent.Alias,
			Color:     nodeEvent.Color,
		}

		var addresses []NodeAddress
		err = json.Unmarshal(nodeEvent.NodeAddresses, &addresses)
		if err != nil {
			return nil, errors.Wrap(err, "Unmarshalling json to project on to nodeAddress struct")
		}

		ni.Addresses = &addresses

		ps = append(ps, ni)
	}
	return ps, nil
}

func getNodesByNetwork(db *sqlx.DB, includeDeleted bool, network core.Network) ([]NodeSummary, error) {
	var nds []NodeSummary

	query := `SELECT n.node_id, n.public_key, n.chain, n.network, n.created_on, ncd.name, ncd.status_id FROM node n JOIN node_connection_details ncd on ncd.node_id = n.node_id where n.network = $1;`

	if !includeDeleted {
		query = `SELECT n.node_id, n.public_key, n.chain, n.network, n.created_on, ncd.name, ncd.status_id FROM node n JOIN node_connection_details ncd on ncd.node_id = n.node_id where ncd.status_id != 3 AND n.network = $1;`
	}

	err := db.Select(&nds, query, network)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []NodeSummary{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nds, nil
}

func getLatestNodeEvent(db *sqlx.DB, nodeId int) (NodeEvent, error) {
	var nodeEvent NodeEvent
	err := db.Get(&nodeEvent, `
		SELECT *
		FROM node_event
		WHERE event_node_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NodeEvent{}, nil
		}
		return NodeEvent{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodeEvent, nil
}

func AddNodeWhenNew(db *sqlx.DB, node Node, peerConnectionHistory *NodeConnectionHistory) (int, error) {
	nodeId := cache.GetNodeIdByPublicKey(node.PublicKey, node.Chain, node.Network)
	if nodeId == 0 {
		node.CreatedOn = time.Now().UTC()
		err := db.QueryRowx(`INSERT INTO node (public_key, chain, network, created_on)
			VALUES ($1, $2, $3, $4) RETURNING node_id;`,
			node.PublicKey, node.Chain, node.Network, node.CreatedOn).Scan(&node.NodeId)
		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				if err.Code == "23505" {
					storedNode, err := GetNodeByPublicKey(db, node.PublicKey)
					return storedNode.NodeId, err
				}
			}
			return 0, errors.Wrap(err, database.SqlExecutionError)
		}

		if peerConnectionHistory != nil {
			peerConnectionHistory.NodeId = node.NodeId
			err = addNodeConnectionHistory(db, peerConnectionHistory)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to store Node Connection History for nodeId: %v", nodeId)
			}
		}

		return node.NodeId, nil
	}
	return nodeId, nil
}

func removeNode(db *sqlx.DB, nodeId int) (int64, error) {
	referencingCorridors, err := corridors.GetCorridorsReferencingNode(db, nodeId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if len(referencingCorridors) > 0 {
		return 0, errors.New(fmt.Sprintf("Could not remove node since it's in use. %v", referencingCorridors))
	}
	res, err := db.Exec(`DELETE FROM node WHERE node_id = $1;`, nodeId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func addNodeConnectionHistory(db *sqlx.DB, nch *NodeConnectionHistory) error {
	if nch == nil {
		return nil
	}
	_, err := db.Exec(
		`INSERT INTO node_connection_history (
                                     node_id,
                                     torq_node_id,
                                     connection_status,
                                     address,
                                     setting,
                                     created_on) VALUES ($1, $2, $3, $4, $5, $6);`,
		nch.NodeId,
		nch.TorqNodeId,
		nch.ConnectionStatus,
		nch.Address,
		nch.Setting,
		time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}
