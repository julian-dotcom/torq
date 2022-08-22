package on_chain_tx

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"testing"
)

type MockNewAddressLClnt struct {
}

func (m MockNewAddressLClnt) NewAddress(ctx context.Context, in *lnrpc.NewAddressRequest, opts ...grpc.CallOption) (*lnrpc.NewAddressResponse, error) {
	resp := lnrpc.NewAddressResponse{Address: "test"}
	return &resp, nil
}

func TestNewInvoice(t *testing.T) {
	addressType := lnrpc.AddressType(1)
	account := "test"

	client := MockNewAddressLClnt{}
	resp, err := newAddress(client, int32(addressType), account)
	if err != nil {
		t.Fail()
	}

	if resp != "test" {
		t.Fail()
	}

}
