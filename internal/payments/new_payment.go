package payments

import (
	"context"
	"encoding/hex"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"google.golang.org/grpc"
	"io"
	"log"
	"strings"
	"time"
)

type rrpcClientSendPayment interface {
	SendPaymentV2(ctx context.Context, in *routerrpc.SendPaymentRequest,
		opts ...grpc.CallOption) (routerrpc.Router_SendPaymentV2Client,
		error)
}

type NewPaymentRequest struct {
	Invoice          string  `json:"invoice"`
	TimeOutSecs      int32   `json:"timeoutSecs"`
	Dest             *[]byte `json:"dest"`
	AmtMSat          *int64  `json:"amountMsat"`
	FeeLimitMsat     *int64  `json:"feeLimitMsat"`
	AllowSelfPayment *bool   `json:"allowSelfPayment"`
}

type MppRecord struct {
	PaymentAddr  string
	TotalAmtMsat int64
}

type hops struct {
	ChanId           string    `json:"chanId"`
	Expiry           uint32    `json:"expiry"`
	AmtToForwardMsat int64     `json:"amtToForwardMsat"`
	PubKey           string    `json:"pubKey"`
	MppRecord        MppRecord `json:"mppRecord"`
	// TODO: Imolement AMP record here when needed
}

type route struct {
	TotalTimeLock uint32 `json:"totalTimeLock"`
	Hops          []hops `json:"hops"`
	TotalAmtMsat  int64  `json:"totalAmtMsat"`
}

type failureDetails struct {
	Reason             string `json:"reason"`
	FailureSourceIndex uint32 `json:"failureSourceIndex"`
	Height             uint32 `json:"height"`
}

type attempt struct {
	AttemptId     uint64         `json:"attemptId"`
	Status        string         `json:"status"`
	Route         route          `json:"route"`
	AttemptTimeNs time.Time      `json:"attemptTimeNs"`
	ResolveTimeNs time.Time      `json:"resolveTimeNs"`
	Preimage      string         `json:"preimage"`
	Failure       failureDetails `json:"failure"`
}
type NewPaymentResponse struct {
	ReqId          string    `json:"reqId"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	Hash           string    `json:"hash"`
	Preimage       string    `json:"preimage"`
	PaymentRequest string    `json:"paymentRequest"`
	AmountMsat     int64     `json:"amountMsat"`
	FeeLimitMsat   int64     `json:"feeLimitMsat"`
	CreationDate   time.Time `json:"creationDate"`
	Attempt        attempt   `json:"path"`
}

type paymentComplete struct {
	ReqId string `json:"id"`
	Type  string `json:"type"`
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
	reqId string,
) (err error) {

	connectionDetails, err := settings.GetConnectionDetails(db)
	if err != nil {
		return errors.New("Error getting node connection details from the db")
	}
	// TODO: which node are you trying to send the payment from?
	conn, err := lnd_connect.Connect(
		connectionDetails[0].GRPCAddress,
		connectionDetails[0].TLSFileBytes,
		connectionDetails[0].MacaroonFileBytes)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Failed connecting to LND")
	}
	defer conn.Close()
	client := routerrpc.NewRouterClient(conn)
	return sendPayment(client, npReq, wChan, reqId)
}

func newSendPaymentRequest(npReq NewPaymentRequest) (r routerrpc.SendPaymentRequest) {
	newPayReq := routerrpc.SendPaymentRequest{
		PaymentRequest: npReq.Invoice,
		TimeoutSeconds: npReq.TimeOutSecs,
	}

	if npReq.FeeLimitMsat != nil {
		newPayReq.FeeLimitMsat = *npReq.FeeLimitMsat
	}

	if npReq.Dest != nil {
		newPayReq.Dest = *npReq.Dest
	}

	if npReq.AmtMSat != nil {
		newPayReq.AmtMsat = *npReq.AmtMSat
	}

	if npReq.AllowSelfPayment != nil {
		newPayReq.AllowSelfPayment = *npReq.AllowSelfPayment
	}
	return newPayReq
}

func sendPayment(client rrpcClientSendPayment, npReq NewPaymentRequest, wChan chan interface{}, reqId string) (err error) {

	// Create and validate payment request details
	newPayReq := newSendPaymentRequest(npReq)

	ctx := context.Background()
	req, err := client.SendPaymentV2(ctx, &newPayReq)
	if err != nil {
		return errors.New("Error sending payment")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		resp, err := req.Recv()
		switch true {
		case err == nil:
			break
		case err == io.EOF:
			return nil
		case err != nil && strings.Contains(err.Error(), "AlreadyExists"):
			return errors.New("ALREADY_PAID")
		case err != nil && strings.Contains(err.Error(), "UnknownPaymentHash"):
			return errors.New("INVALID_HASH")
		case err != nil && strings.Contains(err.Error(), "InvalidPaymentRequest"):
			return errors.New("INVALID_PAYMENT_REQUEST")
		case err != nil && strings.Contains(err.Error(), "checksum failed"):
			return errors.New("CHECKSUM_FAILED")
		default:
			log.Printf("Unknown payment error %v", err)
			return errors.New("UNKNOWN_ERROR")
		}

		// Write the payment status to the client
		wChan <- processResponse(resp, reqId)
	}
}

func processResponse(p *lnrpc.Payment, reqId string) (r NewPaymentResponse) {
	r.ReqId = reqId
	r.Type = "newPayment"
	r.Status = p.Status.String()
	r.Hash = p.PaymentHash
	r.Preimage = p.PaymentPreimage
	r.AmountMsat = p.ValueMsat
	r.CreationDate = time.Unix(0, p.CreationTimeNs)

	for _, attempt := range p.GetHtlcs() {
		r.Attempt.AttemptId = attempt.AttemptId
		r.Attempt.Status = attempt.Status.String()
		r.Attempt.AttemptTimeNs = time.Unix(0, attempt.AttemptTimeNs)
		r.Attempt.ResolveTimeNs = time.Unix(0, attempt.ResolveTimeNs)
		r.Attempt.Preimage = hex.EncodeToString(attempt.Preimage)

		if attempt.Failure != nil {
			r.Attempt.Failure.Reason = attempt.Failure.Code.String()
			r.Attempt.Failure.FailureSourceIndex = attempt.Failure.FailureSourceIndex
			r.Attempt.Failure.Height = attempt.Failure.Height
		}

		for _, hop := range attempt.Route.Hops {
			r.Attempt.Route.Hops = append(r.Attempt.Route.Hops, hops{
				ChanId:           channels.ConvertLNDShortChannelID(hop.ChanId),
				AmtToForwardMsat: hop.AmtToForwardMsat,
				Expiry:           hop.Expiry,
				PubKey:           hop.PubKey,
				MppRecord: MppRecord{
					PaymentAddr:  hex.EncodeToString(hop.MppRecord.PaymentAddr),
					TotalAmtMsat: hop.MppRecord.TotalAmtMsat,
				},
			})
		}

		r.Attempt.Route.TotalTimeLock = attempt.Route.TotalTimeLock
		r.Attempt.Route.TotalAmtMsat = attempt.Route.TotalAmtMsat
	}
	return r
}
