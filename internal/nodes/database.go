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

func GetPeerNodes(db *sqlx.DB, network commons.Network) ([]Node, error) {
	var nodes []Node
	err := db.Select(&nodes, `
	SELECT node_id, public_key, chain, network, created_on, node_connection_details_node_id, connection_status_id, host
	FROM node
	WHERE network = $1 and node_connection_details_node_id IS NOT NULL;`, network)
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
		err := db.QueryRowx(`INSERT INTO node (public_key, chain, network, created_on, node_connection_details_node_id, connection_status_id, host, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING node_id;`,
			node.PublicKey, node.Chain, node.Network, node.CreatedOn, node.NodeConnectionDetailsNodeId, node.ConnectionStatusId, node.Host, time.Now().UTC()).Scan(&node.NodeId)
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

func updateNodeHost(db *sqlx.DB, node Node) error {
	_, err := db.Exec(`UPDATE node SET host = $1, updated_on = $2  WHERE node_id = $3;`, node.Host, time.Now().UTC(), node.NodeId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}
	return nil
}

func updateNodeConnectionStatus(db *sqlx.DB, node Node) error {
	_, err := db.Exec(`UPDATE node SET connection_status_id = $1, updated_on = $2  WHERE node_id = $3;`, node.ConnectionStatusId, time.Now().UTC(), node.NodeId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}

	_, err = db.Exec(`INSERT INTO node_connection_history (node_id, connection_status, created_on) VALUES ($1, $2, $3);`, node.NodeId, node.ConnectionStatusId, time.Now().UTC())
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}

	return nil
}
