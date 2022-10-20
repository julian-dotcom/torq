package settings

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

type ConnectionDetails struct {
	LocalNodeId       int
	Name              string
	GRPCAddress       string
	TLSFileBytes      []byte
	MacaroonFileBytes []byte
	Disabled          bool
	Deleted           bool
}

func GetActiveNodesConnectionDetails(db *sqlx.DB) (activeNodes []ConnectionDetails, err error) {
	//Get all nodes not disabled and not deleted
	localNodes, err := getLocalNodeConnectionDetails(db)
	if err != nil {
		return []ConnectionDetails{}, errors.Wrap(err, "Getting local nodes from db")
	}

	for _, localNodeDetails := range localNodes {
		if (localNodeDetails.GRPCAddress == nil) || (localNodeDetails.TLSDataBytes == nil) || (localNodeDetails.
			MacaroonDataBytes == nil) {
			continue
		}
		activeNodes = append(activeNodes, ConnectionDetails{
			LocalNodeId:       localNodeDetails.LocalNodeId,
			GRPCAddress:       *localNodeDetails.GRPCAddress,
			TLSFileBytes:      localNodeDetails.TLSDataBytes,
			MacaroonFileBytes: localNodeDetails.MacaroonDataBytes,
			Name:              localNodeDetails.Name})
	}

	return activeNodes, nil
}

func GetNodeConnectionDetailsById(db *sqlx.DB, nodeId int) (connectionDetails ConnectionDetails, err error) {
	// will still fetch details even if node is disabled or deleted
	node, err := getLocalNodeConnectionDetailsById(db, nodeId)
	if err != nil {
		return ConnectionDetails{}, err
	}
	cd := ConnectionDetails{
		LocalNodeId:       node.LocalNodeId,
		Name:              node.Name,
		TLSFileBytes:      node.TLSDataBytes,
		MacaroonFileBytes: node.MacaroonDataBytes,
		Disabled:          node.Disabled,
		Deleted:           node.Deleted,
	}
	if node.GRPCAddress != nil {
		cd.GRPCAddress = *node.GRPCAddress
	}
	return cd, nil
}
