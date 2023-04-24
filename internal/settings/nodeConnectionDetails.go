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
	NodeId                 int                                     `json:"nodeId" form:"nodeId" db:"node_id"`
	Name                   string                                  `json:"name" form:"name" db:"name"`
	Implementation         core.Implementation                     `json:"implementation" form:"implementation" db:"implementation"`
	GRPCAddress            *string                                 `json:"grpcAddress" form:"grpcAddress" db:"grpc_address"`
	TLSFileName            *string                                 `json:"tlsFileName" db:"tls_file_name"`
	TLSDataBytes           []byte                                  `db:"tls_data"`
	TLSFile                *multipart.FileHeader                   `form:"tlsFile"`
	MacaroonFileName       *string                                 `json:"macaroonFileName" db:"macaroon_file_name"`
	MacaroonDataBytes      []byte                                  `db:"macaroon_data"`
	MacaroonFile           *multipart.FileHeader                   `form:"macaroonFile"`
	CertificateFileName    *string                                 `json:"certificateFileName" db:"certificate_file_name"`
	CertificateDataBytes   []byte                                  `db:"certificate_data"`
	CertificateFile        *multipart.FileHeader                   `form:"certificateFile"`
	KeyFileName            *string                                 `json:"KeyFileName" db:"key_file_name"`
	KeyDataBytes           []byte                                  `db:"key_data"`
	KeyFile                *multipart.FileHeader                   `form:"keyFile"`
	CaCertificateFileName  *string                                 `json:"caCertificateFileName" db:"ca_certificate_file_name"`
	CaCertificateDataBytes []byte                                  `db:"ca_certificate_data"`
	CaCertificateFile      *multipart.FileHeader                   `form:"caCertificateFile"`
	Status                 core.Status                             `json:"status" db:"status_id"`
	PingSystem             core.PingSystem                         `json:"pingSystem" db:"ping_system"`
	CustomSettings         core.NodeConnectionDetailCustomSettings `json:"customSettings" db:"custom_settings"`
	NodeStartDate          *time.Time                              `json:"nodeStartDate"  db:"node_start_date"`
	CreateOn               time.Time                               `json:"createdOn" db:"created_on"`
	UpdatedOn              *time.Time                              `json:"updatedOn"  db:"updated_on"`
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
	grpcAddress string, certificate []byte, authentication []byte, caCertificate []byte) (NodeConnectionDetails, error) {
	var publicKey string
	var chain core.Chain
	var network core.Network
	var err error
	switch implementation {
	case core.LND:
		publicKey, chain, network, err = getInformationFromLndNode(grpcAddress, certificate, authentication)
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Getting public key from LND node")
		}
	case core.CLN:
		publicKey, chain, network, err = getInformationFromClnNode(grpcAddress, certificate, authentication, caCertificate)
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Getting public key from CLN node")
		}
	}
	nodeId, err := AddNodeWhenNew(db, publicKey, chain, network)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Getting node from db")
	}
	existingNodeConnectionDetails, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Getting all existing node connection details from db")
	}

	if existingNodeConnectionDetails.NodeId == nodeId {
		existingNodeConnectionDetails.Implementation = implementation
		existingNodeConnectionDetails.GRPCAddress = &grpcAddress
		switch implementation {
		case core.LND:
			existingNodeConnectionDetails.TLSDataBytes = certificate
			existingNodeConnectionDetails.MacaroonDataBytes = authentication
		case core.CLN:
			existingNodeConnectionDetails.CertificateDataBytes = certificate
			existingNodeConnectionDetails.KeyDataBytes = authentication
		}
		ncd, err := SetNodeConnectionDetails(db, existingNodeConnectionDetails)
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Updating node connection details in the database")
		}
		return ncd, nil
	}

	nodeStartDate, err := getNodeStartDateFromLndNode(grpcAddress, certificate, authentication)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Getting node start date from lnd node")
	}

	nodeConnectionDetailsData := NodeConnectionDetails{
		NodeId:         nodeId,
		Name:           fmt.Sprintf("Node_%v", nodeId),
		Implementation: implementation,
		GRPCAddress:    &grpcAddress,
		Status:         core.Active,
		NodeStartDate:  nodeStartDate,
		CreateOn:       time.Now().UTC(),
	}
	switch implementation {
	case core.LND:
		existingNodeConnectionDetails.TLSDataBytes = certificate
		existingNodeConnectionDetails.MacaroonDataBytes = authentication
	case core.CLN:
		existingNodeConnectionDetails.CertificateDataBytes = certificate
		existingNodeConnectionDetails.KeyDataBytes = authentication
	}
	ncd, err := addNodeConnectionDetails(db, nodeConnectionDetailsData)
	if err != nil {
		return NodeConnectionDetails{}, errors.Wrap(err, "Inserting node connection details in the database")
	}
	cache.SetTorqNode(nodeId, nodeConnectionDetailsData.Name, nodeConnectionDetailsData.Status, publicKey, chain, network)
	return ncd, nil
}
