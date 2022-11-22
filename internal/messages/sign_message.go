package messages

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

func signMessage(db *sqlx.DB, req SignMessageRequest) (r SignMessageResponse, err error) {
	if req.NodeId == 0 {
		return SignMessageResponse{}, errors.New("Node Id missing")
	}
	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return SignMessageResponse{}, errors.Wrap(err, "Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return SignMessageResponse{}, errors.Wrap(err, "Connecting to LND")
	}
	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)

	ctx := context.Background()

	signMsgReq := lnrpc.SignMessageRequest{
		Msg: []byte(req.Message),
	}

	if req.SingleHash != nil {
		signMsgReq.SingleHash = *req.SingleHash
	}

	signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
	if err != nil {
		return SignMessageResponse{}, errors.Wrap(err, "Signing message")
	}

	r.Signature = signMsgResp.Signature

	return r, nil
}
