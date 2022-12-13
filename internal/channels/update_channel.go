package channels

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

// UpdateChannel
// Returns status, failed updates array
func updateChannels(db *sqlx.DB, req commons.UpdateChannelRequest, eventChannel chan interface{}) (r commons.UpdateChannelResponse, err error) {

	policyReq, err := createPolicyRequest(req)
	if err != nil {
		return commons.UpdateChannelResponse{}, errors.Wrap(err, "Create policy request")
	}

	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)

	if err != nil {
		return commons.UpdateChannelResponse{}, errors.Wrap(err, "Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return commons.UpdateChannelResponse{}, errors.Wrap(err, "Getting node connection details from the db")
	}

	defer conn.Close()

	ctx := context.Background()
	client := lnrpc.NewLightningClient(conn)

	resp, err := client.UpdateChannelPolicy(ctx, policyReq)
	if err != nil {
		return commons.UpdateChannelResponse{}, errors.Wrap(err, "Updating channel policy")
	}

	r = processUpdateResponse(req, resp, eventChannel)

	return r, nil
}

func createPolicyRequest(req commons.UpdateChannelRequest) (r *lnrpc.PolicyUpdateRequest, err error) {

	updChanReq := &lnrpc.PolicyUpdateRequest{}

	if req.NodeId == 0 {
		return r, errors.New("Node id is missing")
	}

	//Minimum supported value for TimeLockDelta is 18
	if req.TimeLockDelta == nil || *req.TimeLockDelta < 18 {
		updChanReq.TimeLockDelta = 18
	} else {
		updChanReq.TimeLockDelta = *req.TimeLockDelta
	}

	if req.ChannelId != nil && *req.ChannelId != 0 {
		channelSettings := commons.GetChannelSettingByChannelId(*req.ChannelId)
		updChanReq.Scope, err = processChannelPoint(channelSettings.FundingTransactionHash,
			uint32(channelSettings.FundingOutputIndex))
		if err != nil {
			return nil, err
		}
	} else {
		updChanReq.Scope = &lnrpc.PolicyUpdateRequest_Global{Global: true}
	}

	if req.FeeRateMilliMsat != nil {
		updChanReq.FeeRatePpm = uint32(*req.FeeRateMilliMsat)
	}
	if req.FeeBaseMsat != nil {
		updChanReq.BaseFeeMsat = int64(*req.FeeBaseMsat)
	}

	if req.MinHtlcMsat != nil {
		updChanReq.MinHtlcMsat = *req.MinHtlcMsat
		updChanReq.MinHtlcMsatSpecified = true
	}
	if req.MaxHtlcMsat != nil {
		updChanReq.MaxHtlcMsat = *req.MaxHtlcMsat
	}

	return updChanReq, nil
}

// processChannelPoint
// Split received channel point string into fundingtxid and outputindex
// Build PolicyUpdateRequest_ChanPoint: ChannelPoint_FundingTxidStr, ChannelPoint,
// Return PolicyUpdateRequest_ChanPoint
func processChannelPoint(fundingTxidStr string, outputIndex uint32) (cp *lnrpc.PolicyUpdateRequest_ChanPoint,
	err error) {

	fundingTxid := lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: fundingTxidStr}

	lnrpcCP := &lnrpc.ChannelPoint{FundingTxid: &fundingTxid, OutputIndex: outputIndex}
	cp = &lnrpc.PolicyUpdateRequest_ChanPoint{ChanPoint: lnrpcCP}

	return cp, nil
}

func processUpdateResponse(req commons.UpdateChannelRequest, resp *lnrpc.PolicyUpdateResponse, eventChannel chan interface{}) commons.UpdateChannelResponse {
	var r commons.UpdateChannelResponse
	var failedUpdSlice []commons.FailedRequest
	if len(resp.GetFailedUpdates()) > 0 {
		for _, failUpdate := range resp.GetFailedUpdates() {
			failedUpd := commons.FailedRequest{
				Reason: failUpdate.UpdateError,
				Error:  failUpdate.UpdateError,
			}
			failedUpdSlice = append(failedUpdSlice, failedUpd)
		}
		r.Status = commons.Inactive
		r.FailedUpdates = failedUpdSlice
	} else {
		r.Status = commons.Active
	}
	r.Request = req
	if eventChannel != nil {
		eventChannel <- r
	}
	return r
}
