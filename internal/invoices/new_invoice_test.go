package invoices

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"testing"
)

type MockAddInvoiceLClnt struct {
}

func (m MockAddInvoiceLClnt) AddInvoice(ctx context.Context, in *lnrpc.Invoice, opts ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	resp := lnrpc.AddInvoiceResponse{
		RHash:          nil,
		PaymentRequest: "test",
		AddIndex:       0,
		PaymentAddr:    nil,
	}
	return &resp, nil
}

func TestNewInvoice(t *testing.T) {
	memo := "test"
	valueMsat := int64(123)
	expiry := int64(123)

	client := MockAddInvoiceLClnt{}
	resp, err := newInvoice(client, memo, valueMsat, expiry, false)
	if err != nil {
		t.Fail()
	}

	if resp != "test" {
		t.Fail()
	}

}
