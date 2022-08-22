package payments

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"io"
)

type rrpcClientSendPayment interface {
	SendPayment(ctx context.Context, in *routerrpc.SendPaymentRequest, opts ...grpc.CallOption) (routerrpc.Router_SendPaymentClient, error)
}

//SendNewPayment - send new payment
//A new payment can be made either by providing an invoice or by providing:
//dest - the identity pubkey of the payment recipient
//amt(number of satoshi) or amt_msat(number of millisatoshi)
//amt and amt_msat are mutually exclusive
//payments hash - the hash to use within the payment's HTLC
//timeout seconds is mandatory
func SendNewPayment(dest []byte,
	amt int64,
	amtMSat int64,
	paymentHash []byte,
	invoice string,
	timeOutSecs int32,
	client rrpcClientSendPayment) (r string, err error) {

	ctx := context.Background()
	errs, ctx := errgroup.WithContext(ctx)

	newPayReq := routerrpc.SendPaymentRequest{
		Dest:           dest,
		Amt:            amt,
		AmtMsat:        amtMSat,
		PaymentHash:    paymentHash,
		PaymentRequest: invoice,
		TimeoutSeconds: timeOutSecs,
	}

	newPayRes, err := client.SendPayment(ctx, &newPayReq)
	if err != nil {
		log.Error().Msgf("Err sending payment: %v", err)
		r = "Err sending payment"
		return r, err
	}
	errs.Go(func() error {
		err = receivePayResponse(newPayRes, ctx)
		if err != nil {
			return err
		}
		return nil
	})
	r = "Payment sending"
	return r, errs.Wait()
}

//Get response for new payment request
func receivePayResponse(req routerrpc.Router_SendPaymentClient, ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Error().Msgf("%v", ctx.Err())
			return ctx.Err()
		default:
		}

		resp, err := req.Recv()
		if err == io.EOF {
			log.Info().Msgf("New payment EOF")
			return nil
		}

		if err != nil {
			log.Error().Msgf("Err receive %v", err.Error())
			return err
		}

		if resp.GetState().String() == "SUCCEEDED" {
			log.Info().Msgf("Payment sent")
			return nil
		}

		//log.Debug().Msgf("Sending payment: %v", resp.GetState().String())
	}
}
