package on_chain_tx

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

func newAddress(db *sqlx.DB, req newAddressRequest) (r string, err error) {
	if req.NodeId == 0 {
		return "", errors.New("Node id is missing")
	}

	addressType := req.Type
	account := req.Account

	connectionDetails, err := settings.GetNodeConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return "", errors.Wrap(err, "Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return "", errors.Wrap(err, "Connecting to LND")
	}

	defer conn.Close()

	lnAddressType := lnrpc.AddressType(addressType)
	newAddressReq := lnrpc.NewAddressRequest{Type: lnAddressType, Account: account}

	client := lnrpc.NewLightningClient(conn)
	ctx := context.Background()

	resp, err := client.NewAddress(ctx, &newAddressReq)
	if err != nil {
		return "", errors.Wrap(err, "Creating new address")
	}

	//log.Debug().Msgf("New address : %v", resp.Address)
	return resp.Address, nil
}
