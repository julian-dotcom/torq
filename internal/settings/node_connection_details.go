package settings

import (
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

type connectionDetails struct {
	GRPCAddress       string
	TLSFileBytes      []byte
	MacaroonFileBytes []byte
}

func GetConnectionDetails(db *sqlx.DB) ([]connectionDetails, error) {
	localNodes, err := getLocalNodeConnectionDetails(db)
	if err != nil {
		return []connectionDetails{}, err
	}
	connectionDetailsList := []connectionDetails{}

	for _, localNodeDetails := range localNodes {
		if (localNodeDetails.GRPCAddress == nil) || (localNodeDetails.TLSDataBytes == nil) || (localNodeDetails.
			MacaroonDataBytes == nil) {
			continue
		}
		connectionDetailsList = append(connectionDetailsList, connectionDetails{
			GRPCAddress:       *localNodeDetails.GRPCAddress,
			TLSFileBytes:      localNodeDetails.TLSDataBytes,
			MacaroonFileBytes: localNodeDetails.MacaroonDataBytes})
	}
	if len(connectionDetailsList) == 0 {
		return []connectionDetails{}, errors.New("Missing node details")
	}
	return connectionDetailsList, nil
}
