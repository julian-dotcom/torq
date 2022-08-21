package channels

import (
	"context"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type failedUpdate struct {
	outpoint struct {
		txid    string
		outIndx uint32
	}
	reason      string
	updateError string
}

type lndClientUpdateChannel interface {
	UpdateChannelPolicy(ctx context.Context, in *lnrpc.PolicyUpdateRequest, opts ...grpc.CallOption) (*lnrpc.PolicyUpdateResponse, error)
}

//UpdateChannel
//Update channel fee rate
//Returns status, failed updates array
func UpdateChannel(client lndClientUpdateChannel,
	fundTxid string,
	outputIndex uint32,
	feeRate float64,
	timeLock uint32) (r UpdateResponse, err error) {

	fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: fundTxid}

	channelPoint := lnrpc.ChannelPoint{
		FundingTxid: fundingTxid,
		OutputIndex: outputIndex,
	}

	scope := lnrpc.PolicyUpdateRequest_ChanPoint{ChanPoint: &channelPoint}
	updChanReq := lnrpc.PolicyUpdateRequest{
		Scope:   &scope,
		FeeRate: feeRate,
		//		FeeRatePpm:           0,
		TimeLockDelta: timeLock,
		//		MaxHtlcMsat:          0,
		//		MinHtlcMsat:          0,
		//		MinHtlcMsatSpecified: false,
	}
	ctx := context.Background()
	resp, err := client.UpdateChannelPolicy(ctx, &updChanReq)

	if err != nil {
		log.Error().Msgf("Err updating channel: %v", err)
		r.Status = "Err updating channel"
		return r, err
	}

	var failedUpdSlice []failedUpdate
	if len(resp.GetFailedUpdates()) > 0 {
		//log.Debug().Msgf("There are failed updates")
		for _, failUpdate := range resp.GetFailedUpdates() {
			failedUpd := failedUpdate{}
			failedUpd.reason = failUpdate.UpdateError
			failedUpd.updateError = failUpdate.UpdateError
			failedUpd.outpoint.outIndx = failUpdate.Outpoint.OutputIndex
			failedUpd.outpoint.txid = failUpdate.Outpoint.TxidStr
			failedUpdSlice = append(failedUpdSlice, failedUpd)
		}
		return UpdateResponse{
			Status:        "Channel update failed",
			FailedUpdates: failedUpdSlice,
		}, nil
	}

	updResp := UpdateResponse{
		Status:        "Channel updated",
		FailedUpdates: failedUpdSlice,
	}
	//log.Debug().Msgf("response: %v", updResp)

	return updResp, nil
}
