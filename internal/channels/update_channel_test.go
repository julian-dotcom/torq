package channels

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"testing"
)

type MockUpdateChannelLC struct {
}

func (m MockUpdateChannelLC) UpdateChannelPolicy(ctx context.Context, in *lnrpc.PolicyUpdateRequest, opts ...grpc.CallOption) (*lnrpc.PolicyUpdateResponse, error) {
	testOutPoint := lnrpc.OutPoint{
		TxidBytes:   nil,
		TxidStr:     "test",
		OutputIndex: 0,
	}
	testFailUpd := lnrpc.FailedUpdate{
		Outpoint:    &testOutPoint,
		Reason:      0,
		UpdateError: "test",
	}

	var testFailedUpds []*lnrpc.FailedUpdate
	testFailedUpds = append(testFailedUpds, &testFailUpd)

	resp := lnrpc.PolicyUpdateResponse{FailedUpdates: testFailedUpds}
	return &resp, nil
}

func TestUpdateChannel(t *testing.T) {
	fundingTxid := "test"
	outIndx := uint32(0)
	feeRate := float64(123)
	timeLock := uint32(18)

	client := MockUpdateChannelLC{}
	resp, err := UpdateChannel(client, fundingTxid, outIndx, feeRate, timeLock)

	if err != nil {
		t.Fail()
	}

	for _, failedUpd := range resp.FailedUpdates {
		if failedUpd.reason != "test" {
			t.Fail()
		}
	}

	if resp.Status != "Channel update failed" {
		t.Fail()
	}
}
