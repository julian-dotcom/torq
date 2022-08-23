package payments

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"google.golang.org/grpc"
	"io"
)

type rrpcClientSendPayment interface {
	SendPayment(ctx context.Context, in *routerrpc.SendPaymentRequest, opts ...grpc.CallOption) (routerrpc.Router_SendPaymentClient, error)
}

type NewPaymentRequest struct {
	Id          string  `json:"id"`
	Type        string  `json:"type"`
	Invoice     string  `json:"invoice"`
	Dest        *[]byte `json:"dest"`
	Amt         *int64  `json:"amt"`
	AmtMSat     *int64  `json:"amtMsat"`
	PaymentHash *[]byte `json:"payment_Hash"`
	TimeOutSecs int32   `json:"timeoutSecs"`
}

type NewPaymentResponse struct {
	Id     string  `json:"id"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

//SendNewPayment - send new payment
//A new payment can be made either by providing an invoice or by providing:
//dest - the identity pubkey of the payment recipient
//amt(number of satoshi) or amt_msat(number of millisatoshi)
//amt and amt_msat are mutually exclusive
//payments hash - the hash to use within the payment's HTLC
//timeout seconds is mandatory
func SendNewPayment(
	wChan chan interface{},
	db *sqlx.DB,
	c *gin.Context,
	npReq NewPaymentRequest,
) (err error) {

	connectionDetails, err := settings.GetConnectionDetails(db)
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Failed connecting to LND")
	}
	defer conn.Close()
	client := routerrpc.NewRouterClient(conn)

	ctx := context.Background()

	newPayReq := routerrpc.SendPaymentRequest{
		PaymentRequest: npReq.Invoice,
		TimeoutSeconds: npReq.TimeOutSecs,
	}
	req, err := client.SendPaymentV2(ctx, &newPayReq)
	if err != nil {
		return errors.Newf("Err sending payment: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		resp, err := req.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return errors.Newf("Err sending payment: %v", err)
		}

		// Write the payment status to the client
		wChan <- resp
	}
}
