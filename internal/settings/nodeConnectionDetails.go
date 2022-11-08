package settings

import (
	"fmt"
	"mime/multipart"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"
)

type nodeConnectionDetails struct {
	NodeId            int                    `json:"nodeId" form:"nodeId" db:"node_id"`
	Name              string                 `json:"name" form:"name" db:"name"`
	Implementation    commons.Implementation `json:"implementation" form:"implementation" db:"implementation"`
	GRPCAddress       *string                `json:"grpcAddress" form:"grpcAddress" db:"grpc_address"`
	TLSFileName       *string                `json:"tlsFileName" db:"tls_file_name"`
	TLSDataBytes      []byte                 `db:"tls_data"`
	TLSFile           *multipart.FileHeader  `form:"tlsFile"`
	MacaroonFileName  *string                `json:"macaroonFileName" db:"macaroon_file_name"`
	MacaroonDataBytes []byte                 `db:"macaroon_data"`
	MacaroonFile      *multipart.FileHeader  `form:"macaroonFile"`
	Status            commons.Status         `json:"status" db:"status_id"`
	CreateOn          time.Time              `json:"createdOn" db:"created_on"`
	UpdatedOn         *time.Time             `json:"updatedOn"  db:"updated_on"`
}

func GetNodeIdByGRPC(db *sqlx.DB, grpcAddress string) (int, error) {
	allNodeConnectionDetails, err := getAllNodeConnectionDetails(db, true)
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

func AddNodeToDB(db *sqlx.DB, implementation commons.Implementation,
	grpcAddress string, tlsDataBytes []byte, macaroonDataBytes []byte) (nodeConnectionDetails, error) {
	publicKey, chain, network, err := getInformationFromLndNode(grpcAddress, tlsDataBytes, macaroonDataBytes)
	if err != nil {
		return nodeConnectionDetails{}, errors.Wrap(err, "Getting public key from node")
	}
	newNodeFromConfig := nodes.Node{
		PublicKey: publicKey,
		Chain:     chain,
		Network:   network,
	}
	nodeId, err := nodes.AddNodeWhenNew(db, newNodeFromConfig)
	if err != nil {
		return nodeConnectionDetails{}, errors.Wrap(err, "Getting node from db")
	}
	existingNodeConnectionDetails, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		return nodeConnectionDetails{}, errors.Wrap(err, "Getting all existing node connection details from db")
	}

	if existingNodeConnectionDetails.NodeId == nodeId {
		existingNodeConnectionDetails.GRPCAddress = &grpcAddress
		existingNodeConnectionDetails.TLSDataBytes = tlsDataBytes
		existingNodeConnectionDetails.MacaroonDataBytes = macaroonDataBytes
		ncd, err := SetNodeConnectionDetails(db, existingNodeConnectionDetails)
		if err != nil {
			return nodeConnectionDetails{}, errors.Wrap(err, "Updating node connection details in the database")
		}
		return ncd, nil
	}

	nodeConnectionDetailsData := nodeConnectionDetails{
		NodeId:            nodeId,
		Name:              fmt.Sprintf("Node_%v", nodeId),
		Implementation:    implementation,
		GRPCAddress:       &grpcAddress,
		Status:            commons.Active,
		TLSDataBytes:      tlsDataBytes,
		MacaroonDataBytes: macaroonDataBytes,
		CreateOn:          time.Now().UTC(),
	}
	ncd, err := addNodeConnectionDetails(db, nodeConnectionDetailsData)
	if err != nil {
		return nodeConnectionDetails{}, errors.Wrap(err, "Inserting node connection details in the database")
	}
	commons.SetTorqNode(nodeId, nodeConnectionDetailsData.Status, publicKey, chain, network)
	return ncd, nil
}
