package settings

import (
	"github.com/jmoiron/sqlx"
)

type connectionDetails struct {
	GRPCAddress       string
	TLSFileBytes      []byte
	MacaroonFileBytes []byte
}

func GetConnectionDetails(db *sqlx.DB) (connectionDetails, error) {
	localNodeDetails, err := getLocalNodeConnectionDetails(db)
	if err != nil {
		return connectionDetails{}, err
	}
	return connectionDetails{
		GRPCAddress:       *localNodeDetails.GRPCAddress,
		TLSFileBytes:      localNodeDetails.TLSDataBytes,
		MacaroonFileBytes: localNodeDetails.MacaroonDataBytes}, nil
}
