package on_chain_tx

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"testing"
)

type MockSendCoinsLClnt struct {
}

func (m MockSendCoinsLClnt) SendCoins(ctx context.Context, in *lnrpc.SendCoinsRequest, opts ...grpc.CallOption) (*lnrpc.SendCoinsResponse, error) {
	resp := lnrpc.SendCoinsResponse{Txid: "test"}
	return &resp, nil
}

func TestSendCoins(t *testing.T) {
	address := "test"
	amount := 123
	targetConf := 0
	var satPerVbyte *uint64

	client := MockSendCoinsLClnt{}
	resp, err := sendCoins(client, address, int64(amount), int32(targetConf), satPerVbyte)

	if err != nil {
		t.Fail()
	}

	if resp != "test" {
		t.Fail()
	}
}
