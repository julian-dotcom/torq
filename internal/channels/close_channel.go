package channels

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/lncapital/torq/pkg/commons"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

type CloseChannelRequest struct {
	NodeId          int     `json:"nodeId"`
	ChannelId       int     `json:"channelId"`
	Force           *bool   `json:"force"`
	TargetConf      *int32  `json:"targetConf"`
	DeliveryAddress *string `json:"deliveryAddress"`
	SatPerVbyte     *uint64 `json:"satPerVbyte"`
}

type CloseChannelResponse struct {
	Request                  CloseChannelRequest   `json:"request"`
	Status                   commons.ChannelStatus `json:"status"`
	ClosingTransactionHash   string                `json:"closingTransactionHash"`
	ClosingOutputIndex       uint32                `json:"closingOutputIndex"`
	ClosePendingChannelPoint commons.ChannelPoint  `json:"closePendingChannelPoint"`
}

func CloseChannel(db *sqlx.DB, req CloseChannelRequest) (response CloseChannelResponse, err error) {
	connectionDetails, err := settings.GetConnectionDetailsById(db, req.NodeId)
	if err != nil {
		return CloseChannelResponse{}, errors.New("Getting node connection details from the db")
	}

	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		return CloseChannelResponse{}, errors.Wrap(err, "Connecting to LND")
	}
	defer conn.Close()
	client := lnrpc.NewLightningClient(conn)

	closeChanReq, err := prepareCloseRequest(req)
	if err != nil {
		return CloseChannelResponse{}, errors.Wrap(err, "Preparing close request")
	}

	return closeChannelResp(db, client, closeChanReq, req)
}

func prepareCloseRequest(ccReq CloseChannelRequest) (r *lnrpc.CloseChannelRequest, err error) {

	if ccReq.NodeId == 0 {
		return &lnrpc.CloseChannelRequest{}, errors.New("Node id is missing")
	}

	if ccReq.SatPerVbyte != nil && ccReq.TargetConf != nil {
		return &lnrpc.CloseChannelRequest{}, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	channelSettings := commons.GetChannelSettingByChannelId(ccReq.ChannelId)

	//Make the close channel request
	closeChanReq := &lnrpc.CloseChannelRequest{
		ChannelPoint: &lnrpc.ChannelPoint{
			FundingTxid: &lnrpc.ChannelPoint_FundingTxidStr{
				FundingTxidStr: channelSettings.FundingTransactionHash,
			},
			OutputIndex: uint32(channelSettings.FundingOutputIndex),
		},
	}

	if ccReq.Force != nil {
		closeChanReq.Force = *ccReq.Force
	}

	if ccReq.TargetConf != nil {
		closeChanReq.TargetConf = *ccReq.TargetConf
	}

	if ccReq.SatPerVbyte != nil {
		closeChanReq.SatPerVbyte = *ccReq.SatPerVbyte
	}

	if ccReq.DeliveryAddress != nil {
		closeChanReq.DeliveryAddress = *ccReq.DeliveryAddress
	}

	return closeChanReq, nil
}

func closeChannelResp(db *sqlx.DB, client lnrpc.LightningClient, closeChanReq *lnrpc.CloseChannelRequest,
	ccReq CloseChannelRequest) (CloseChannelResponse, error) {

	ctx := context.Background()
	closeChanRes, err := client.CloseChannel(ctx, closeChanReq)
	if err != nil {
		return CloseChannelResponse{}, errors.Wrap(err, "Closing channel")
	}

	for {
		select {
		case <-ctx.Done():
			//log.Debug().Msgf("%v", ctx.Err())
			return CloseChannelResponse{}, nil
		default:
		}

		resp, err := closeChanRes.Recv()
		if err == io.EOF {
			//log.Debug().Msgf("Close channel EOF")
			return CloseChannelResponse{}, nil
		}
		if err != nil {
			return CloseChannelResponse{}, errors.Wrap(err, "Close channel request receive")
		}

		r, err := processCloseResponse(resp, ccReq)
		if err != nil {
			return CloseChannelResponse{}, errors.Wrap(err, "Process close response")
		}

		pendingChannel, err := client.PendingChannels(ctx, &lnrpc.PendingChannelsRequest{})
		if err != nil {
			return CloseChannelResponse{}, errors.Wrap(err, "Getting pending channels")
		}

		for _, closing := range pendingChannel.WaitingCloseChannels {
			stringOutputIndex := strconv.FormatUint(uint64(closeChanReq.ChannelPoint.GetOutputIndex()), 10)
			if closeChanReq.ChannelPoint.GetFundingTxidStr()+":"+stringOutputIndex == closing.Channel.ChannelPoint {
				r.ClosingTransactionHash = closing.ClosingTxid
			}

			err = updateChannelToClosingByChannelId(db, r.Request.ChannelId, closing.ClosingTxid)
			if err != nil {
				return CloseChannelResponse{}, errors.Wrap(err, "Updating channel to closing status in the db")
			}
		}
		return r, nil
	}
}

func convertChannelPoint(chanPointStr string) (chanPoint *lnrpc.ChannelPoint, err error) {
	splitChanPoint := strings.Split(chanPointStr, ":")
	if len(splitChanPoint) != 2 {
		return chanPoint, errors.New("Channel point missing a colon")
	}

	fundingTxid := &lnrpc.ChannelPoint_FundingTxidStr{FundingTxidStr: splitChanPoint[0]}

	oIndxUint, err := strconv.ParseUint(splitChanPoint[1], 10, 1)
	if err != nil {
		return chanPoint, errors.New("Parsing channel point output index")
	}

	outputIndex := uint32(oIndxUint)
	chanPoint = &lnrpc.ChannelPoint{
		FundingTxid: fundingTxid,
		OutputIndex: outputIndex,
	}

	return chanPoint, nil
}

func processCloseResponse(resp *lnrpc.CloseStatusUpdate, req CloseChannelRequest) (r CloseChannelResponse, err error) {
	switch resp.GetUpdate().(type) {
	case *lnrpc.CloseStatusUpdate_ClosePending:
		ccPending := resp.GetClosePending()
		r = CloseChannelResponse{
			Request:                  req,
			Status:                   commons.Closing,
			ClosePendingChannelPoint: commons.ChannelPoint{TxId: ccPending.Txid, OutputIndex: ccPending.OutputIndex},
		}
	default:
	}

	return r, nil
}
