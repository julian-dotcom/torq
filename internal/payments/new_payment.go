package payments

import (
	"context"
	"encoding/hex"
	"io"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
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
	npReq commons.NewPaymentRequest,
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

func newSendPaymentRequest(npReq commons.NewPaymentRequest) (r *routerrpc.SendPaymentRequest, err error) {
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
	npReq commons.NewPaymentRequest,
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
		case err != nil && strings.Contains(err.Error(), "amount must be specified when paying a zero amount invoice"):
			return errors.New("AMOUNT_REQUIRED")
		case err != nil && strings.Contains(err.Error(), "amount must not be specified when paying a non-zero  amount invoice"):
			return errors.New("AMOUNT_NOT_ALLOWED")
		default:
			log.Error().Msgf("Unknown payment error %v", err)
			return errors.New("UNKNOWN_ERROR")
		}

		if webSocketResponseChannel != nil {
			// Write the payment status to the client
			webSocketResponseChannel <- processResponse(resp, npReq, requestId)
		}
	}
}

func processResponse(p *lnrpc.Payment, req commons.NewPaymentRequest, requestId string) commons.NewPaymentResponse {
	r := commons.NewPaymentResponse{
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
			h := commons.Hops{
				ChanId:           commons.ConvertLNDShortChannelID(hop.ChanId),
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
