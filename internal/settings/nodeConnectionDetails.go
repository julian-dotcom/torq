package settings

import (
	"fmt"
	"mime/multipart"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
)

type NodeConnectionDetails struct {
	NodeId            int                                     `json:"nodeId" form:"nodeId" db:"node_id"`
	Name              string                                  `json:"name" form:"name" db:"name"`
	Implementation    core.Implementation                     `json:"implementation" form:"implementation" db:"implementation"`
	GRPCAddress       *string                                 `json:"grpcAddress" form:"grpcAddress" db:"grpc_address"`
	TLSFileName       *string                                 `json:"tlsFileName" db:"tls_file_name"`
	TLSDataBytes      []byte                                  `db:"tls_data"`
	TLSFile           *multipart.FileHeader                   `form:"tlsFile"`
	MacaroonFileName  *string                                 `json:"macaroonFileName" db:"macaroon_file_name"`
	MacaroonDataBytes []byte                                  `db:"macaroon_data"`
	MacaroonFile      *multipart.FileHeader                   `form:"macaroonFile"`
	Status            core.Status                             `json:"status" db:"status_id"`
	PingSystem        core.PingSystem                         `json:"pingSystem" db:"ping_system"`
	CustomSettings    core.NodeConnectionDetailCustomSettings `json:"customSettings" db:"custom_settings"`
	CreateOn          time.Time                               `json:"createdOn" db:"created_on"`
	UpdatedOn         *time.Time                              `json:"updatedOn"  db:"updated_on"`
}

func GetNodeIdByGRPC(db *sqlx.DB, grpcAddress string) (int, error) {
	allNodeConnectionDetails, err := GetAllNodeConnectionDetails(db, true)
	if err != nil {
		return 0, errors.Wrap(err, "Getting local nodes from db")
	}
	for _, nodeConnectionDetailsData := range allNodeConnectionDetails {
		if nodeConnectionDetailsData.GRPCAddress != nil &&
			*nodeConnectionDetailsData.GRPCAddress == grpcAddress {
			return nodeConnectionDetailsData.NodeId, nil
		}
	}
	return 0, nil
}

func AddNodeToDB(db *sqlx.DB, implementation core.Implementation,
	grpcAddress string, tlsDataBytes []byte, macaroonDataBytes []byte) (NodeConnectionDetails, error) {
	publicKey, chain, network, err := getInformationFromLndNode(grpcAddress, tlsDataBytes, macaroonDataBytes)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Getting public key from node")
	}
	newNodeFromConfig := nodes.Node{
		PublicKey: publicKey,
		Chain:     chain,
		Network:   network,
	}
	nodeId, err := nodes.AddNodeWhenNew(db, newNodeFromConfig, nil)
	nodeId, err := AddNodeWhenNew(db, publicKey, chain, network)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Getting node from db")
	}
	existingNodeConnectionDetails, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Getting all existing node connection details from db")
	}

	if existingNodeConnectionDetails.NodeId == nodeId {
		existingNodeConnectionDetails.GRPCAddress = &grpcAddress
		existingNodeConnectionDetails.TLSDataBytes = tlsDataBytes
		existingNodeConnectionDetails.MacaroonDataBytes = macaroonDataBytes
		ncd, err := SetNodeConnectionDetails(db, existingNodeConnectionDetails)
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Updating node connection details in the database")
		}
		return ncd, nil
	}

	nodeConnectionDetailsData := NodeConnectionDetails{
		NodeId:            nodeId,
		Name:              fmt.Sprintf("Node_%v", nodeId),
		Implementation:    implementation,
		GRPCAddress:       &grpcAddress,
		Status:            core.Active,
		TLSDataBytes:      tlsDataBytes,
		MacaroonDataBytes: macaroonDataBytes,
		CreateOn:          time.Now().UTC(),
	}
	ncd, err := addNodeConnectionDetails(db, nodeConnectionDetailsData)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Inserting node connection details in the database")
	}
	cache.SetTorqNode(nodeId, nodeConnectionDetailsData.Name, nodeConnectionDetailsData.Status, publicKey, chain, network)
	return ncd, nil
}
