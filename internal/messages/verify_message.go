package messages

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

func verifyMessage(db *sqlx.DB, req VerifyMessageRequest) (r VerifyMessageResponse, err error) {
	if req.NodeId == 0 {
		return VerifyMessageResponse{}, errors.Newf("Node Id missing")
	}

	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return VerifyMessageResponse{}, errors.Wrap(err, "Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return VerifyMessageResponse{}, errors.Wrap(err, "Connecting to LND")
	}
	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)

	verifyMsgReq := lnrpc.VerifyMessageRequest{
		Msg:       []byte(req.Message),
		Signature: req.Signature,
	}

	ctx := context.Background()

	verifyMsgResp, err := client.VerifyMessage(ctx, &verifyMsgReq)
	if err != nil {
		return VerifyMessageResponse{}, errors.Wrap(err, "Verifying message")
	}

	r.Valid = verifyMsgResp.GetValid()
	r.PubKey = verifyMsgResp.GetPubkey()

	return r, nil
}
