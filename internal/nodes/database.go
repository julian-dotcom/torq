package nodes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/commons"
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

func GetNodeById(db *sqlx.DB, nodeId int) (Node, error) {
	var n Node
	err := db.Get(&n, `SELECT * FROM node WHERE node_id=$1;`, nodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Node{}, nil
		}
		return Node{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return n, nil
}

func getAllNodeInformationByNetwork(db *sqlx.DB, network commons.Network) ([]NodeInformation, error) {
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

func getNodesByNetwork(db *sqlx.DB, includeDeleted bool, network commons.Network) ([]NodeSummary, error) {
	var nds []NodeSummary

	query := `SELECT n.*, ncd.name, ncd.status_id FROM node n JOIN node_connection_details ncd on ncd.node_id = n.node_id where n.network = $1;`

	if !includeDeleted {
		query = `SELECT n.*, ncd.name, ncd.status_id FROM node n JOIN node_connection_details ncd on ncd.node_id = n.node_id where ncd.status_id != 3 AND n.network = $1;`
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

func GetPeerNodes(db *sqlx.DB, network commons.Network) ([]Node, error) {
	var nodes []Node
	err := db.Select(&nodes, `
	SELECT n.node_id, public_key, chain, network, n.created_on
	FROM node n
	LEFT JOIN node_connection_details ncd ON n.node_id = ncd.node_id
	WHERE n.network = $1 and ncd.node_id IS NULL;`, network)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Node{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return nodes, nil
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

func AddNodeWhenNew(db *sqlx.DB, node Node) (int, error) {
	nodeId := commons.GetNodeIdByPublicKey(node.PublicKey, node.Chain, node.Network)
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
