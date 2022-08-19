package invoices

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type lndClientAddInvoice interface {
	AddInvoice(ctx context.Context, in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error)
}

func newInvoice(client lndClientAddInvoice, memo string, valueMsat int64, expiry int64, amp bool) (r string, err error) {
	ctx := context.Background()
	invoice := lnrpc.Invoice{
		Memo:      memo,
		ValueMsat: valueMsat,
		Expiry:    expiry,
		IsAmp:     amp,
	}

	resp, err := client.AddInvoice(ctx, &invoice)
	if err != nil {
		log.Error().Msgf("Err creating new invoice: %v", err)
		return "Err creating new invoice", err
	}

	//log.Debug().Msgf("Invoice : %v", resp.PaymentRequest)
	return resp.PaymentRequest, nil
}
