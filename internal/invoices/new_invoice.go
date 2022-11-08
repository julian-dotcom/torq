package invoices

import (
	"context"
	"encoding/hex"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

type newInvoiceRequest struct {
	NodeId          int     `json:"nodeId"`
	Memo            *string `json:"memo"`
	RPreImage       *string `json:"rPreImage"`
	ValueMsat       *int64  `json:"valueMsat"`
	Expiry          *int64  `json:"expiry"`
	FallBackAddress *string `json:"fallBackAddress"`
	Private         *bool   `json:"private"`
	IsAmp           *bool   `json:"isAmp"`
}

type newInvoiceResponse struct {
	NodeId         int    `json:"nodeId"`
	PaymentRequest string `json:"paymentRequest"`
	AddIndex       uint64 `json:"addIndex"`
	PaymentAddress string `json:"paymentAddress"`
}

func newInvoice(db *sqlx.DB, req newInvoiceRequest) (r newInvoiceResponse, err error) {
	newInvoiceReq, err := processInvoiceReq(req)
	if err != nil {
		return r, err
	}

	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return r, errors.Wrap(err, "Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return r, errors.Wrap(err, "Connecting to LND")
	}

	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)

	ctx := context.Background()

	resp, err := client.AddInvoice(ctx, newInvoiceReq)
	if err != nil {
		return newInvoiceResponse{}, errors.Wrap(err, "Creating invoice on node")
	}

	r.NodeId = req.NodeId
	r.PaymentRequest = resp.GetPaymentRequest()
	r.AddIndex = resp.GetAddIndex()
	r.PaymentAddress = hex.EncodeToString(resp.GetPaymentAddr())

	return r, nil
}

func processInvoiceReq(req newInvoiceRequest) (inv *lnrpc.Invoice, err error) {
	inv = &lnrpc.Invoice{}

	if req.NodeId == 0 {
		return &lnrpc.Invoice{}, errors.New("Node id is missing")
	}

	if req.Memo != nil {
		inv.Memo = *req.Memo
	}

	if req.RPreImage != nil {
		rPreImage, err := hex.DecodeString(*req.RPreImage)
		if err != nil {
			return &lnrpc.Invoice{}, errors.New("error decoding preimage")
		}
		inv.RPreimage = rPreImage
	}

	if req.ValueMsat != nil {
		inv.ValueMsat = *req.ValueMsat
	}

	if req.Expiry != nil {
		inv.Expiry = *req.Expiry
	}

	if req.FallBackAddress != nil {
		inv.FallbackAddr = *req.FallBackAddress
	}

	if req.Private != nil {
		inv.Private = *req.Private
	}

	if req.IsAmp != nil {
		inv.IsAmp = *req.IsAmp
	}

	return inv, nil
}
