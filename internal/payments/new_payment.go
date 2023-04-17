package payments

import (
	"context"
	"encoding/hex"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/proto/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

type rrpcClientSendPayment interface {
	SendPaymentV2(ctx context.Context, in *routerrpc.SendPaymentRequest,
		opts ...grpc.CallOption) (routerrpc.Router_SendPaymentV2Client,
		error)
}

// SendNewPayment - send new payment
// A new payment can be made either by providing an invoice or by providing:
// dest - the identity pubkey of the payment recipient
// amt(number of satoshi) or amt_msat(number of millisatoshi)
// amt and amt_msat are mutually exclusive
// payments hash - the hash to use within the payment's HTLC
// timeout seconds is mandatory
func SendNewPayment(
	webSocketResponseChannel chan<- interface{},
	db *sqlx.DB,
	npReq core.NewPaymentRequest,
	requestId string,
) (err error) {

	if npReq.NodeId == 0 {
		return errors.New("Node id is missing")
	}

	connectionDetails, err := settings.GetConnectionDetailsById(db, npReq.NodeId)
	if err != nil {
		return errors.Wrap(err, "Getting node connection details from the db")
	}
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return errors.Wrap(err, "Getting node connection details from the db")
	}
	defer conn.Close()
	client := routerrpc.NewRouterClient(conn)
	return sendPayment(client, npReq, webSocketResponseChannel, requestId)
}

func newSendPaymentRequest(npReq core.NewPaymentRequest) (r *routerrpc.SendPaymentRequest, err error) {
	newPayReq := &routerrpc.SendPaymentRequest{
		TimeoutSeconds: npReq.TimeOutSecs,
	}

	if npReq.Invoice != nil {
		newPayReq.PaymentRequest = *npReq.Invoice
	}

	if npReq.FeeLimitMsat != nil && *npReq.FeeLimitMsat != 0 {
		newPayReq.FeeLimitMsat = *npReq.FeeLimitMsat
	}

	if npReq.AmtMSat != nil {
		newPayReq.AmtMsat = *npReq.AmtMSat
	}

	if npReq.AllowSelfPayment != nil {
		newPayReq.AllowSelfPayment = *npReq.AllowSelfPayment
	}

	// TODO: Add support for Keysend, needs to solve issue related to payment hash generation
	//if npReq.Dest != nil {
	//	fmt.Println("It was a keysend")
	//	destHex, err := hex.DecodeString(*npReq.Dest)
	//	if err != nil {
	//		return r, errors.New("Could not decode destination pubkey (keysend)")
	//	}
	//	newPayReq.Dest = destHex
	// //	newPayReq.PaymentHash = make([]byte, 32)
	//}

	return newPayReq, nil
}

func sendPayment(client rrpcClientSendPayment,
	npReq core.NewPaymentRequest,
	webSocketResponseChannel chan<- interface{},
	requestId string) (err error) {

	// Create and validate payment request details
	newPayReq, err := newSendPaymentRequest(npReq)
	if err != nil {
		return err
	}

	ctx := context.Background()
	req, err := client.SendPaymentV2(ctx, newPayReq)
	if err != nil {
		return errors.Wrap(err, "Sending payment")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		resp, err := req.Recv()
		switch err {
		case nil:
			break
		case io.EOF:
		case status.Error(codes.AlreadyExists, ""):
			return errors.New("ALREADY_PAID")
		case status.Error(codes.NotFound, "lnrpc.Lightning_PayInvoice.UnknownPaymentHash"):
			return errors.New("INVALID_HASH")
		case status.Error(codes.InvalidArgument, "lnrpc.Lightning_PayReq.InvalidPaymentRequest"):
			return errors.New("INVALID_PAYMENT_REQUEST")
		case status.Error(codes.InvalidArgument, "lnrpc.Lightning_SendPaymentRequest.CheckPaymentRequest"):
			return errors.New("CHECKSUM_FAILED")
		case status.Error(codes.InvalidArgument, "amount must be specified when paying a zero amount invoice"):
			return errors.New("AMOUNT_REQUIRED")
		case status.Error(codes.InvalidArgument, "amount must not be specified when paying a non-zero amount invoice"):
			return errors.New("AMOUNT_NOT_ALLOWED")
		default:
			log.Error().Msgf("Unknown payment error %v", err)
			return errors.New("UNKNOWN_ERROR")
		}

		if webSocketResponseChannel != nil {
			// Do a non-blocking write as the pub sub + websockets current dead locks itself
			// TODO: Make it so it can't deadlock itself
			select {
			case webSocketResponseChannel <- processResponse(resp, npReq, requestId):
			default:
			}
		}

		// TODO: If LND fails to update us that the payment succeeded or failed for whatever reason
		// this for loop will run forever (a memory leak).
		// We should probably have some kind of timeout which exits regarless after a certain amount of time
		if resp.GetStatus() == lnrpc.Payment_SUCCEEDED || resp.GetStatus() == lnrpc.Payment_FAILED {
			return nil
		}

	}
}

func processResponse(p *lnrpc.Payment, req core.NewPaymentRequest, requestId string) core.NewPaymentResponse {
	r := core.NewPaymentResponse{
		RequestId:     requestId,
		Request:       req,
		Status:        p.Status.String(),
		Hash:          p.PaymentHash,
		Preimage:      p.PaymentPreimage,
		AmountMsat:    p.ValueMsat,
		CreationDate:  time.Unix(0, p.CreationTimeNs),
		FailureReason: p.FailureReason.String(),
		FeePaidMsat:   p.FeeMsat,
	}
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
			h := core.Hops{
				ChanId:           core.ConvertLNDShortChannelID(hop.ChanId),
				AmtToForwardMsat: hop.AmtToForwardMsat,
				Expiry:           hop.Expiry,
				PubKey:           hop.PubKey,
			}
			if hop.MppRecord != nil {
				h.MppRecord.TotalAmtMsat = hop.MppRecord.TotalAmtMsat
				h.MppRecord.PaymentAddr = hex.EncodeToString(hop.MppRecord.PaymentAddr)
			}
			r.Attempt.Route.Hops = append(r.Attempt.Route.Hops, h)
		}

		r.Attempt.Route.TotalTimeLock = attempt.Route.TotalTimeLock
		r.Attempt.Route.TotalAmtMsat = attempt.Route.TotalAmtMsat
	}
	return r
}
