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

func GetConnectionDetails(db *sqlx.DB) (connectionDetails, error) {
	localNodeDetails, err := getLocalNodeConnectionDetails(db)
	if err != nil {
		return connectionDetails{}, err
	}
	if (localNodeDetails.GRPCAddress == nil) || (localNodeDetails.TLSDataBytes == nil) || (localNodeDetails.
		MacaroonDataBytes == nil) {
		return connectionDetails{}, errors.New("Missing node details")
	}
	return connectionDetails{
		GRPCAddress:       *localNodeDetails.GRPCAddress,
		TLSFileBytes:      localNodeDetails.TLSDataBytes,
		MacaroonFileBytes: localNodeDetails.MacaroonDataBytes}, nil
}
