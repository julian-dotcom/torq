package channels

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"strconv"
	"strings"
)

type failedUpdate struct {
	outpoint struct {
		txid    string
		outIndx uint32
	}
	reason      string
	updateError string
}

type updateChanRequestBody struct {
	ChannelPoint  *string
	FeeRatePpm    *uint32
	BaseFeeMsat   *int64
	MaxHtlcMsat   *uint64
	MinHtlcMsat   *uint64
	TimeLockDelta uint32
}

type lndClientUpdateChannel interface {
	UpdateChannelPolicy(ctx context.Context, in *lnrpc.PolicyUpdateRequest, opts ...grpc.CallOption) (*lnrpc.PolicyUpdateResponse, error)
}

func createPolicyRequest(req updateChanRequestBody) (r lnrpc.PolicyUpdateRequest, err error) {

	var fundingTxid *string
	var outputIndex *uint32

	if req.ChannelPoint != nil {
		splitChanPoint := strings.Split(*req.ChannelPoint, ":")
		if len(splitChanPoint) != 2 {
			return r, errors.New("channel point missing a colon")
		}

		fundingTxid = &splitChanPoint[0]

		oIndxUint, err := strconv.ParseUint(splitChanPoint[1], 10, 1)
		if err != nil {
			return r, errors.Newf("parsing channel point output index: %v", err)
		}
		outputIndexUint32 := uint32(oIndxUint)
		outputIndex = &outputIndexUint32
	}

	timeLock := req.TimeLockDelta

	//Minimum supported value for TimeLockDelta supported is 18
	if timeLock < 18 {
		timeLock = 18
	}

	updChanReq := lnrpc.PolicyUpdateRequest{
		TimeLockDelta: timeLock,
	}
	if (fundingTxid == nil && outputIndex != nil) || (fundingTxid != nil && outputIndex == nil) {
		return r, errors.New("Err updating channel: both fundTxid and outputIndex must be specified or none of them")
	}

	if fundingTxid != nil && outputIndex != nil {
		fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: *fundingTxid}

		channelPoint := lnrpc.ChannelPoint{
			FundingTxid: fundingTxid,
			OutputIndex: *outputIndex,
		}

		updChanReq.Scope = &lnrpc.PolicyUpdateRequest_ChanPoint{ChanPoint: &channelPoint}
	}

	if req.FeeRatePpm != nil {
		updChanReq.FeeRatePpm = *req.FeeRatePpm
	}

	if req.MaxHtlcMsat != nil {
		updChanReq.MaxHtlcMsat = *req.MaxHtlcMsat
	}

	if req.MinHtlcMsat != nil {
		updChanReq.MinHtlcMsat = *req.MinHtlcMsat
		updChanReq.MinHtlcMsatSpecified = true
	}

	if req.BaseFeeMsat != nil {
		updChanReq.BaseFeeMsat = *req.BaseFeeMsat
	}
	return r, nil
}

//UpdateChannel
//Returns status, failed updates array
func UpdateChannel(client lndClientUpdateChannel, req updateChanRequestBody) (r UpdateResponse, err error) {

	policyReq, err := createPolicyRequest(req)
	if err != nil {
		log.Error().Msgf("Err updating channel: %v", err)
		r.Status = "Err updating channel"
		return r, err
	}

	ctx := context.Background()
	resp, err := client.UpdateChannelPolicy(ctx, &policyReq)

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
